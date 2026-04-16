package httpserver

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/danielgtaylor/huma/v2"
)

func schemaNameForType(t reflect.Type, hint string) string {
	defaultName := sanitizeSchemaName(huma.DefaultSchemaNamer(t, hint))

	baseType := t
	for baseType.Kind() == reflect.Pointer {
		baseType = baseType.Elem()
	}
	if baseType.Name() == "" || baseType.PkgPath() == "" {
		return defaultName
	}

	typeName := sanitizeSchemaName(baseType.Name())
	if typeName == "" {
		return defaultName
	}

	return sanitizeSchemaName(baseType.PkgPath()) + "__" + typeName
}

func sanitizeSchemaName(name string) string {
	var b strings.Builder
	b.Grow(len(name))

	lastUnderscore := false
	for _, ch := range name {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			b.WriteRune(ch)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}

	result := strings.Trim(b.String(), "_")
	if result == "" {
		return "schema"
	}
	return result
}
