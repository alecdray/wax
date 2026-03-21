package genres

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

//go:embed data.json
var dataJSON []byte

// Entry is the raw record from the Wikidata JSON.
type Entry struct {
	Genre       string `json:"genre"`
	GenreLabel  string `json:"genre_label"`
	Parent      string `json:"parent,omitempty"`
	ParentLabel string `json:"parent_label,omitempty"`
}

// Node is a genre vertex in the DAG.
type Node struct {
	ID       string
	Label    string
	Parents  []*Node
	Children []*Node
}

// DAG is a directed acyclic graph of music genres.
type DAG struct {
	nodes map[string]*Node
	Root  *Node
}

// Load builds a DAG from the embedded data.json.
func Load() (*DAG, error) {
	var entries []Entry
	if err := json.Unmarshal(dataJSON, &entries); err != nil {
		return nil, err
	}
	return Build(entries), nil
}

// Build constructs a DAG from a slice of Wikidata entries.
func Build(entries []Entry) *DAG {
	d := &DAG{nodes: make(map[string]*Node)}

	for _, e := range entries {
		if _, ok := d.nodes[e.Genre]; !ok {
			d.nodes[e.Genre] = &Node{ID: e.Genre, Label: e.GenreLabel}
		} else if d.nodes[e.Genre].Label == "" {
			d.nodes[e.Genre].Label = e.GenreLabel
		}

		if e.Parent != "" && e.Parent != "." {
			if _, ok := d.nodes[e.Parent]; !ok {
				d.nodes[e.Parent] = &Node{ID: e.Parent, Label: e.ParentLabel}
			} else if d.nodes[e.Parent].Label == "" {
				d.nodes[e.Parent].Label = e.ParentLabel
			}
		}
	}

	for _, e := range entries {
		if e.Parent == "" || e.Parent == "." {
			continue
		}
		child := d.nodes[e.Genre]
		parent := d.nodes[e.Parent]

		if !hasNode(parent.Children, child) {
			parent.Children = append(parent.Children, child)
		}
		if !hasNode(child.Parents, parent) {
			child.Parents = append(child.Parents, parent)
		}
	}

	for _, n := range d.nodes {
		if len(n.Parents) == 0 {
			d.Root = n
			break
		}
	}

	return d
}

// Validate checks the DAG for structural issues: cycles and multiple roots.
// Returns a list of error strings; empty means valid.
func (d *DAG) Validate() []string {
	var errs []string

	if cycles := d.findCycles(); len(cycles) > 0 {
		for _, n := range cycles {
			errs = append(errs, fmt.Sprintf("cycle detected involving %s (%s)", n.ID, n.Label))
		}
	}

	if roots := d.Roots(); len(roots) != 1 {
		errs = append(errs, fmt.Sprintf("expected exactly 1 root, got %d", len(roots)))
	}

	return errs
}

// findCycles returns nodes involved in a cycle using DFS coloring.
func (d *DAG) findCycles() []*Node {
	const (
		white = 0
		grey  = 1
		black = 2
	)
	color := make(map[string]int, len(d.nodes))
	var cyclic []*Node

	var dfs func(n *Node)
	dfs = func(n *Node) {
		color[n.ID] = grey
		for _, child := range n.Children {
			switch color[child.ID] {
			case grey:
				cyclic = append(cyclic, child)
			case white:
				dfs(child)
			}
		}
		color[n.ID] = black
	}

	for _, n := range d.nodes {
		if color[n.ID] == white {
			dfs(n)
		}
	}
	return cyclic
}

// Search returns nodes whose labels fuzzy-match the query, ranked by closeness.
func (d *DAG) Search(query string) []*Node {
	labels := make([]string, 0, len(d.nodes))
	byLabel := make(map[string]*Node, len(d.nodes))
	for _, n := range d.nodes {
		labels = append(labels, n.Label)
		byLabel[n.Label] = n
	}
	matches := fuzzy.RankFindFold(query, labels)
	sort.Sort(matches)
	result := make([]*Node, 0, len(matches))
	for _, m := range matches {
		result = append(result, byLabel[m.Target])
	}
	return result
}

// Get returns the node for the given Wikidata ID, or nil if not found.
func (d *DAG) Get(id string) *Node {
	return d.nodes[id]
}

// Nodes returns all nodes in the DAG.
func (d *DAG) Nodes() map[string]*Node {
	return d.nodes
}

// Roots returns all nodes that have no parents.
func (d *DAG) Roots() []*Node {
	var roots []*Node
	for _, n := range d.nodes {
		if len(n.Parents) == 0 {
			roots = append(roots, n)
		}
	}
	return roots
}

// Ancestors returns all ancestor nodes of the given ID (breadth-first, no duplicates).
func (d *DAG) Ancestors(id string) []*Node {
	return walk(d.nodes[id], func(n *Node) []*Node { return n.Parents })
}

// Descendants returns all descendant nodes of the given ID (breadth-first, no duplicates).
func (d *DAG) Descendants(id string) []*Node {
	return walk(d.nodes[id], func(n *Node) []*Node { return n.Children })
}


func walk(start *Node, next func(*Node) []*Node) []*Node {
	if start == nil {
		return nil
	}
	seen := map[string]bool{start.ID: true}
	queue := []*Node{start}
	var result []*Node
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, n := range next(cur) {
			if !seen[n.ID] {
				seen[n.ID] = true
				result = append(result, n)
				queue = append(queue, n)
			}
		}
	}
	return result
}

func hasNode(slice []*Node, n *Node) bool {
	for _, x := range slice {
		if x.ID == n.ID {
			return true
		}
	}
	return false
}
