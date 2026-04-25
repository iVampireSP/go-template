package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// ImportManager tracks imports and handles alias conflicts.
type ImportManager struct {
	imports map[string]string // pkgPath → alias (or empty for default)
	used    map[string]string // pkgName → pkgPath (first use wins)
}

func NewImportManager() *ImportManager {
	return &ImportManager{
		imports: make(map[string]string),
		used:    make(map[string]string),
	}
}

// Add registers an import and returns the qualifier to use in code.
func (im *ImportManager) Add(pkgPath, pkgName string) string {
	if pkgPath == "" {
		return ""
	}

	// Already registered
	if alias, ok := im.imports[pkgPath]; ok {
		if alias != "" {
			return alias
		}
		return pkgName
	}

	// Check for name conflict
	if existingPath, conflict := im.used[pkgName]; conflict && existingPath != pkgPath {
		// Need an alias
		alias := im.makeAlias(pkgPath, pkgName)
		im.imports[pkgPath] = alias
		im.used[alias] = pkgPath
		return alias
	}

	im.imports[pkgPath] = "" // no alias needed
	im.used[pkgName] = pkgPath
	return pkgName
}

// AddWithAlias registers an import with an explicit alias.
func (im *ImportManager) AddWithAlias(pkgPath, alias string) string {
	if pkgPath == "" {
		return ""
	}
	if existing, ok := im.imports[pkgPath]; ok {
		if existing != "" {
			return existing
		}
		// Already registered without alias — keep it
		return pkgShortName(pkgPath)
	}
	im.imports[pkgPath] = alias
	im.used[alias] = pkgPath
	return alias
}

func (im *ImportManager) makeAlias(pkgPath, pkgName string) string {
	parts := strings.Split(pkgPath, "/")
	if len(parts) >= 2 {
		parent := parts[len(parts)-2]
		alias := parent + pkgName
		if _, exists := im.used[alias]; !exists {
			return alias
		}
	}
	// Fallback
	for i := 2; ; i++ {
		alias := fmt.Sprintf("%s%d", pkgName, i)
		if _, exists := im.used[alias]; !exists {
			return alias
		}
	}
}

// FormatBlock returns the import block as Go source.
func (im *ImportManager) FormatBlock() string {
	if len(im.imports) == 0 {
		return ""
	}

	var paths []string
	for p := range im.imports {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var buf bytes.Buffer
	buf.WriteString("import (\n")
	for _, p := range paths {
		alias := im.imports[p]
		if alias != "" {
			fmt.Fprintf(&buf, "\t%s %q\n", alias, p)
		} else {
			fmt.Fprintf(&buf, "\t%q\n", p)
		}
	}
	buf.WriteString(")\n")
	return buf.String()
}

// IsQualifier checks if a name is used as an import qualifier.
func (im *ImportManager) IsQualifier(name string) bool {
	_, ok := im.used[name]
	return ok
}

// Reset clears all imports for a new file.
func (im *ImportManager) Reset() {
	im.imports = make(map[string]string)
	im.used = make(map[string]string)
}

// pkgShortName extracts the short package name from a full path.
func pkgShortName(pkgPath string) string {
	parts := strings.Split(pkgPath, "/")
	last := parts[len(parts)-1]

	if len(last) >= 2 && last[0] == 'v' && last[1] >= '0' && last[1] <= '9' {
		if len(parts) >= 2 {
			candidate := parts[len(parts)-2]
			if idx := strings.LastIndex(candidate, "-"); idx >= 0 {
				return candidate[idx+1:]
			}
			return candidate
		}
	}
	return last
}
