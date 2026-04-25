package main

import (
	"fmt"
	"strings"
)

// ── Color palette (mirrors Mermaid classDef palette) ─────────────────────────

const (
	sgColorLeaf      = "#27ae60"
	sgColorProvider  = "#2980b9"
	sgColorInvoke    = "#8e44ad"
	sgColorCommand   = "#4a6fa5"
	sgColorIface     = "#e67e22"
	sgColorDecorator = "#c0392b"
	sgColorEdgeDep   = "#6c757d" // dependency arrow
	sgColorEdgeImpl  = "#e67e22" // implements arrow (orange)
	sgColorEdgeSlice = "#5dade2" // slice/auto-collect arrow (blue)
)

// ── Graphology-compatible JSON types ─────────────────────────────────────────

type sigmaNode struct {
	Key        string         `json:"key"`
	Attributes map[string]any `json:"attributes"`
}

type sigmaEdge struct {
	Key        string         `json:"key"`
	Source     string         `json:"source"`
	Target     string         `json:"target"`
	Attributes map[string]any `json:"attributes"`
}

type sigmaGraph struct {
	Nodes []sigmaNode `json:"nodes"`
	Edges []sigmaEdge `json:"edges"`
}

// ── HTML Assembly ─────────────────────────────────────────────────────────────

// buildSigmaHTML assembles the self-contained HTML string.
// graphJSON must be valid JSON; tooltipFn is a raw JS function body.
func buildSigmaHTML(title, legend, graphJSON, tooltipFn string) []byte {
	html := strings.NewReplacer(
		"{{TITLE}}", title,
		"{{LEGEND}}", legend,
		"{{GRAPH_JSON}}", graphJSON,
		"{{TOOLTIP_FN}}", tooltipFn,
	).Replace(sigmaHTMLTemplate)
	return []byte(html)
}

// ── Legend HTML ───────────────────────────────────────────────────────────────

func diLegend() string {
	items := []struct{ color, label string }{
		{sgColorLeaf, "Leaf provider"},
		{sgColorProvider, "Provider"},
		{sgColorInvoke, "Invoke"},
		{sgColorDecorator, "Decorator"},
		{sgColorIface, "Interface"},
		{sgColorCommand, "Command"},
	}
	return legendHTML(items)
}

func pkgLegend() string {
	items := []struct{ color, label string }{
		{sgColorProvider, "Struct"},
		{sgColorIface, "Interface"},
		{sgColorDecorator, "Decorator"},
		{sgColorEdgeImpl, "→ implements"},
		{sgColorEdgeDep, "→ depends"},
	}
	return legendHTML(items)
}

func legendHTML(items []struct{ color, label string }) string {
	var sb strings.Builder
	for _, it := range items {
		fmt.Fprintf(&sb,
			`<div class="li"><span class="dot" style="background:%s"></span>%s</div>`,
			it.color, it.label)
	}
	return sb.String()
}

// ── Tooltip JS functions ──────────────────────────────────────────────────────

func diTooltipFn() string {
	return `function buildTooltip(node, a) {
  const typeLabel = {leaf:'Leaf',provider:'Provider',invoke:'Invoke',decorator:'Decorator',iface:'Interface',command:'Command'};
  let h = '<div class="tt-title">' + a.label.replace(/\n/g,'<br>') + '</div>';
  if (a.pkg) h += '<div class="tt-sub">' + a.pkg + '</div>';
  h += '<span class="tt-badge" style="background:' + a.color + '">' + (typeLabel[a.nodeType]||a.nodeType) + '</span>';
  if (a.depCount > 0)   h += '<div class="tt-stat">→ used by <b>' + a.depCount + '</b></div>';
  if (a.implCount > 0 || a.useCount > 0)
    h += '<div class="tt-stat">' + a.implCount + ' impl · ' + a.useCount + ' use</div>';
  return h;
}`
}

func pkgTooltipFn() string {
	return `function buildTooltip(node, a) {
  const typeLabel = {struct:'Struct',interface:'Interface',decorator:'Decorator'};
  let h = '<div class="tt-title">' + a.label + '</div>';
  if (a.pkg) h += '<div class="tt-sub">' + a.pkg + '</div>';
  h += '<span class="tt-badge" style="background:' + a.color + '">' + (typeLabel[a.nodeType]||a.nodeType) + '</span>';
  if (a.nodeType === 'interface' && (a.implCount > 0 || a.useCount > 0))
    h += '<div class="tt-stat">' + a.implCount + ' impl · ' + a.useCount + ' use</div>';
  else if (a.usedBy > 0)
    h += '<div class="tt-stat">used by <b>' + a.usedBy + '</b></div>';
  if (a.methods && a.methods.length > 0) {
    h += '<div class="tt-divider"></div>';
    a.methods.forEach(m => { h += '<div class="tt-method">+' + m + '</div>'; });
  }
  return h;
}`
}

// ── HTML Template ─────────────────────────────────────────────────────────────

const sigmaHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{TITLE}}</title>
<script src="https://cdn.jsdelivr.net/npm/graphology@0.25.4/dist/graphology.umd.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/sigma@2.4.0/build/sigma.min.js"></script>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:system-ui,-apple-system,sans-serif;background:#0f1117;color:#ddd;overflow:hidden}
#toolbar{position:fixed;top:0;left:0;right:0;height:48px;background:#1a1d27;border-bottom:1px solid #2a2d3a;z-index:10;display:flex;align-items:center;padding:0 16px;gap:10px}
#toolbar h1{font-size:14px;font-weight:600;white-space:nowrap}
#search{flex:1;max-width:260px;padding:5px 10px;background:#252836;border:1px solid #3a3d4a;border-radius:6px;color:#ddd;font-size:13px;outline:none}
#search::placeholder{color:#666}
#node-count{font-size:11px;color:#666;white-space:nowrap}
#hint{font-size:11px;color:#555;white-space:nowrap}
#sigma-container{position:absolute;top:48px;left:0;right:0;bottom:0}
#tooltip{position:fixed;background:#1c1f2eee;border:1px solid #3a3d4a;border-radius:8px;padding:10px 13px;font-size:12px;line-height:1.5;max-width:260px;z-index:20;pointer-events:none;display:none;box-shadow:0 4px 12px #0006}
.tt-title{font-weight:600;font-size:13px;margin-bottom:3px}
.tt-sub{color:#888;font-size:11px;margin-bottom:4px}
.tt-badge{display:inline-block;padding:1px 7px;border-radius:10px;font-size:10px;font-weight:600;color:#fff;margin-bottom:5px}
.tt-stat{color:#aaa;margin-top:2px}
.tt-divider{border-top:1px solid #2a2d3a;margin:6px 0}
.tt-method{color:#8ab;font-size:11px}
#legend{position:fixed;bottom:16px;right:16px;background:#1a1d27cc;border:1px solid #2a2d3a;border-radius:8px;padding:10px 13px;z-index:10;font-size:12px;backdrop-filter:blur(4px)}
.li{display:flex;align-items:center;gap:7px;line-height:2}
.dot{width:11px;height:11px;border-radius:50%;flex-shrink:0}
</style>
</head>
<body>
<div id="toolbar">
  <h1>{{TITLE}}</h1>
  <input id="search" type="search" placeholder="Search nodes…">
  <span id="node-count"></span>
  <span id="hint">click=node chain · bg-click=reset · search=highlight</span>
</div>
<div id="sigma-container"></div>
<div id="tooltip"></div>
<div id="legend">{{LEGEND}}</div>
<script>
(function(){
'use strict';

const DATA = {{GRAPH_JSON}};

{{TOOLTIP_FN}}

// ── Build graphology graph ────────────────────────────────────────────────────
const graph = new graphology.Graph({multi: true, type: 'directed'});
DATA.nodes.forEach(n => graph.addNode(n.key, n.attributes));
DATA.edges.forEach(e => {
  try { graph.addEdge(e.source, e.target, e.attributes); } catch(_) {}
});

// Node x/y positions are pre-computed in Go (hierarchical for DI, package-clustered
// for the package diagram) and embedded in DATA.nodes[*].attributes.

// ── Sigma renderer ────────────────────────────────────────────────────────────
const container = document.getElementById('sigma-container');
const renderer = new Sigma(graph, container, {
  defaultEdgeType: 'arrow',
  renderEdgeLabels: false,
  labelFont: 'system-ui, sans-serif',
  labelSize: 12,
  labelWeight: '500',
  labelColor: {color: '#c8c8c8'},
  edgeLabelFont: 'system-ui, sans-serif',
  edgeLabelSize: 9,
	edgeLabelColor: {color: '#666'},
	minCameraRatio: 0.02,
	maxCameraRatio: 200,
	labelThreshold: 1.2,
});

// ── Full graph interaction (no filtering) ─────────────────────────────────────
const origNodeColors = {};
const origNodeLabels = {};
const origEdgeColors = {};
const origEdgeSizes  = {};

graph.forEachNode((n, a) => {
  origNodeColors[n] = a.color;
  origNodeLabels[n] = a.label;
  graph.setNodeAttribute(n, 'hidden', false);
});
graph.forEachEdge((e, a) => {
  origEdgeColors[e] = a.color;
  origEdgeSizes[e]  = a.size || 1.5;
  graph.setEdgeAttribute(e, 'hidden', false);
});

document.getElementById('node-count').textContent =
  DATA.nodes.length + ' nodes · ' + DATA.edges.length + ' edges';

const DIM_NODE_COLOR = '#1f2433';
const DIM_EDGE_COLOR = '#2a3042';

let activeNode = null;
let activeQuery = '';

function resetGraphStyles() {
  graph.forEachNode(n => {
    graph.setNodeAttribute(n, 'color', origNodeColors[n]);
    graph.setNodeAttribute(n, 'label', origNodeLabels[n]);
  });
  graph.forEachEdge(e => {
    graph.setEdgeAttribute(e, 'color', origEdgeColors[e]);
    graph.setEdgeAttribute(e, 'size', origEdgeSizes[e]);
  });
}

function collectDirectedChain(seed) {
  const nodes = new Set([seed]);
  const edgePairs = new Set();

  // Upstream dependencies (dep -> consumer)
  const upQueue = [seed];
  for (let i = 0; i < upQueue.length; i++) {
    const cur = upQueue[i];
    graph.inNeighbors(cur).forEach(dep => {
      edgePairs.add(dep + '|' + cur);
      if (!nodes.has(dep)) {
        nodes.add(dep);
        upQueue.push(dep);
      }
    });
  }

  // Downstream consumers (producer -> consumer)
  const downQueue = [seed];
  for (let i = 0; i < downQueue.length; i++) {
    const cur = downQueue[i];
    graph.outNeighbors(cur).forEach(cons => {
      edgePairs.add(cur + '|' + cons);
      if (!nodes.has(cons)) {
        nodes.add(cons);
        downQueue.push(cons);
      }
    });
  }

  return {nodes, edgePairs};
}

function applyHighlight(nodes, edgePairs) {
  graph.forEachNode(n => {
    const on = nodes.has(n);
    graph.setNodeAttribute(n, 'color', on ? origNodeColors[n] : DIM_NODE_COLOR);
    graph.setNodeAttribute(n, 'label', on ? origNodeLabels[n] : '');
  });

  graph.forEachEdge((e, _a, src, tgt) => {
    const on = edgePairs.has(src + '|' + tgt);
    graph.setEdgeAttribute(e, 'color', on ? origEdgeColors[e] : DIM_EDGE_COLOR);
    graph.setEdgeAttribute(e, 'size', on ? Math.max(2.2, (origEdgeSizes[e] || 1.5) * 1.5) : 0.8);
  });
}

function applySearch(query) {
  if (!query) {
    resetGraphStyles();
    return;
  }

  const hits = new Set();
  graph.forEachNode((n, a) => {
    if (String(origNodeLabels[n] || '').toLowerCase().includes(query) ||
        (a.pkg && a.pkg.toLowerCase().includes(query))) {
      hits.add(n);
    }
  });

  if (hits.size === 0) {
    resetGraphStyles();
    return;
  }

  // Search mode: highlight matching nodes only.
  applyHighlight(hits, new Set());
}

function renderState() {
  if (activeNode && graph.hasNode(activeNode)) {
    const chain = collectDirectedChain(activeNode);
    applyHighlight(chain.nodes, chain.edgePairs);
  } else {
    applySearch(activeQuery);
  }
  renderer.refresh();
}

// ── Tooltip ───────────────────────────────────────────────────────────────────
const tooltip = document.getElementById('tooltip');
let mouseX = 0, mouseY = 0;
document.addEventListener('mousemove', evt => {
  mouseX = evt.clientX; mouseY = evt.clientY;
  if (tooltip.style.display !== 'none') {
    tooltip.style.left = (mouseX + 16) + 'px';
    tooltip.style.top  = Math.min(mouseY + 8, window.innerHeight - 20) + 'px';
  }
});
renderer.on('enterNode', ({node}) => {
  const attrs = graph.getNodeAttributes(node);
  tooltip.innerHTML = buildTooltip(node, attrs);
  tooltip.style.left = (mouseX + 16) + 'px';
  tooltip.style.top  = (mouseY + 8) + 'px';
  tooltip.style.display = 'block';
});
renderer.on('leaveNode', () => { tooltip.style.display = 'none'; });

// ── Click: highlight full dependency/call chain ────────────────────────────────
renderer.on('clickNode', ({node}) => {
  activeNode = node;
  renderState();
});
renderer.on('clickStage', () => {
  activeNode = null;
  renderState();
});

// ── Search: highlight matching nodes ───────────────────────────────────────────
document.getElementById('search').addEventListener('input', evt => {
  activeNode = null;
  activeQuery = evt.target.value.toLowerCase().trim();
  renderState();
});

// Initial render.
renderState();

})();
</script>
</body>
</html>
`
