package i18n

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type ctxKey struct{}

var (
	global *i18n
	mu     sync.RWMutex
)

type i18n struct {
	// locale → flatKey → value
	locales        map[string]map[string]string
	defaultLocale  string
	fallbackLocale string
}

// MustInitWithFS 从指定 fs.FS 和目录加载所有语言文件，失败则 panic
func MustInitWithFS(cfg Config, langFS fs.FS, dir string) {
	if err := InitWithFS(cfg, langFS, dir); err != nil {
		panic("i18n: " + err.Error())
	}
}

// InitWithFS 从指定 fs.FS 和目录加载所有语言文件
// 目录结构：dir/{locale}/*.yaml
// 文件名作为 key 前缀：email.yaml 中的 key "Foo" → "email.Foo"
func InitWithFS(cfg Config, langFS fs.FS, dir string) error {
	mu.Lock()
	defer mu.Unlock()

	inst := &i18n{
		locales:        make(map[string]map[string]string),
		defaultLocale:  cfg.DefaultLocale,
		fallbackLocale: cfg.FallbackLocale,
	}

	// 枚举 locale 目录
	entries, err := fs.ReadDir(langFS, dir)
	if err != nil {
		return fmt.Errorf("failed to read lang dir %q: %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		locale := entry.Name()
		localeDir := path.Join(dir, locale)

		flat, err := loadLocale(langFS, localeDir)
		if err != nil {
			return fmt.Errorf("failed to load locale %q: %w", locale, err)
		}
		inst.locales[locale] = flat
	}

	global = inst
	return nil
}

// loadLocale 加载单个 locale 目录下所有 YAML 文件，返回扁平 key → value 映射
func loadLocale(langFS fs.FS, localeDir string) (map[string]string, error) {
	flat := make(map[string]string)

	entries, err := fs.ReadDir(langFS, localeDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		content, err := fs.ReadFile(langFS, path.Join(localeDir, name))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", name, err)
		}

		var fileData any
		if err := yaml.Unmarshal(content, &fileData); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", name, err)
		}

		// 文件名（去掉扩展名）作为顶级 key 前缀
		prefix := strings.TrimSuffix(name, ".yaml")
		prefix = strings.TrimSuffix(prefix, ".yml")

		flattenInto(flat, fileData, prefix)
	}

	return flat, nil
}

// flattenInto 递归扁平化 YAML 数据到 flat map，叶节点值转为 string
func flattenInto(flat map[string]string, data any, prefix string) {
	switch v := data.(type) {
	case map[string]any:
		for k, child := range v {
			key := prefix + "." + k
			flattenInto(flat, child, key)
		}
	case string:
		flat[prefix] = v
	case nil:
		// 忽略 null 值
	default:
		flat[prefix] = fmt.Sprint(v)
	}
}

// T 从 ctx 中读取 locale，翻译 key，找不到时使用 defaultValue
func T(ctx context.Context, key string, defaultValue ...string) string {
	return TWithLocale(Locale(ctx), key, defaultValue...)
}

// TWithLocale 使用指定 locale 翻译 key，找不到时按 fallback 链查找，最终使用 defaultValue
func TWithLocale(locale, key string, defaultValue ...string) string {
	mu.RLock()
	inst := global
	mu.RUnlock()

	def := key
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	if inst == nil {
		return def
	}

	// 1. 查找指定 locale
	if v, ok := inst.locales[locale][key]; ok {
		return v
	}

	// 2. 查找 fallback locale
	if locale != inst.fallbackLocale {
		if v, ok := inst.locales[inst.fallbackLocale][key]; ok {
			return v
		}
	}

	// 3. 使用 defaultValue
	return def
}

// WithLocale 将 locale 写入 ctx
func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, ctxKey{}, locale)
}

// Locale 从 ctx 读取 locale，未设置时返回 defaultLocale
func Locale(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok && v != "" {
		return v
	}
	return DefaultLocale()
}

// DefaultLocale 返回配置的默认 locale
func DefaultLocale() string {
	mu.RLock()
	inst := global
	mu.RUnlock()
	if inst == nil {
		return "zh_CN"
	}
	return inst.defaultLocale
}

// FallbackLocale 返回配置的 fallback locale
func FallbackLocale() string {
	mu.RLock()
	inst := global
	mu.RUnlock()
	if inst == nil {
		return "en"
	}
	return inst.fallbackLocale
}

// Locales 返回所有已加载的 locale 列表
func Locales() []string {
	mu.RLock()
	inst := global
	mu.RUnlock()
	if inst == nil {
		return nil
	}
	locales := make([]string, 0, len(inst.locales))
	for k := range inst.locales {
		locales = append(locales, k)
	}
	return locales
}

// ChiMiddleware 返回一个 Chi HTTP 中间件，解析 Accept-Language header 并写入 ctx
func ChiMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			locale := resolveLocale(r.Header.Get("Accept-Language"))
			ctx := WithLocale(r.Context(), locale)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// resolveLocale 解析 Accept-Language header，匹配已加载的 locale
// 规范化：zh-CN → zh_CN，en-US → en（取语言部分做最长前缀匹配）
func resolveLocale(header string) string {
	mu.RLock()
	inst := global
	mu.RUnlock()

	if inst == nil || header == "" {
		return DefaultLocale()
	}

	// 解析 Accept-Language：取第一个 tag（quality 最高的通常排最前）
	tags := strings.Split(header, ",")
	for _, tag := range tags {
		// 去掉 ;q=xxx 部分
		tag = strings.TrimSpace(strings.SplitN(tag, ";", 2)[0])
		if tag == "" {
			continue
		}

		// 规范化 zh-CN → zh_CN
		normalized := strings.ReplaceAll(tag, "-", "_")

		// 精确匹配
		if _, ok := inst.locales[normalized]; ok {
			return normalized
		}

		// 语言前缀匹配（en_US → en, zh_TW → zh_CN 等）
		lang := strings.SplitN(normalized, "_", 2)[0]
		for k := range inst.locales {
			if strings.HasPrefix(k, lang) {
				return k
			}
		}
	}

	return inst.defaultLocale
}
