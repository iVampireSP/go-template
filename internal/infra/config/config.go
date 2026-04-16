package config

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	data     map[string]any
	flatData map[string]any // 扁平化缓存，避免运行时路径解析
	mu       sync.RWMutex
	loadedAt time.Time
)

// MustInit 从指定目录加载配置文件，失败则 panic
// 加载顺序：
// 1. 加载 .env 文件到环境变量
// 2. 从文件系统加载所有 YAML 文件
// 3. YAML 中的 ${VAR} 和 ${VAR:-default} 会被环境变量替换
func MustInit(dir string) {
	if err := Init(dir); err != nil {
		panic("config: " + err.Error())
	}
}

// Init 从指定目录加载配置文件
func Init(dir string) error {
	return InitWithFS(os.DirFS(dir), ".")
}

// MustInitWithFS 使用指定的文件系统和目录初始化配置，失败则 panic
// 加载顺序：
// 1. 加载 .env 文件到环境变量
// 2. 从嵌入的 fs.FS 加载所有 YAML 文件
// 3. YAML 中的 ${VAR} 和 ${VAR:-default} 会被环境变量替换
func MustInitWithFS(configFS fs.FS, dir string) {
	if err := InitWithFS(configFS, dir); err != nil {
		panic("config: " + err.Error())
	}
}

// InitWithFS 使用指定的文件系统和目录初始化配置
func InitWithFS(configFS fs.FS, dir string) error {
	mu.Lock()
	defer mu.Unlock()

	// 加载 .env 文件
	LoadDotEnv()

	// 从嵌入的文件系统加载所有 YAML 文件
	data = make(map[string]any)
	entries, err := fs.ReadDir(configFS, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		content, err := fs.ReadFile(configFS, path.Join(dir, name))
		if err != nil {
			return err
		}

		// 替换环境变量
		replaced := substituteEnv(string(content))

		var fileData any
		if err := yaml.Unmarshal([]byte(replaced), &fileData); err != nil {
			return fmt.Errorf("failed to parse %s: %w", name, err)
		}

		// 文件名作为顶级 key
		key := strings.TrimSuffix(name, ".yaml")
		key = strings.TrimSuffix(key, ".yml")
		data[key] = fileData
	}

	// 预计算扁平化缓存，避免运行时路径解析和遍历
	flatData = flattenMap(data, "")

	loadedAt = time.Now()
	return nil
}

// flattenMap 递归扁平化嵌套 map，将 "a.b.c" 格式的键映射到值
func flattenMap(m map[string]any, prefix string) map[string]any {
	result := make(map[string]any, len(m)*2) // 预分配容量
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			for fk, fv := range flattenMap(val, key) {
				result[fk] = fv
			}
		default:
			result[key] = v
		}
	}
	return result
}

// Get 获取配置值，支持点号分隔的路径
// 优先使用扁平化缓存（O(1) 查找），未命中时回退到原始遍历
// 示例: Get("app.name"), Get("database.host")
func Get(path string, defaultValue ...any) any {
	mu.RLock()
	defer mu.RUnlock()

	// 优先从扁平缓存查找（叶子节点，O(1)）
	if v, ok := flatData[path]; ok {
		return v
	}

	// 回退到原始遍历（支持获取中间节点如 Map("app")）
	parts := strings.Split(path, ".")
	var current any = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = v[part]
			if !ok {
				return getDefault(defaultValue)
			}
		default:
			return getDefault(defaultValue)
		}
	}

	return current
}

// String 获取字符串配置
func String(path string, defaultValue ...string) string {
	def := ""
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	if v, ok := Get(path).(string); ok {
		return v
	}
	return def
}

// Int 获取整数配置
func Int(path string, defaultValue ...int) int {
	def := 0
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	switch v := Get(path).(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return def
}

// Int64 获取 int64 配置
func Int64(path string, defaultValue ...int64) int64 {
	var def int64
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	switch v := Get(path).(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return def
}

// Bool 获取布尔配置
func Bool(path string, defaultValue ...bool) bool {
	def := false
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	if v, ok := Get(path).(bool); ok {
		return v
	}
	return def
}

// Duration 获取时间段配置
func Duration(path string, defaultValue ...time.Duration) time.Duration {
	var def time.Duration
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	if v, ok := Get(path).(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// Strings 获取字符串切片
func Strings(path string, defaultValue ...[]string) []string {
	var def []string
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	if v, ok := Get(path).([]any); ok {
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return def
}

// Map 获取 map 配置
func Map(path string) map[string]any {
	if v, ok := Get(path).(map[string]any); ok {
		return v
	}
	return nil
}

// Has 检查配置是否存在
func Has(path string) bool {
	return Get(path) != nil
}

// All 返回所有配置
func All() map[string]any {
	mu.RLock()
	defer mu.RUnlock()
	return data
}

// LoadedAt 返回配置加载时间
func LoadedAt() time.Time {
	return loadedAt
}

// Unmarshal 将配置路径下的数据解析到目标结构体
func Unmarshal(path string, target any) error {
	d := Map(path)
	if d == nil {
		return nil
	}
	bytes, err := yaml.Marshal(d)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, target)
}

func getDefault(defaultValue []any) any {
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// GetFloat gets a float64 config value (config package doesn't have Float method)
func GetFloat(path string, defaultValue float64) float64 {
	val := Get(path)
	if val == nil {
		return defaultValue
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return defaultValue
}

// GetIntSlice gets an int slice from config
func GetIntSlice(path string, defaultValue []int) []int {
	val := Get(path)
	if val == nil {
		return defaultValue
	}
	if arr, ok := val.([]any); ok {
		result := make([]int, 0, len(arr))
		for _, item := range arr {
			switch v := item.(type) {
			case int:
				result = append(result, v)
			case int64:
				result = append(result, int(v))
			case float64:
				result = append(result, int(v))
			}
		}
		return result
	}
	return defaultValue
}

// Env returns the current application environment
func Env() string {
	return String("app.env", "development")
}

// IsDevelopment returns true if the application is running in development environment
func IsDevelopment() bool {
	env := Env()
	return env == "development" || env == "dev" || env == "local"
}

// IsProduction returns true if the application is running in production environment
func IsProduction() bool {
	env := Env()
	return env == "production" || env == "prod"
}
