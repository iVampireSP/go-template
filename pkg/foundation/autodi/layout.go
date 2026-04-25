package main

import (
	"math"
	"sort"
	"strings"
)

// clampInt returns v clamped to [lo, hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// assignDIPositions computes hierarchical x/y positions for the DI dependency graph.
// Edges flow from dependency (source) to consumer (target).
// Layout goals:
//   - keep directionality (deps on top, consumers below),
//   - reduce edge crossings within layers,
//   - avoid node collisions with size-aware spacing,
//   - wrap overly wide layers to prevent single long lines.
func assignDIPositions(nodes []sigmaNode, edges []sigmaEdge) {
	if len(nodes) == 0 {
		return
	}

	nodeIdx := make(map[string]int, len(nodes))
	for i, n := range nodes {
		nodeIdx[n.Key] = i
	}

	nodeSize := func(key string) float64 {
		idx := nodeIdx[key]
		v, ok := nodes[idx].Attributes["size"]
		if !ok {
			return 10
		}
		switch t := v.(type) {
		case int:
			return float64(t)
		case int32:
			return float64(t)
		case int64:
			return float64(t)
		case float32:
			return float64(t)
		case float64:
			return t
		default:
			return 10
		}
	}

	nodeBoxHalf := func(key string) (float64, float64) {
		idx := nodeIdx[key]
		label, _ := nodes[idx].Attributes["label"].(string)
		lines := strings.Split(label, "\n")
		if len(lines) == 0 {
			lines = []string{""}
		}

		maxChars := 0
		for _, line := range lines {
			if l := len([]rune(line)); l > maxChars {
				maxChars = l
			}
		}

		textW := float64(maxChars)*7.0 + 24.0
		textH := float64(len(lines))*15.0 + 12.0

		r := math.Max(20.0, nodeSize(key)*2.8)
		halfW := math.Max(r, textW/2.0)
		halfH := math.Max(r*0.7, textH/2.0)
		return halfW, halfH
	}

	// Build adjacency: inDegree and outgoing neighbours.
	outAdj := make(map[string][]string, len(nodes))
	inAdj := make(map[string][]string, len(nodes))
	inDegree := make(map[string]int, len(nodes))
	for _, n := range nodes {
		outAdj[n.Key] = nil
		inAdj[n.Key] = nil
		inDegree[n.Key] = 0
	}
	for _, e := range edges {
		if _, ok := nodeIdx[e.Source]; !ok {
			continue
		}
		if _, ok := nodeIdx[e.Target]; !ok {
			continue
		}
		outAdj[e.Source] = append(outAdj[e.Source], e.Target)
		inAdj[e.Target] = append(inAdj[e.Target], e.Source)
		inDegree[e.Target]++
	}
	for _, n := range nodes {
		sort.Strings(outAdj[n.Key])
		sort.Strings(inAdj[n.Key])
	}

	// Kahn's BFS with longest-path layer assignment.
	layer := make(map[string]int, len(nodes))
	queue := []string{}
	for _, n := range nodes {
		if inDegree[n.Key] == 0 {
			layer[n.Key] = 0
			queue = append(queue, n.Key)
		}
	}
	sort.Strings(queue)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range outAdj[cur] {
			if layer[cur]+1 > layer[next] {
				layer[next] = layer[cur] + 1
			}
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	// Nodes in cycles not yet assigned — place them one past the max layer.
	maxLayer := 0
	for _, l := range layer {
		if l > maxLayer {
			maxLayer = l
		}
	}
	for _, n := range nodes {
		if _, ok := layer[n.Key]; !ok {
			layer[n.Key] = maxLayer + 1
		}
	}

	// Group nodes by layer; sort within each layer for stable output.
	layerNodes := make(map[int][]string)
	for key, l := range layer {
		layerNodes[l] = append(layerNodes[l], key)
	}
	for l := range layerNodes {
		sort.Strings(layerNodes[l])
	}

	// Crossing reduction: barycenter sweeps.
	order := make(map[string]float64, len(nodes))
	for l := 0; l <= maxLayer+1; l++ {
		for i, key := range layerNodes[l] {
			order[key] = float64(i)
		}
	}

	barySort := func(l int, deps map[string][]string, neighbourOK func(int, int) bool) {
		keys := layerNodes[l]
		if len(keys) <= 1 {
			return
		}
		sort.SliceStable(keys, func(i, j int) bool {
			avg := func(key string) float64 {
				var sum float64
				count := 0
				for _, nb := range deps[key] {
					if !neighbourOK(layer[nb], l) {
						continue
					}
					sum += order[nb]
					count++
				}
				if count == 0 {
					return order[key]
				}
				return sum / float64(count)
			}
			bi := avg(keys[i])
			bj := avg(keys[j])
			if bi == bj {
				return keys[i] < keys[j]
			}
			return bi < bj
		})
		layerNodes[l] = keys
		for i, key := range keys {
			order[key] = float64(i)
		}
	}

	for pass := 0; pass < 3; pass++ {
		for l := 1; l <= maxLayer+1; l++ {
			barySort(l, inAdj, func(nbLayer, curLayer int) bool { return nbLayer < curLayer })
		}
		for l := maxLayer; l >= 0; l-- {
			barySort(l, outAdj, func(nbLayer, curLayer int) bool { return nbLayer > curLayer })
		}
	}

	const layerGapY = 180.0
	const rowGapY = 56.0
	const baseGapX = 28.0
	const maxRowWidth = 9200.0

	y := 0.0
	for l := 0; l <= maxLayer+1; l++ {
		keys := layerNodes[l]
		if len(keys) == 0 {
			continue
		}

		// Split each layer into multiple rows by real text-aware width, so labels don't collide.
		var rows [][]string
		var curr []string
		currW := 0.0
		for _, key := range keys {
			halfW, _ := nodeBoxHalf(key)
			addW := 2.0 * halfW
			if len(curr) > 0 {
				prevHalfW, _ := nodeBoxHalf(curr[len(curr)-1])
				addW = prevHalfW + baseGapX + halfW
			}
			if len(curr) > 0 && currW+addW > maxRowWidth {
				rows = append(rows, curr)
				curr = []string{key}
				currW = 2.0 * halfW
				continue
			}
			curr = append(curr, key)
			currW += addW
		}
		if len(curr) > 0 {
			rows = append(rows, curr)
		}

		rowHeights := make([]float64, len(rows))
		for i, rowKeys := range rows {
			maxHalfH := 0.0
			for _, key := range rowKeys {
				_, halfH := nodeBoxHalf(key)
				if halfH > maxHalfH {
					maxHalfH = halfH
				}
			}
			rowHeights[i] = 2.0 * maxHalfH
		}

		layerH := 0.0
		for i, h := range rowHeights {
			if i > 0 {
				layerH += rowGapY
			}
			layerH += h
		}

		yCursor := y
		for rowIdx, rowKeys := range rows {
			if len(rowKeys) == 0 {
				continue
			}

			rowW := 0.0
			for i, key := range rowKeys {
				halfW, _ := nodeBoxHalf(key)
				if i == 0 {
					rowW += 2.0 * halfW
					continue
				}
				prevHalfW, _ := nodeBoxHalf(rowKeys[i-1])
				rowW += prevHalfW + baseGapX + halfW
			}

			x := -rowW / 2.0
			for i, key := range rowKeys {
				halfW, _ := nodeBoxHalf(key)
				if i == 0 {
					x += halfW
				} else {
					prevHalfW, _ := nodeBoxHalf(rowKeys[i-1])
					x += prevHalfW + baseGapX + halfW
				}

				idx := nodeIdx[key]
				nodes[idx].Attributes["x"] = x
				nodes[idx].Attributes["y"] = yCursor + rowHeights[rowIdx]/2.0
			}

			yCursor += rowHeights[rowIdx] + rowGapY
		}

		y += layerH + layerGapY
	}
}

// assignPkgPositions assigns x/y to package-diagram nodes by clustering them
// by package. Packages are arranged in a square grid; nodes within each
// package are placed in a compact 2-column sub-grid.
func assignPkgPositions(nodes []sigmaNode) {
	if len(nodes) == 0 {
		return
	}

	// Group node indices by package path.
	pkgOrder := []string{}
	pkgNodes := make(map[string][]int)
	for i, n := range nodes {
		pkg, _ := n.Attributes["pkg"].(string)
		if _, seen := pkgNodes[pkg]; !seen {
			pkgOrder = append(pkgOrder, pkg)
		}
		pkgNodes[pkg] = append(pkgNodes[pkg], i)
	}
	sort.Strings(pkgOrder)

	cols := int(math.Ceil(math.Sqrt(float64(len(pkgOrder)))))
	const pkgW = 340.0 // horizontal space per package cluster
	const pkgH = 280.0 // vertical space per package cluster
	const nodeH = 80.0 // vertical spacing within a cluster
	const nodeColW = 150.0

	for pi, pkg := range pkgOrder {
		col := pi % cols
		row := pi / cols
		baseX := float64(col) * pkgW
		baseY := float64(row) * pkgH

		for ni, idx := range pkgNodes[pkg] {
			c := ni % 2
			r := ni / 2
			nodes[idx].Attributes["x"] = baseX + float64(c)*nodeColW
			nodes[idx].Attributes["y"] = baseY + float64(r)*nodeH
		}
	}
}
