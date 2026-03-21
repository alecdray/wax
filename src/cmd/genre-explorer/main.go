package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecdray/wax/src/internal/genres"
)

var dag *genres.DAG

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	var err error
	dag, err = genres.Load()
	if err != nil {
		slog.Error("Failed to load genres", "error", err)
		os.Exit(1)
	}
	slog.Info("Loaded genre DAG", "nodes", len(dag.Nodes()), "root", dag.Root.Label)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /api/search", handleSearch)
	mux.HandleFunc("GET /api/node/{id}", handleNode)
	mux.HandleFunc("GET /api/graph/{id}", handleGraph)

	addr := "localhost:7331"
	slog.Info("Listening", "url", "http://"+addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

type nodeResponse struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Parents     []idLabel `json:"parents"`
	Children    []idLabel `json:"children"`
	AncestorCount   int  `json:"ancestor_count"`
	DescendantCount int  `json:"descendant_count"`
}

type idLabel struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type graphResponse struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

type graphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Group string `json:"group"` // "focus", "parent", "child"
}

type graphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		w.Header().Set("Content-Type", "text/html")
		return
	}

	results := dag.Search(q)
	if len(results) > 20 {
		results = results[:20]
	}

	w.Header().Set("Content-Type", "text/html")
	for _, n := range results {
		w.Write([]byte(`<li class="result-item" onclick="loadGraph('` + n.ID + `')">` + n.Label + `</li>`))
	}
}

func handleNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n := dag.Get(id)
	if n == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp := nodeResponse{
		ID:              n.ID,
		Label:           n.Label,
		AncestorCount:   len(dag.Ancestors(id)),
		DescendantCount: len(dag.Descendants(id)),
	}
	for _, p := range n.Parents {
		resp.Parents = append(resp.Parents, idLabel{p.ID, p.Label})
	}
	for _, c := range n.Children {
		resp.Children = append(resp.Children, idLabel{c.ID, c.Label})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGraph(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n := dag.Get(id)
	if n == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var nodes []graphNode
	var edges []graphEdge

	nodes = append(nodes, graphNode{n.ID, n.Label, "focus"})

	for _, p := range n.Parents {
		nodes = append(nodes, graphNode{p.ID, p.Label, "parent"})
		edges = append(edges, graphEdge{p.ID, n.ID})
	}
	children := n.Children
	if len(children) > 30 {
		children = children[:30]
	}
	for _, c := range children {
		nodes = append(nodes, graphNode{c.ID, c.Label, "child"})
		edges = append(edges, graphEdge{n.ID, c.ID})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphResponse{nodes, edges})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Genre Explorer</title>
<script src="https://unpkg.com/htmx.org@2.0.4"></script>
<script src="https://unpkg.com/vis-network@9.1.9/standalone/umd/vis-network.min.js"></script>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: system-ui, sans-serif; display: flex; height: 100vh; background: #f5f5f5; }

#sidebar {
  width: 300px;
  min-width: 300px;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-right: 1px solid #ddd;
}
#search-box {
  padding: 12px;
  border-bottom: 1px solid #eee;
}
#search-box input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #ccc;
  border-radius: 6px;
  font-size: 14px;
}
#results {
  list-style: none;
  overflow-y: auto;
  flex: 1;
  font-size: 14px;
}
.result-item {
  cursor: pointer;
  padding: 6px 10px;
  border-bottom: 1px solid #eee;
}
.result-item:hover { background: #f0f4ff; }

#main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}
#graph {
  flex: 1;
  min-height: 0;
  background: #fafafa;
}
#detail {
  height: 200px;
  overflow-y: auto;
  padding: 16px;
  background: #fff;
  border-top: 1px solid #ddd;
  font-size: 13px;
}
#detail h2 { font-size: 16px; margin-bottom: 8px; }
#detail .meta { color: #666; margin-bottom: 8px; }
#detail .section { margin-top: 8px; }
#detail .section h3 { font-size: 13px; font-weight: 600; margin-bottom: 4px; }
#detail .tag {
  display: inline-block;
  background: #e8eef8;
  border-radius: 4px;
  padding: 2px 7px;
  margin: 2px;
  cursor: pointer;
  font-size: 12px;
}
#detail .tag:hover { background: #c8d8f8; }
</style>
</head>
<body>

<div id="sidebar">
  <div id="search-box">
    <input
      type="text"
      placeholder="Search genres..."
      name="q"
      hx-get="/api/search"
      hx-trigger="input changed delay:200ms"
      hx-target="#results"
    />
  </div>
  <ul id="results"></ul>
</div>

<div id="main">
  <div id="graph"></div>
  <div id="detail"><p style="color:#aaa">Select a genre to explore.</p></div>
</div>

<script>
let network = null;

function loadGraph(id) {
  fetch('/api/graph/' + id)
    .then(r => r.json())
    .then(data => {
      const nodeColors = {
        focus:  { background: '#4f6ef7', border: '#2a4ad0', font: { color: '#fff' } },
        parent: { background: '#f0f4ff', border: '#a0b0e0' },
        child:  { background: '#f0fff4', border: '#80c090' },
      };
      const nodes = new vis.DataSet(data.nodes.map(n => ({
        id: n.id,
        label: n.label,
        ...nodeColors[n.group],
        shape: n.group === 'focus' ? 'box' : 'ellipse',
      })));
      const edges = new vis.DataSet(data.edges.map(e => ({
        from: e.from,
        to: e.to,
        arrows: 'to',
        color: { color: '#aaa' },
      })));

      const container = document.getElementById('graph');
      if (network) network.destroy();
      network = new vis.Network(container, { nodes, edges }, {
        physics: {
          solver: 'forceAtlas2Based',
          forceAtlas2Based: { gravitationalConstant: -80, springLength: 120 },
          stabilization: { iterations: 200 },
        },
        edges: { smooth: { type: 'curvedCW', roundness: 0.2 } },
        interaction: { hover: true },
      });
      network.once('stabilizationIterationsDone', () => network.fit());

      network.on('click', params => {
        if (params.nodes.length) selectNode(params.nodes[0]);
      });
    });

  selectNode(id);
}

function selectNode(id) {
  fetch('/api/node/' + id)
    .then(r => r.json())
    .then(n => {
      const parents = (n.parents || []).map(p =>
        '<span class="tag" onclick="loadGraph(\'' + p.id + '\')">' + p.label + '</span>'
      ).join('');
      const children = (n.children || []).map(c =>
        '<span class="tag" onclick="loadGraph(\'' + c.id + '\')">' + c.label + '</span>'
      ).join('');

      document.getElementById('detail').innerHTML =
        '<h2>' + n.label + '</h2>' +
        '<div class="meta">' + n.id + ' &mdash; ' + n.ancestor_count + ' ancestors, ' + n.descendant_count + ' descendants</div>' +
        (parents ? '<div class="section"><h3>Parents</h3>' + parents + '</div>' : '') +
        (children ? '<div class="section"><h3>Children</h3>' + children + '</div>' : '');
    });
}
</script>
</body>
</html>`
