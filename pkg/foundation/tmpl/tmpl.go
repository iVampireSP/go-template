package tmpl

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"strings"

	"github.com/iVampireSP/go-template/pkg/foundation/i18n"
)

// defaultFS 通过 MustInitWithFS 注册的模板文件系统。
var (
	defaultFS    embed.FS
	defaultFSSet bool
	// locale → slash-path → compiled template（如 "email/welcome"）
	defaultTemplates map[string]map[string]*template.Template
)

// MustInitWithFS 注册模板 FS，必须在任何 Render() 调用前执行。
func MustInitWithFS(templateFS embed.FS) {
	defaultFS = templateFS
	defaultFSSet = true

	sub, err := fs.Sub(defaultFS, "templates")
	if err != nil {
		panic(fmt.Sprintf("tmpl: failed to sub template fs: %v", err))
	}

	templates, err := loadTemplates(sub)
	if err != nil {
		panic(fmt.Sprintf("tmpl: failed to load templates: %v", err))
	}
	defaultTemplates = templates
}

// Render 渲染指定模板，返回 subject 和 HTML body。
// ctx 用于读取 locale（i18n.Locale(ctx)）。
func Render(ctx context.Context, templatePath string, data any) (subject, html string, err error) {
	if !defaultFSSet {
		return "", "", fmt.Errorf("tmpl: MustInitWithFS has not been called")
	}

	locale := resolveLocale(i18n.Locale(ctx))
	localeTemplates, ok := defaultTemplates[locale]
	if !ok {
		return "", "", fmt.Errorf("tmpl: locale %q not found", locale)
	}

	tmpl, ok := localeTemplates[templatePath]
	if !ok {
		return "", "", fmt.Errorf("tmpl: template %q does not exist", templatePath)
	}

	var subjectBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&subjectBuf, "subject", data); err != nil {
		return "", "", fmt.Errorf("tmpl: failed to render subject for %q: %w", templatePath, err)
	}

	var bodyBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&bodyBuf, "base", data); err != nil {
		return "", "", fmt.Errorf("tmpl: failed to render template %q: %w", templatePath, err)
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}

// loadTemplates 为每个已加载的 locale 编译一套模板。
func loadTemplates(fsys fs.FS) (map[string]map[string]*template.Template, error) {
	locales := i18n.Locales()
	if len(locales) == 0 {
		locales = []string{i18n.DefaultLocale()}
	}

	templates := make(map[string]map[string]*template.Template, len(locales))
	for _, locale := range locales {
		compiled, err := compileLocale(fsys, locale)
		if err != nil {
			return nil, fmt.Errorf("locale %s: %w", locale, err)
		}
		templates[locale] = compiled
	}
	return templates, nil
}

// compileLocale 扫描模板文件系统中的所有子目录，编译每个目录下的模板。
func compileLocale(fsys fs.FS, locale string) (map[string]*template.Template, error) {
	loc := locale
	funcs := template.FuncMap{
		"t": func(key string, defaultValue ...string) string {
			return i18n.TWithLocale(loc, key, defaultValue...)
		},
	}

	compiled := make(map[string]*template.Template)

	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read template root: %w", err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		dirName := dir.Name()

		baseContent, err := fs.ReadFile(fsys, path.Join(dirName, "base.html"))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s/base.html: %w", dirName, err)
		}

		entries, err := fs.ReadDir(fsys, dirName)
		if err != nil {
			return nil, fmt.Errorf("failed to read dir %s: %w", dirName, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || entry.Name() == "base.html" || !strings.HasSuffix(entry.Name(), ".html") {
				continue
			}

			content, err := fs.ReadFile(fsys, path.Join(dirName, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read template %s/%s: %w", dirName, entry.Name(), err)
			}

			t, err := template.New("base").Funcs(funcs).Parse(string(baseContent))
			if err != nil {
				return nil, fmt.Errorf("failed to parse base for %s: %w", dirName, err)
			}
			if _, err = t.Parse(string(content)); err != nil {
				return nil, fmt.Errorf("failed to parse %s/%s: %w", dirName, entry.Name(), err)
			}

			key := dirName + "/" + strings.TrimSuffix(entry.Name(), ".html")
			compiled[key] = t
		}
	}
	return compiled, nil
}

// resolveLocale 解析最佳匹配 locale。
func resolveLocale(requested string) string {
	if _, ok := defaultTemplates[requested]; ok {
		return requested
	}
	lang := strings.SplitN(requested, "_", 2)[0]
	for loc := range defaultTemplates {
		if strings.HasPrefix(loc, lang+"_") {
			return loc
		}
	}
	fb := i18n.FallbackLocale()
	if _, ok := defaultTemplates[fb]; ok {
		return fb
	}
	for loc := range defaultTemplates {
		return loc
	}
	return requested
}
