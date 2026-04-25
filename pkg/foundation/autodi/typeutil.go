package main

import (
	"go/types"
	"sort"
	"strings"
	"unicode"
)

// toShortTypeName converts a full type string to its short form.
// "*github.com/.../iam.IAM" → "*iam.IAM"
func toShortTypeName(typeStr string) string {
	prefix := ""
	s := typeStr
	if strings.HasPrefix(s, "*") {
		prefix = "*"
		s = s[1:]
	}

	dotIdx := strings.LastIndex(s, ".")
	if dotIdx < 0 {
		return typeStr
	}

	pkgPath := s[:dotIdx]
	typeName := s[dotIdx+1:]

	parts := strings.Split(pkgPath, "/")
	pkgName := parts[len(parts)-1]

	// Handle versioned paths (v2, v9)
	if len(pkgName) >= 2 && pkgName[0] == 'v' && pkgName[1] >= '0' && pkgName[1] <= '9' && len(parts) >= 2 {
		candidate := parts[len(parts)-2]
		if idx := strings.LastIndex(candidate, "-"); idx >= 0 {
			pkgName = candidate[idx+1:]
		} else {
			pkgName = candidate
		}
	}

	return prefix + pkgName + "." + typeName
}

// sanitizeName replaces dots and slashes with underscores, removes asterisks.
func sanitizeName(s string) string {
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "*", "")
	return s
}

// localVarName converts a FieldName to a local variable name.
func localVarName(fieldName string) string {
	if fieldName == "" {
		return ""
	}
	runes := []rune(fieldName)

	// Count leading uppercase characters
	upperCount := 0
	for upperCount < len(runes) && unicode.IsUpper(runes[upperCount]) {
		upperCount++
	}

	if upperCount == 0 {
		return fieldName
	}

	if upperCount == len(runes) {
		// All uppercase: JWT → jwt, MQ → mq, IAM → iam
		return strings.ToLower(fieldName)
	}

	if upperCount == 1 {
		// Single uppercase: EntClient → entClient
		return strings.ToLower(string(runes[0])) + string(runes[1:])
	}

	// Multiple uppercase prefix: IamAuthN → iamAuthN, HTTPServer → httpServer
	return strings.ToLower(string(runes[:upperCount-1])) + string(runes[upperCount-1:])
}

// zeroValueForType returns the zero value literal for a Go type.
func zeroValueForType(t types.Type) string {
	switch u := t.Underlying().(type) {
	case *types.Pointer, *types.Interface, *types.Map, *types.Slice, *types.Chan:
		return "nil"
	case *types.Basic:
		switch u.Kind() {
		case types.String:
			return `""`
		case types.Bool:
			return "false"
		default:
			return "0"
		}
	default:
		return "nil"
	}
}

// cmdExportName converts a command name to an exported function name.
// "admin_api" → "AdminAPI", "admin" → "Admin"
func cmdExportName(name string) string {
	parts := strings.Split(name, "_")
	var result string
	for _, p := range parts {
		if len(p) <= 3 {
			result += strings.ToUpper(p)
		} else {
			result += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return result
}

// deriveSliceVarName generates a local variable name from an interface type string.
// "github.com/.../seed.Seeder" → "seeders"
// "github.com/.../provider.CNIDriver" → "cniDrivers"
func deriveSliceVarName(elemTypeStr string) string {
	// Extract the short type name from the full path
	short := toShortTypeName(elemTypeStr)
	// Remove package prefix: "seed.Seeder" → "Seeder", "provider.CNIDriver" → "CNIDriver"
	if dotIdx := strings.LastIndex(short, "."); dotIdx >= 0 {
		short = short[dotIdx+1:]
	}
	// Pluralize and lowercase
	return localVarName(pluralizeName(short))
}

// pluralizeName appends "s" or "es" to make a plural form.
// "Seeder" → "Seeders", "Driver" → "Drivers"
func pluralizeName(name string) string {
	if strings.HasSuffix(name, "s") || strings.HasSuffix(name, "x") || strings.HasSuffix(name, "z") {
		return name + "es"
	}
	return name + "s"
}

// sortedGroupNames returns group names in deterministic order.
func sortedGroupNames(groups map[string]GroupConfig) []string {
	var names []string
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// typePkgPathFromTypeStr extracts the package path from a full type string.
func typePkgPathFromTypeStr(typeStr string) string {
	s := strings.TrimPrefix(typeStr, "*")
	if idx := strings.LastIndex(s, "."); idx >= 0 {
		return s[:idx]
	}
	return ""
}
