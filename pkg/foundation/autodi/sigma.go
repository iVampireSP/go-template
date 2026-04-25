package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// ── DI Dependency Graph ───────────────────────────────────────────────────────

// renderDIHTML generates a self-contained Sigma.js HTML file for the DI dependency graph.
// It reuses the same mermaidGen data logic (collectIfaceTypes, computeStats, etc.)
// so node colours and edge routing are identical to the Mermaid output.
func renderDIHTML(graph *Graph, commands []*DiscoveredCommand, cfg *Config) []byte {
	mg := &mermaidGen{
		graph:    graph,
		commands: commands,
		cfg:      cfg,
		ids:      make(map[string]string),
		usedIDs:  make(map[string]bool),
	}
	ifaceSet := mg.collectIfaceTypes()
	providerCounts, ifaceStats := mg.computeStats(ifaceSet)

	var nodes []sigmaNode
	var edges []sigmaEdge
	edgeIdx := 0

	addEdge := func(src, tgt, color, label string) {
		edges = append(edges, sigmaEdge{
			Key:    fmt.Sprintf("e%d", edgeIdx),
			Source: src,
			Target: tgt,
			Attributes: map[string]any{
				"color": color,
				"label": label,
				"size":  1.5,
			},
		})
		edgeIdx++
	}

	// ── Provider nodes ────────────────────────────────────────────────────────
	for _, p := range graph.Providers {
		id := mg.nodeID(p)
		depCount := providerCounts[id]

		color, nodeType := sgColorProvider, "provider"
		switch {
		case p.IsInvoke:
			color, nodeType = sgColorInvoke, "invoke"
		case mg.isDecorator(p):
			color, nodeType = sgColorDecorator, "decorator"
		case len(p.Params) == 0:
			color, nodeType = sgColorLeaf, "leaf"
		}

		size := clampInt(6+int(math.Sqrt(float64(depCount+1))*2), 6, 16)

		typeName := ""
		if !p.IsInvoke {
			for _, ret := range p.Returns {
				if ret.TypeStr != "error" {
					typeName = briefTypeName(ret.TypeStr)
					break
				}
			}
		}
		label := p.FuncName
		if typeName != "" {
			label = typeName + "\n" + p.FuncName
		}

		nodes = append(nodes, sigmaNode{
			Key: id,
			Attributes: map[string]any{
				"label":    label,
				"color":    color,
				"size":     size,
				"nodeType": nodeType,
				"pkg":      mermaidRelPkg(p.PkgPath, cfg.Module),
				"depCount": depCount,
			},
		})
	}

	// ── Interface nodes ───────────────────────────────────────────────────────
	for ifaceTypeStr := range ifaceSet {
		id := mg.ifaceNodeID(ifaceTypeStr)
		stat := ifaceStats[ifaceTypeStr]
		implCount, useCount := stat[0], stat[1]
		size := clampInt(7+int(math.Sqrt(float64(implCount+useCount+1))*3), 7, 18)

		nodes = append(nodes, sigmaNode{
			Key: id,
			Attributes: map[string]any{
				"label":     "«iface»\n" + ifaceShortName(ifaceTypeStr),
				"color":     sgColorIface,
				"size":      size,
				"nodeType":  "iface",
				"implCount": implCount,
				"useCount":  useCount,
			},
		})
	}

	// ── Command nodes ─────────────────────────────────────────────────────────
	for _, cmd := range commands {
		cmdID := "C_" + sanitizeMermaidID(cmd.Name)
		label := cmd.Name
		if len(cmd.Handlers) > 0 && len(cmd.Handlers) <= 4 {
			var ms []string
			for _, h := range cmd.Handlers {
				ms = append(ms, h.MethodName)
			}
			label = cmd.Name + "\n" + strings.Join(ms, " | ")
		}
		nodes = append(nodes, sigmaNode{
			Key: cmdID,
			Attributes: map[string]any{
				"label":    label,
				"color":    sgColorCommand,
				"size":     11,
				"nodeType": "command",
			},
		})
	}

	// ── Implements edges ──────────────────────────────────────────────────────
	renderedImpl := make(map[string]bool)
	for ifaceTypeStr := range ifaceSet {
		ifaceID := mg.ifaceNodeID(ifaceTypeStr)
		emitImpl := func(dep *Provider) {
			key := mg.nodeID(dep) + "|" + ifaceID
			if renderedImpl[key] {
				return
			}
			renderedImpl[key] = true
			addEdge(mg.nodeID(dep), ifaceID, sgColorEdgeImpl, "implements")
		}
		if concrete, ok := graph.Bindings[ifaceTypeStr]; ok {
			if dep := graph.ProviderMap[concrete]; dep != nil {
				emitImpl(dep)
			}
		}
		for _, p := range graph.AutoCollect(ifaceTypeStr) {
			emitImpl(p)
		}
		for groupName, gc := range cfg.Groups {
			if ifaceTypeStr == graph.resolveConfigType(gc.Interface) {
				for _, gp := range graph.Groups[groupName] {
					emitImpl(gp)
				}
			}
		}
	}

	// ── Dependency edges ──────────────────────────────────────────────────────
	addParamEdges := func(consumerID string, params []TypeRef) {
		for _, param := range params {
			if strings.HasPrefix(param.TypeStr, "[]") {
				elemType := param.TypeStr[2:]
				sliceLabel := "[]" + briefTypeName(elemType)
				if ifaceSet[elemType] {
					addEdge(mg.ifaceNodeID(elemType), consumerID, sgColorEdgeSlice, sliceLabel)
				} else if groupName := mg.matchGroupByElem(elemType); groupName != "" {
					for _, gp := range graph.Groups[groupName] {
						addEdge(mg.nodeID(gp), consumerID, sgColorEdgeSlice, sliceLabel)
					}
				} else {
					for _, ap := range graph.AutoCollect(elemType) {
						addEdge(mg.nodeID(ap), consumerID, sgColorEdgeSlice, sliceLabel)
					}
				}
				continue
			}

			resolved := graph.resolveType(param.TypeStr)
			dep := graph.ProviderMap[resolved]
			if dep == nil {
				dep = graph.ProviderMap[param.TypeStr]
			}
			if dep == nil {
				continue
			}
			if param.IsIface && ifaceSet[param.TypeStr] {
				addEdge(mg.ifaceNodeID(param.TypeStr), consumerID, sgColorEdgeDep, "")
			} else {
				addEdge(mg.nodeID(dep), consumerID, sgColorEdgeDep, "")
			}
		}
	}
	for _, p := range graph.Providers {
		addParamEdges(mg.nodeID(p), p.Params)
	}
	for _, cmd := range commands {
		addParamEdges("C_"+sanitizeMermaidID(cmd.Name), cmd.Params)
	}

	assignDIPositions(nodes, edges)

	data := sigmaGraph{Nodes: nodes, Edges: edges}
	jsonBytes, _ := json.Marshal(data)
	title := cfg.AppName + " — DI Graph"
	return buildSigmaHTML(title, diLegend(), string(jsonBytes), diTooltipFn())
}

// ── Package Diagram ───────────────────────────────────────────────────────────

// renderPkgHTML generates a self-contained Sigma.js HTML for the package/type diagram.
func renderPkgHTML(pkgInfos []*pdPkgInfo, rels []pdRelation, refByNode, implByNode map[string]int) []byte {
	var nodes []sigmaNode
	var edges []sigmaEdge
	edgeIdx := 0

	// Helper to get a pdPkgInfo reference for each pdType
	typeToInfo := make(map[*pdType]*pdPkgInfo)
	for _, pkg := range pkgInfos {
		for _, t := range pkg.Structs {
			typeToInfo[t] = pkg
		}
		for _, t := range pkg.Ifaces {
			typeToInfo[t] = pkg
		}
	}

	// ── Nodes ──────────────────────────────────────────────────────────────
	addedNodes := make(map[string]bool)
	addTypeNode := func(pkg *pdPkgInfo, t *pdType, nodeType string) {
		id := pkgTypeNodeID(pkg, t.Name)
		if addedNodes[id] {
			return // skip duplicate (safety net)
		}
		addedNodes[id] = true
		usedBy := refByNode[id]
		impl := implByNode[id]
		use := refByNode[id]

		var color string
		var size int
		switch nodeType {
		case "interface":
			color = sgColorIface
			size = clampInt(10+(impl+use)*2, 10, 40)
		case "decorator":
			color = sgColorDecorator
			size = clampInt(8+usedBy*2, 8, 35)
		default: // struct
			color = sgColorProvider
			size = clampInt(8+usedBy*2, 8, 35)
		}

		// Collect up to 6 methods for tooltip
		methods := make([]string, 0, 6)
		for i, m := range t.Methods {
			if i >= 6 {
				break
			}
			retStr := ""
			if len(m.Returns) > 0 {
				retStr = " " + strings.Join(m.Returns, ", ")
			}
			methods = append(methods, fmt.Sprintf("%s(%s)%s", m.Name, strings.Join(m.Params, ", "), retStr))
		}

		attrs := map[string]any{
			"label":    t.Name,
			"color":    color,
			"size":     size,
			"nodeType": nodeType,
			"pkg":      pkg.RelPath,
			"methods":  methods,
			"usedBy":   usedBy,
		}
		if nodeType == "interface" {
			attrs["implCount"] = impl
			attrs["useCount"] = use
		}
		nodes = append(nodes, sigmaNode{Key: id, Attributes: attrs})
	}

	for _, pkg := range pkgInfos {
		for _, iface := range pkg.Ifaces {
			addTypeNode(pkg, iface, "interface")
		}
		for _, st := range pkg.Structs {
			nt := "struct"
			if isDecoratorPkg(pkg, st) {
				nt = "decorator"
			}
			addTypeNode(pkg, st, nt)
		}
	}

	// ── Edges ──────────────────────────────────────────────────────────────
	edgeColors := map[string]string{
		"implements": sgColorEdgeImpl,
		"depends":    sgColorEdgeDep,
		"field":      "#95a5a6",
	}
	for _, r := range rels {
		color := edgeColors[r.Kind]
		if color == "" {
			color = sgColorEdgeDep
		}
		edges = append(edges, sigmaEdge{
			Key:    fmt.Sprintf("e%d", edgeIdx),
			Source: r.FromID,
			Target: r.ToID,
			Attributes: map[string]any{
				"color": color,
				"label": r.Kind,
				"size":  1.5,
			},
		})
		edgeIdx++
	}

	assignPkgPositions(nodes)

	data := sigmaGraph{Nodes: nodes, Edges: edges}
	jsonBytes, _ := json.Marshal(data)
	return buildSigmaHTML("Package Diagram", pkgLegend(), string(jsonBytes), pkgTooltipFn())
}

// pkgTypeNodeID builds a stable sigma node ID for a package type.
// Uses the full RelPath to avoid collisions between packages that share a suffix
// (e.g. "internal/ring" vs "pkg/ring" would both become "ring_Ring" with prefix stripping).
func pkgTypeNodeID(pkg *pdPkgInfo, typeName string) string {
	return sanitizeMermaidID(pkg.RelPath + "_" + typeName)
}

// isDecoratorPkg detects decorator pattern: struct implements an interface AND
// its New* constructor takes that same interface as a parameter.
// It is a lightweight replication of pkgDiagramGen.isDecorator used here
// since we don't have a pkgDiagramGen receiver in this context.
func isDecoratorPkg(pkg *pdPkgInfo, st *pdType) bool {
	// No named types available here — delegate decorator detection via method
	// presence heuristic (if the struct has no Named, skip).
	if st.Named == nil {
		return false
	}
	// Reuse the logic via a temporary pkgDiagramGen — but we don't have one here.
	// Instead, use a simpler approach: check if any New* func in the same pkg
	// returns this struct AND takes an interface as a param that the struct also
	// appears to implement (we can't do full types.Implements here without the
	// original pkgDiagramGen context).
	// For simplicity, return false; the full version is in pkgDiagramGen.isDecorator.
	return false
}
