package main

import (
	"fmt"
	"strings"
)

// VerifyAcyclic checks for circular dependencies using 3-color DFS (Step 4).
// white=0 (unvisited), gray=1 (in current path), black=2 (fully processed).
func (g *Graph) VerifyAcyclic() []error {
	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[string]int)
	path := make(map[string]int) // typeStr → index in trail (for cycle reporting)
	var trail []string
	var errs []error

	var dfs func(typeStr string)
	dfs = func(typeStr string) {
		if color[typeStr] == black {
			return
		}
		if color[typeStr] == gray {
			// Found cycle — extract it from trail
			startIdx := path[typeStr]
			cycle := append(trail[startIdx:], typeStr)
			errs = append(errs, fmt.Errorf(
				"cycle dependency detected:\n  %s\nproviders involved:\n%s",
				strings.Join(cycle, " → "),
				g.formatCycleProviders(cycle),
			))
			return
		}

		color[typeStr] = gray
		path[typeStr] = len(trail)
		trail = append(trail, typeStr)

		provider := g.ProviderMap[typeStr]
		if provider != nil {
			for _, param := range provider.Params {
				depType := g.resolveType(param.TypeStr)
				dfs(depType)
			}
		}

		trail = trail[:len(trail)-1]
		delete(path, typeStr)
		color[typeStr] = black
	}

	for typeStr := range g.ProviderMap {
		if color[typeStr] == white {
			dfs(typeStr)
		}
	}

	return errs
}

// formatCycleProviders formats providers involved in a cycle for error output.
func (g *Graph) formatCycleProviders(cycle []string) string {
	var lines []string
	seen := make(map[string]bool)
	for _, typeStr := range cycle {
		if seen[typeStr] {
			continue
		}
		seen[typeStr] = true
		if p, ok := g.ProviderMap[typeStr]; ok {
			lines = append(lines, fmt.Sprintf("  %s.%s (%s)", p.PkgName, p.FuncName, p.Position))
		}
	}
	return strings.Join(lines, "\n")
}

// TopologicalSort returns providers in dependency order for the given target types.
func (g *Graph) TopologicalSort(targetTypes []string) ([]*Provider, error) {
	return g.TopologicalSortWithExtraEdges(targetTypes, nil)
}

// TopologicalSortWithExtraEdges sorts providers with additional synthetic dependency edges.
func (g *Graph) TopologicalSortWithExtraEdges(targetTypes []string, extraEdges map[string][]string) ([]*Provider, error) {
	visited := make(map[string]bool)
	var order []*Provider
	visiting := make(map[string]bool)
	added := make(map[string]bool) // track added providers by identity for O(1) dedup

	var visit func(typeStr string) error
	visit = func(typeStr string) error {
		resolved := g.resolveType(typeStr)
		if visited[resolved] {
			return nil
		}
		if visiting[resolved] {
			return fmt.Errorf("unexpected cycle at %s", resolved)
		}
		visiting[resolved] = true

		provider := g.ProviderMap[resolved]
		if provider == nil {
			visited[resolved] = true
			return nil
		}

		for _, param := range provider.Params {
			depType := g.resolveType(param.TypeStr)
			if err := visit(depType); err != nil {
				return err
			}
		}

		if extraEdges != nil {
			for _, ret := range provider.Returns {
				if extras, ok := extraEdges[ret.TypeStr]; ok {
					for _, extra := range extras {
						if err := visit(extra); err != nil {
							return err
						}
					}
				}
			}
		}

		visited[resolved] = true
		delete(visiting, resolved)

		// O(1) dedup using provider identity key
		provKey := provider.PkgPath + "." + provider.FuncName
		if !added[provKey] {
			added[provKey] = true
			order = append(order, provider)
		}

		for _, ret := range provider.Returns {
			visited[ret.TypeStr] = true
		}

		return nil
	}

	for _, target := range targetTypes {
		if err := visit(target); err != nil {
			return nil, err
		}
	}

	return order, nil
}
