package main

import (
	"go/token"
	"go/types"
	"strings"
)

// Provider represents a discovered New* constructor function.
type Provider struct {
	FuncName    string         // e.g., "NewIAM"
	PkgPath     string         // e.g., "github.com/LeaflowNET/cloud/internal/services/iam"
	PkgName     string         // e.g., "iam"
	Params      []TypeRef      // input parameters (dependencies)
	Returns     []TypeRef      // return values (provided types)
	HasError    bool           // last return is error
	IsInvoke    bool           // call-only, no stored result
	Annotations []Annotation   // parsed //autodi: directives
	Position    token.Position // source location for errors

	// Resolved during graph building
	Groups []string // group memberships
}

// TypeRef describes a single type in a provider's signature.
type TypeRef struct {
	Type     types.Type
	TypeStr  string // qualified string like "*ent.Client", "iam.AuthN"
	PkgPath  string // package path for this type
	IsIface  bool   // whether this is an interface type
	Optional bool   // from //autodi:optional
}

// RelPath returns the relative package path within the module.
func (p *Provider) RelPath(module string) string {
	return strings.TrimPrefix(p.PkgPath, module+"/")
}

// FieldName generates a Container field name for this provider's return type.
// Uses the package short name + type name to produce unique, readable names.
func FieldName(typeStr string) string {
	s := typeStr
	s = strings.TrimPrefix(s, "*")

	// Split into package path and type name at the last dot
	dotIdx := strings.LastIndex(s, ".")
	if dotIdx < 0 {
		return exportName(s)
	}

	pkgPath := s[:dotIdx]
	typeName := s[dotIdx+1:]

	// Get short package name
	pkg := pkgPath
	if idx := strings.LastIndex(pkg, "/"); idx >= 0 {
		pkg = pkg[idx+1:]
	}
	// Handle versioned paths (v2, v9)
	if len(pkg) >= 2 && pkg[0] == 'v' && pkg[1] >= '0' && pkg[1] <= '9' {
		parts := strings.Split(pkgPath, "/")
		if len(parts) >= 2 {
			pkg = parts[len(parts)-2]
			if idx := strings.LastIndex(pkg, "-"); idx >= 0 {
				pkg = pkg[idx+1:]
			}
		}
	}

	// If type name already incorporates the package name, skip prefix
	// e.g., pkg="iam", name="IAM" → just "IAM"
	// e.g., pkg="redisx", name="Locker" → "RedisxLocker"
	// e.g., pkg="ent", name="Client" → "EntClient"
	if strings.EqualFold(pkg, typeName) {
		return exportName(typeName)
	}
	if len(typeName) > len(pkg) && strings.EqualFold(typeName[:len(pkg)], pkg) {
		return exportName(typeName)
	}
	return exportName(pkg) + exportName(typeName)
}

// exportName ensures first letter is uppercase.
func exportName(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// ImportAlias returns the import alias needed for a package, or empty if default is fine.
func ImportAlias(pkgPath, pkgName string, used map[string]string) string {
	// Check if this package name is already used by a different path
	if existingPath, ok := used[pkgName]; ok && existingPath != pkgPath {
		// Need alias — use parent dir + pkg name
		parts := strings.Split(pkgPath, "/")
		if len(parts) >= 2 {
			parent := parts[len(parts)-2]
			alias := parent + pkgName
			// Check this alias isn't also taken
			if _, exists := used[alias]; !exists {
				return alias
			}
			// Fallback to more segments
			if len(parts) >= 3 {
				return parts[len(parts)-3] + parent + pkgName
			}
		}
		return pkgName + "2"
	}
	return ""
}

// isErrorType checks if a type is the built-in error interface.
func isErrorType(t types.Type) bool {
	return types.Identical(t, types.Universe.Lookup("error").Type())
}

// isInterface checks if a type is an interface (not including error).
func isInterface(t types.Type) bool {
	// Dereference pointer
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	_, ok := t.Underlying().(*types.Interface)
	return ok && !isErrorType(t)
}

// typePkgPath extracts the package path from a type.
func typePkgPath(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		if t.Obj().Pkg() != nil {
			return t.Obj().Pkg().Path()
		}
	case *types.Pointer:
		return typePkgPath(t.Elem())
	}
	return ""
}
