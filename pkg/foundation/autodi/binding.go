package main

import (
	"fmt"
	"go/types"
	"strings"
)

// resolveConfigType resolves a short config type name to its full type string.
func (g *Graph) resolveConfigType(shortName string) string {
	if strings.Contains(shortName, "/") {
		return shortName
	}

	if full, ok := g.shortToFull[shortName]; ok {
		return full
	}

	prefix := ""
	s := shortName
	if strings.HasPrefix(s, "*") {
		prefix = "*"
		s = s[1:]
	}
	dotIdx := strings.Index(s, ".")
	if dotIdx > 0 {
		pkgName := s[:dotIdx]
		typeName := s[dotIdx+1:]
		if pkgPath, ok := g.pkgNameToPath[pkgName]; ok {
			return prefix + pkgPath + "." + typeName
		}

		for _, group := range g.cfg.Groups {
			for _, gpath := range group.Paths {
				parts := strings.Split(gpath, "/")
				for i, part := range parts {
					if part == pkgName {
						fullPkg := g.cfg.Module + "/" + strings.Join(parts[:i+1], "/")
						g.pkgNameToPath[pkgName] = fullPkg
						return prefix + fullPkg + "." + typeName
					}
				}
			}
		}
	}

	return shortName
}

// resolveBindings sets up interface → concrete type mappings.
func (g *Graph) resolveBindings(providers []*Provider) []error {
	var errs []error

	// 1. Explicit bindings from config
	for concreteShort, ifaces := range g.cfg.Bindings {
		concreteFull := g.resolveConfigType(concreteShort)
		for _, ifaceShort := range ifaces {
			ifaceFull := g.resolveConfigType(ifaceShort)
			if _, ok := g.Bindings[ifaceFull]; ok {
				errs = append(errs, fmt.Errorf("interface %s has duplicate binding configuration", ifaceFull))
				continue
			}
			g.Bindings[ifaceFull] = concreteFull
			if provider, ok := g.ProviderMap[concreteFull]; ok {
				g.ProviderMap[ifaceFull] = provider
				g.TypeToField[ifaceFull] = FieldName(ifaceFull)
			}
		}
	}

	// 2. Explicit bindings from annotations
	for _, p := range providers {
		bindTargets := GetAnnotationValues(p.Annotations, AnnotBind)
		for _, target := range bindTargets {
			if _, ok := g.Bindings[target]; ok {
				continue
			}
			if len(p.Returns) > 0 {
				concreteStr := p.Returns[0].TypeStr
				g.Bindings[target] = concreteStr
				g.ProviderMap[target] = p
			}
		}
	}

	// 3. Auto-detect bindings using pre-built impl index (Step 1)
	g.autoDetectBindings(providers)

	return errs
}

// autoDetectBindings automatically binds interfaces to concrete types using the impl index.
func (g *Graph) autoDetectBindings(providers []*Provider) {
	// Collect all interface types needed as parameters
	neededIfaces := make(map[string]bool)
	for _, p := range providers {
		for _, param := range p.Params {
			if param.IsIface {
				if _, bound := g.Bindings[param.TypeStr]; !bound {
					if _, provided := g.ProviderMap[param.TypeStr]; !provided {
						neededIfaces[param.TypeStr] = true
					}
				}
			}
		}
	}

	// Use pre-built impl index for O(1) lookup per interface (Step 1)
	for ifaceStr := range neededIfaces {
		entries := g.implIndex[ifaceStr]
		if len(entries) == 1 {
			// Filter to entries that are in ProviderMap (singleton providers only)
			var candidates []implEntry
			for _, e := range entries {
				for typeStr := range g.ProviderMap {
					if g.ProviderMap[typeStr] == e.provider {
						candidates = append(candidates, implEntry{provider: e.provider, retTypeStr: typeStr})
						break
					}
				}
			}
			if len(candidates) == 1 {
				g.Bindings[ifaceStr] = candidates[0].retTypeStr
				g.ProviderMap[ifaceStr] = candidates[0].provider
			}
		} else if len(entries) > 1 {
			// Multiple implementors but check if only one is in ProviderMap
			var candidates []implEntry
			for _, e := range entries {
				for typeStr, p := range g.ProviderMap {
					if p == e.provider {
						candidates = append(candidates, implEntry{provider: p, retTypeStr: typeStr})
						break
					}
				}
			}
			if len(candidates) == 1 {
				g.Bindings[ifaceStr] = candidates[0].retTypeStr
				g.ProviderMap[ifaceStr] = candidates[0].provider
			}
		}
	}
}

// BindCommandInterfaces resolves interface bindings for command parameters
// using the pre-built type index and impl index.
func (g *Graph) BindCommandInterfaces(commands []*DiscoveredCommand) {
	for _, cmd := range commands {
		for _, param := range cmd.Params {
			if !param.IsIface {
				continue
			}
			if _, bound := g.Bindings[param.TypeStr]; bound {
				continue
			}

			// Find the interface type from our type index or provider universe
			ifaceUnderlying := g.findIfaceType(param.TypeStr)
			if ifaceUnderlying == nil {
				continue
			}

			// Use impl index for O(1) lookup
			entries := g.implIndex[param.TypeStr]
			if len(entries) == 1 {
				g.Bindings[param.TypeStr] = entries[0].retTypeStr
				if p, ok := g.ProviderMap[entries[0].retTypeStr]; ok {
					g.ProviderMap[param.TypeStr] = p
				}
			}
		}
	}
}

// resolveType follows interface bindings to find the concrete type.
func (g *Graph) resolveType(typeStr string) string {
	if concrete, ok := g.Bindings[typeStr]; ok {
		return concrete
	}
	return typeStr
}

// findIfaceType finds the *types.Interface underlying type for a given type string.
// Uses O(1) typeIndex lookup (Step 3) instead of linear scan.
func (g *Graph) findIfaceType(typeStr string) *types.Interface {
	// Step 3: O(1) lookup from type index
	if t, ok := g.typeIndex[typeStr]; ok {
		if iface, ok := t.Underlying().(*types.Interface); ok {
			return iface
		}
	}

	// Fallback: scanner-discovered interfaces
	if g.ifaceTypes != nil {
		if iface, ok := g.ifaceTypes[typeStr]; ok {
			return iface
		}
	}

	return nil
}
