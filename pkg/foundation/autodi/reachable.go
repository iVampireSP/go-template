package main

import (
	"fmt"
	"go/types"
	"os"
	"strings"
)

// FilterReachable returns only providers reachable from command entry points.
// A provider is reachable if its return type is consumed (directly or transitively)
// as a parameter by a command or another reachable provider.
// Pinned: //autodi:bind, //autodi:invoke, and group-path providers are always included.
func FilterReachable(
	candidates []*Provider,
	commands []*DiscoveredCommand,
	cfg *Config,
	ifaceTypes map[string]*types.Interface,
	verbose bool,
) []*Provider {
	// Pre-build type index from candidates for O(1) interface lookup
	candidateTypeIndex := make(map[string]*types.Interface)
	for _, p := range candidates {
		for _, param := range p.Params {
			if param.IsIface {
				if iface, ok := param.Type.Underlying().(*types.Interface); ok {
					candidateTypeIndex[param.TypeStr] = iface
				}
			}
			if strings.HasPrefix(param.TypeStr, "[]") {
				if sl, ok := param.Type.Underlying().(*types.Slice); ok {
					if iface, ok := sl.Elem().Underlying().(*types.Interface); ok {
						candidateTypeIndex[param.TypeStr[2:]] = iface
					}
				}
			}
		}
		for _, ret := range p.Returns {
			if ret.IsIface {
				if iface, ok := ret.Type.Underlying().(*types.Interface); ok {
					candidateTypeIndex[ret.TypeStr] = iface
				}
			}
		}
	}
	// Merge scanner ifaceTypes
	for ts, iface := range ifaceTypes {
		if _, exists := candidateTypeIndex[ts]; !exists {
			candidateTypeIndex[ts] = iface
		}
	}

	// Step 1: Index candidates by return type; classify pinned/group
	returnIndex := make(map[string][]*Provider) // typeStr → providers returning it
	reachable := make(map[*Provider]bool)       // result set
	var queue []string                          // BFS queue of needed typeStrs

	for _, p := range candidates {
		// Pin annotated providers
		if HasAnnotation(p.Annotations, AnnotBind) || HasAnnotation(p.Annotations, AnnotInvoke) {
			if !reachable[p] {
				reachable[p] = true
				for _, param := range p.Params {
					queue = append(queue, param.TypeStr)
				}
			}
		}

		// Pin group-path providers
		rel := p.RelPath(cfg.Module)
		for _, groupCfg := range cfg.Groups {
			for _, gpath := range groupCfg.Paths {
				if strings.HasPrefix(rel, gpath) && !reachable[p] {
					reachable[p] = true
					for _, param := range p.Params {
						queue = append(queue, param.TypeStr)
					}
				}
			}
		}

		// Index by return types
		for _, ret := range p.Returns {
			returnIndex[ret.TypeStr] = append(returnIndex[ret.TypeStr], p)
		}
	}

	// Step 2: Seed from command params
	for _, cmd := range commands {
		for _, param := range cmd.Params {
			queue = append(queue, param.TypeStr)
		}
	}

	// Step 3: BFS — expand needed types to find reachable providers
	visited := make(map[string]bool)
	for len(queue) > 0 {
		typeStr := queue[0]
		queue = queue[1:]
		if visited[typeStr] {
			continue
		}
		visited[typeStr] = true

		// A) Direct concrete match
		if providers, ok := returnIndex[typeStr]; ok {
			for _, p := range providers {
				if !reachable[p] {
					reachable[p] = true
					for _, param := range p.Params {
						queue = append(queue, param.TypeStr)
					}
				}
			}
			continue
		}

		// B) Interface → find implementors (use pre-built index)
		if iface, ok := candidateTypeIndex[typeStr]; ok {
			for _, p := range candidates {
				for _, ret := range p.Returns {
					if implementsIface(ret.Type, iface) && !reachable[p] {
						reachable[p] = true
						for _, param := range p.Params {
							queue = append(queue, param.TypeStr)
						}
						break
					}
				}
			}
			continue
		}

		// C) Slice-of-interface ([]SomeIface) → include ALL implementors
		if strings.HasPrefix(typeStr, "[]") {
			elemStr := typeStr[2:]
			if iface, ok := candidateTypeIndex[elemStr]; ok {
				for _, p := range candidates {
					for _, ret := range p.Returns {
						if implementsIface(ret.Type, iface) && !reachable[p] {
							reachable[p] = true
							for _, param := range p.Params {
								queue = append(queue, param.TypeStr)
							}
							break
						}
					}
				}
			}
		}
	}

	if verbose {
		for _, p := range candidates {
			if !reachable[p] {
				fmt.Fprintf(os.Stderr, "autodi: skip %s.%s (not reachable from any entry point)\n",
					p.PkgName, p.FuncName)
			}
		}
	}

	// Collect results preserving original order
	var result []*Provider
	for _, p := range candidates {
		if reachable[p] {
			result = append(result, p)
		}
	}
	return result
}

// findIfaceFromCandidates finds *types.Interface for a typeStr by searching
// a pre-built type index, then falling back to linear scan and scanner's interface type index.
func findIfaceFromCandidates(typeStr string, candidates []*Provider, ifaceTypes map[string]*types.Interface) *types.Interface {
	// Fast path: check ifaceTypes first (O(1) lookup)
	if ifaceTypes != nil {
		if iface, ok := ifaceTypes[typeStr]; ok {
			return iface
		}
	}

	// Build a local type index on first call (cached via closure in FilterReachable would be better,
	// but for correctness we do a linear scan here — this is called infrequently in practice)
	for _, p := range candidates {
		for _, param := range p.Params {
			if param.TypeStr == typeStr && param.IsIface {
				if iface, ok := param.Type.Underlying().(*types.Interface); ok {
					return iface
				}
			}
			if strings.HasPrefix(param.TypeStr, "[]") && param.TypeStr[2:] == typeStr {
				if sl, ok := param.Type.Underlying().(*types.Slice); ok {
					if iface, ok := sl.Elem().Underlying().(*types.Interface); ok {
						return iface
					}
				}
			}
		}
		for _, ret := range p.Returns {
			if ret.TypeStr == typeStr && ret.IsIface {
				if iface, ok := ret.Type.Underlying().(*types.Interface); ok {
					return iface
				}
			}
		}
	}
	return nil
}

// implementsIface checks if t implements iface, handling both T and *T.
func implementsIface(t types.Type, iface *types.Interface) bool {
	if types.Implements(t, iface) {
		return true
	}
	if _, ok := t.(*types.Pointer); !ok {
		return types.Implements(types.NewPointer(t), iface)
	}
	return false
}
