package validator

import (
	"context"
	"reflect"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"
)

// FieldValidator 字段级自定义验证器。
// 每个实现注册自己的 struct tag 名，httpserver.Register 自动扫描并调用。
type FieldValidator interface {
	// Tag 返回验证器注册的 struct tag 名（如 "image"）。
	Tag() string
	// ValidateField 验证单个字段值。param 是 tag 值（如 `image:"strict"` 中的 "strict"）。
	ValidateField(ctx context.Context, value any, param string) []error
}

// fieldRule 描述一个字段上匹配到的验证规则。
type fieldRule struct {
	path      string // JSON 路径，如 "body.containers"
	fieldIdx  []int  // reflect 字段索引链
	tag       string // tag 名
	param     string // tag 值
	validator FieldValidator
}

var (
	registry   = make(map[string]FieldValidator)
	registryMu sync.RWMutex
	tagNames   []string // 已注册 tag 名列表

	// 缓存：reflect.Type → []fieldRule（注册时解析一次）
	rulesCache sync.Map
)

// MustInit 注册所有字段验证器。启动时调用一次。
func MustInit(validators ...FieldValidator) {
	registryMu.Lock()
	defer registryMu.Unlock()
	for _, v := range validators {
		name := v.Tag()
		registry[name] = v
		tagNames = append(tagNames, name)
	}
}

// HasRegisteredTags 检查类型（含 Body 嵌套）是否有任何已注册 tag。结果缓存。
func HasRegisteredTags(t reflect.Type) bool {
	rules := resolveRules(t)
	return len(rules) > 0
}

// ValidateFields 反射遍历 struct，对每个匹配已注册 tag 的字段调用对应验证器。
func ValidateFields(ctx context.Context, v any) []error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rules := resolveRules(rv.Type())
	if len(rules) == 0 {
		return nil
	}

	var errs []error
	for _, rule := range rules {
		fieldVal := rv.FieldByIndex(rule.fieldIdx)
		fieldErrs := rule.validator.ValidateField(ctx, fieldVal.Interface(), rule.param)
		for _, e := range fieldErrs {
			// 如果是 huma.ErrorDetail，补充 Location
			if detail, ok := e.(*huma.ErrorDetail); ok {
				if detail.Location == "" {
					detail.Location = rule.path
				} else if !strings.HasPrefix(detail.Location, rule.path) {
					detail.Location = rule.path + "." + detail.Location
				}
			}
			errs = append(errs, e)
		}
	}
	return errs
}

// resolveRules 解析类型的验证规则，结果缓存。
func resolveRules(t reflect.Type) []fieldRule {
	if cached, ok := rulesCache.Load(t); ok {
		return cached.([]fieldRule)
	}

	registryMu.RLock()
	names := make([]string, len(tagNames))
	copy(names, tagNames)
	validators := make(map[string]FieldValidator, len(registry))
	for k, v := range registry {
		validators[k] = v
	}
	registryMu.RUnlock()

	var rules []fieldRule
	scanFields(t, nil, "", names, validators, &rules)

	rulesCache.Store(t, rules)
	return rules
}

// scanFields 递归扫描 struct 字段，查找匹配已注册 tag 的字段。
func scanFields(t reflect.Type, parentIdx []int, pathPrefix string, tags []string, validators map[string]FieldValidator, rules *[]fieldRule) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		idx := append(append([]int{}, parentIdx...), i)

		// 构建 JSON 路径
		jsonName := jsonFieldName(f)
		var path string
		if pathPrefix != "" {
			path = pathPrefix + "." + jsonName
		} else {
			path = jsonName
		}

		// 检查字段是否有已注册 tag
		for _, tag := range tags {
			if param, ok := f.Tag.Lookup(tag); ok {
				if v, exists := validators[tag]; exists {
					*rules = append(*rules, fieldRule{
						path:      path,
						fieldIdx:  idx,
						tag:       tag,
						param:     param,
						validator: v,
					})
				}
			}
		}

		// 递归嵌套 struct（包括 Body 字段）
		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			// Body 字段特殊处理：路径前缀用 "body"
			if f.Name == "Body" {
				scanFields(ft, idx, "body", tags, validators, rules)
			} else if f.Anonymous {
				scanFields(ft, idx, pathPrefix, tags, validators, rules)
			} else {
				scanFields(ft, idx, path, tags, validators, rules)
			}
		}
	}
}

// jsonFieldName 从 struct field 获取 JSON 字段名。
func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" || tag == "-" {
		return strings.ToLower(f.Name)
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return strings.ToLower(f.Name)
	}
	return name
}
