package main

import (
	"fmt"
	"go/types"
	"sort"
	"strings"
)

// implEntry records that a provider's return type implements an interface.
type implEntry struct {
	provider   *Provider
	retTypeStr string
}

// Graph holds the resolved dependency graph.
type Graph struct {
	Providers   []*Provider
	ProviderMap map[string]*Provider   // typeStr → provider
	Bindings    map[string]string      // interface typeStr → concrete typeStr
	Groups      map[string][]*Provider // group name → providers
	TypeToField map[string]string      // typeStr → Container field name

	cfg           *Config
	shortToFull   map[string]string           // short type name → full type string
	pkgNameToPath map[string]string           // pkg short name → full pkg path
	ifaceTypes    map[string]*types.Interface // full typeStr → interface type from loaded packages

	// Performance indexes (built once, queried many times)
	typeIndex    map[string]types.Type  // typeStr → types.Type (Step 3)
	implIndex    map[string][]implEntry // ifaceTypeStr → implementors (Step 1)
	implCache    map[implCacheKey]bool  // cached types.Implements results (Step 2)
	fieldToGroup map[string]string      // fieldName → groupName reverse index (Step 5)
	sortedTypes  []string               // pre-sorted ProviderMap keys (Step 7)
}

// implCacheKey is the key for caching types.Implements() results.
type implCacheKey struct {
	typeStr  string
	ifaceStr string
}

// BuildGraph constructs the dependency graph from discovered providers.
func BuildGraph(providers []*Provider, cfg *Config, pkgIndex map[string]string, ifaceTypes map[string]*types.Interface) (*Graph, []error) {
	g := &Graph{
		Providers:     providers,
		ProviderMap:   make(map[string]*Provider),
		Bindings:      make(map[string]string),
		Groups:        make(map[string][]*Provider),
		TypeToField:   make(map[string]string),
		cfg:           cfg,
		shortToFull:   make(map[string]string),
		pkgNameToPath: make(map[string]string),
		ifaceTypes:    ifaceTypes,
		typeIndex:     make(map[string]types.Type),
		implCache:     make(map[implCacheKey]bool),
		fieldToGroup:  make(map[string]string),
	}

	// Seed pkgNameToPath with the full package index from scanner
	for name, path := range pkgIndex {
		g.pkgNameToPath[name] = path
	}

	// Build short-to-full type name mapping and typeStr→Type index
	g.buildTypeIndex(providers)

	var errs []error

	// Phase 1: Classify providers into groups
	for _, p := range providers {
		rel := p.RelPath(cfg.Module)
		for groupName, groupCfg := range cfg.Groups {
			for _, gpath := range groupCfg.Paths {
				if strings.HasPrefix(rel, gpath) {
					p.Groups = append(p.Groups, groupName)
				}
			}
		}
	}

	// Phase 2: Register each provider's return types in the provider map
	for _, p := range providers {
		if p.IsInvoke {
			continue
		}

		for _, ret := range p.Returns {
			typeStr := ret.TypeStr

			// Skip grouped providers from the singleton map
			if len(p.Groups) > 0 {
				continue
			}

			if existing, ok := g.ProviderMap[typeStr]; ok {
				errs = append(errs, fmt.Errorf(
					"type %s has multiple providers:\n  1. %s.%s (%s)\n  2. %s.%s (%s)\n  hint: mark one with //autodi:ignore",
					typeStr,
					existing.PkgName, existing.FuncName, existing.Position,
					p.PkgName, p.FuncName, p.Position,
				))
				continue
			}
			g.ProviderMap[typeStr] = p
			g.TypeToField[typeStr] = FieldName(typeStr)
		}
	}

	// Add grouped providers + build fieldToGroup reverse index (Step 5)
	for _, p := range providers {
		for _, groupName := range p.Groups {
			g.Groups[groupName] = append(g.Groups[groupName], p)
		}
	}
	for name := range g.Groups {
		g.fieldToGroup[GroupFieldName(name)] = name
	}

	// Build pre-sorted provider keys (Step 7)
	g.rebuildSortedTypes()

	// Build interface→implementors index (Step 1) — after ProviderMap is populated
	g.buildImplIndex()

	// Phase 3: Resolve interface bindings
	bindErrs := g.resolveBindings(providers)
	errs = append(errs, bindErrs...)

	if len(errs) > 0 {
		return nil, errs
	}
	return g, nil
}

// rebuildSortedTypes rebuilds the pre-sorted ProviderMap keys.
func (g *Graph) rebuildSortedTypes() {
	g.sortedTypes = make([]string, 0, len(g.ProviderMap))
	for typeStr := range g.ProviderMap {
		g.sortedTypes = append(g.sortedTypes, typeStr)
	}
	sort.Strings(g.sortedTypes)
}

// buildTypeIndex builds lookup maps from all discovered types.
func (g *Graph) buildTypeIndex(providers []*Provider) {
	for _, p := range providers {
		g.pkgNameToPath[p.PkgName] = p.PkgPath

		for _, ret := range p.Returns {
			short := toShortTypeName(ret.TypeStr)
			if short != ret.TypeStr {
				g.shortToFull[short] = ret.TypeStr
			}
			if ret.PkgPath != "" {
				parts := strings.Split(ret.PkgPath, "/")
				g.pkgNameToPath[parts[len(parts)-1]] = ret.PkgPath
			}
			// Step 3: index typeStr → Type
			g.typeIndex[ret.TypeStr] = ret.Type
		}

		for _, param := range p.Params {
			short := toShortTypeName(param.TypeStr)
			if short != param.TypeStr {
				g.shortToFull[short] = param.TypeStr
			}
			if param.PkgPath != "" {
				parts := strings.Split(param.PkgPath, "/")
				g.pkgNameToPath[parts[len(parts)-1]] = param.PkgPath
			}
			// Step 3: index typeStr → Type
			g.typeIndex[param.TypeStr] = param.Type
			// Also index slice element types
			if strings.HasPrefix(param.TypeStr, "[]") {
				elemStr := param.TypeStr[2:]
				if sliceType, ok := param.Type.Underlying().(*types.Slice); ok {
					g.typeIndex[elemStr] = sliceType.Elem()
				}
			}
		}
	}
}

// buildImplIndex pre-computes the interface→implementors index and populates the implCache.
func (g *Graph) buildImplIndex() {
	g.implIndex = make(map[string][]implEntry)

	// Collect all known interface types from providers + ifaceTypes
	allIfaces := make(map[string]*types.Interface)
	for typeStr, t := range g.typeIndex {
		if iface, ok := t.Underlying().(*types.Interface); ok {
			allIfaces[typeStr] = iface
		}
	}
	for typeStr, iface := range g.ifaceTypes {
		if _, exists := allIfaces[typeStr]; !exists {
			allIfaces[typeStr] = iface
		}
	}

	// For each interface, check which providers implement it
	for ifaceStr, iface := range allIfaces {
		for _, p := range g.Providers {
			if p.IsInvoke {
				continue
			}
			for _, ret := range p.Returns {
				if g.cachedImplements(ret.Type, ret.TypeStr, iface, ifaceStr) {
					g.implIndex[ifaceStr] = append(g.implIndex[ifaceStr], implEntry{
						provider:   p,
						retTypeStr: ret.TypeStr,
					})
					break
				}
				// Also check *T
				if _, isPtr := ret.Type.(*types.Pointer); !isPtr {
					ptrType := types.NewPointer(ret.Type)
					ptrStr := "*" + ret.TypeStr
					if g.cachedImplements(ptrType, ptrStr, iface, ifaceStr) {
						g.implIndex[ifaceStr] = append(g.implIndex[ifaceStr], implEntry{
							provider:   p,
							retTypeStr: ret.TypeStr,
						})
						break
					}
				}
			}
		}
		// Sort entries by PkgPath for deterministic output
		entries := g.implIndex[ifaceStr]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].provider.PkgPath < entries[j].provider.PkgPath
		})
	}
}

// cachedImplements checks types.Implements with caching (Step 2).
func (g *Graph) cachedImplements(t types.Type, tStr string, iface *types.Interface, ifaceStr string) bool {
	key := implCacheKey{typeStr: tStr, ifaceStr: ifaceStr}
	if result, ok := g.implCache[key]; ok {
		return result
	}
	result := types.Implements(t, iface)
	g.implCache[key] = result
	return result
}

func isPointer(t types.Type) bool {
	_, ok := t.(*types.Pointer)
	return ok
}

// AllSingletonProviders returns all non-group, non-invoke providers in dependency order.
func (g *Graph) AllSingletonProviders() ([]*Provider, error) {
	// Use pre-sorted keys (Step 7)
	return g.TopologicalSort(g.sortedTypes)
}

// EntryProviders returns the singleton providers needed for an entry point, in dependency order.
func (g *Graph) EntryProviders(fieldNames []string) ([]*Provider, error) {
	// Build reverse map: fieldName → typeStr
	fieldToType := make(map[string]string)
	for typeStr, fieldName := range g.TypeToField {
		fieldToType[fieldName] = typeStr
	}

	needed := make(map[string]bool)

	for _, fieldName := range fieldNames {
		// O(1) group lookup via reverse index (Step 5)
		if groupName, ok := g.fieldToGroup[fieldName]; ok {
			for _, p := range g.Groups[groupName] {
				for _, param := range p.Params {
					needed[param.TypeStr] = true
				}
			}
			continue
		}

		if typeStr, ok := fieldToType[fieldName]; ok {
			needed[typeStr] = true
		}
	}

	// Use shared expansion helper (Step 6)
	expanded := g.expandTransitive(needed)

	var targets []string
	for t := range expanded {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	return g.TopologicalSort(targets)
}

// expandTransitive expands a set of needed types to include all transitive dependencies
// and invoke providers whose dependencies are satisfied. Shared helper for Step 6.
func (g *Graph) expandTransitive(needed map[string]bool) map[string]bool {
	expanded := make(map[string]bool)
	var expand func(string)
	expand = func(typeStr string) {
		resolved := g.resolveType(typeStr)
		if expanded[resolved] {
			return
		}
		expanded[resolved] = true

		provider := g.ProviderMap[resolved]
		if provider == nil {
			return
		}
		for _, param := range provider.Params {
			expand(param.TypeStr)
		}
	}

	for t := range needed {
		expand(t)
	}

	// Include invoke providers whose dependencies are all satisfied
	for _, p := range g.Providers {
		if !p.IsInvoke {
			continue
		}
		allSatisfied := true
		for _, param := range p.Params {
			resolved := g.resolveType(param.TypeStr)
			if !expanded[resolved] {
				allSatisfied = false
				break
			}
		}
		if allSatisfied {
			for _, ret := range p.Returns {
				expanded[ret.TypeStr] = true
			}
		}
	}

	return expanded
}

// ValidateEntry checks that all providers for an entry have their dependencies satisfied.
func (g *Graph) ValidateEntry(name string, providers []*Provider) []error {
	provided := make(map[string]bool)
	for _, p := range providers {
		for _, ret := range p.Returns {
			provided[ret.TypeStr] = true
		}
	}
	for iface, concrete := range g.Bindings {
		if provided[concrete] {
			provided[iface] = true
		}
	}

	var errs []error
	for _, p := range providers {
		for _, param := range p.Params {
			if param.Optional {
				continue
			}
			resolved := g.resolveType(param.TypeStr)
			if !provided[resolved] {
				if strings.HasPrefix(param.TypeStr, "[]") {
					elemType := param.TypeStr[2:]
					if autoProviders := g.AutoCollect(elemType); len(autoProviders) > 0 {
						continue
					}
				}
				errs = append(errs, fmt.Errorf(
					"entry %q: %s.%s missing dependency %s",
					name, p.PkgName, p.FuncName, toShortTypeName(param.TypeStr),
				))
			}
		}
	}
	return errs
}

// fieldNameToGroup returns the group name for a Container field name, or "" if not a group.
// Uses O(1) reverse index lookup (Step 5).
func (g *Graph) fieldNameToGroup(fieldName string) string {
	return g.fieldToGroup[fieldName]
}

// GroupFieldName converts a group config name to a Container field name.
// "admin_controllers" → "AdminControllers"
// Uses strings.Builder instead of += (Step 8).
func GroupFieldName(name string) string {
	parts := strings.Split(name, "_")
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(exportName(p))
	}
	return b.String()
}

// ProvidersForTypes returns singleton providers needed for the given type strings, in dependency order.
// Step 6: delegates to ProvidersForTypesWithExtraEdges to eliminate duplication.
func (g *Graph) ProvidersForTypes(typeStrs []string) ([]*Provider, error) {
	return g.ProvidersForTypesWithExtraEdges(typeStrs, nil)
}

// ProvidersForTypesWithExtraEdges is like ProvidersForTypes but accepts extra synthetic
// dependency edges for the topological sort.
func (g *Graph) ProvidersForTypesWithExtraEdges(typeStrs []string, extraEdges map[string][]string) ([]*Provider, error) {
	needed := make(map[string]bool)
	for _, t := range typeStrs {
		needed[t] = true
	}
	expanded := g.expandTransitive(needed)

	var targets []string
	for t := range expanded {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	return g.TopologicalSortWithExtraEdges(targets, extraEdges)
}

// AutoCollect scans all providers and returns those whose return type implements
// the given interface type string. Uses the pre-built impl index (Step 1).
func (g *Graph) AutoCollect(elemTypeStr string) []*Provider {
	// Use impl index for O(1) lookup
	entries := g.implIndex[elemTypeStr]
	if len(entries) == 0 {
		return nil
	}

	// Filter out invoke providers (already filtered during index build)
	matches := make([]*Provider, 0, len(entries))
	for _, e := range entries {
		matches = append(matches, e.provider)
	}
	// Already sorted by PkgPath during index build
	return matches
}
