package discogs

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/alecdray/wax/src/internal/genres"
)

var splitRe = regexp.MustCompile(`[/,]+`)

// splitTerm splits a Discogs compound term (e.g. "Funk / Soul", "Folk, World, & Country")
// into individual sub-terms.
func splitTerm(term string) []string {
	parts := splitRe.Split(term, -1)
	var out []string
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), "&")
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// resolveOne tries to find the best DAG node for a single term.
// Returns nil if no match.
func resolveOne(dag *genres.DAG, term string) *genres.Node {
	if m := dag.Search(term); len(m) > 0 {
		return m[0]
	}
	return nil
}

// resolveItemGenres combines the genre and style fields from a Discogs search item,
// fuzzy-matches them against the DAG, and returns the resolved genre labels.
func resolveItemGenres(dag *genres.DAG, item *SearchItem) []string {
	terms := append(item.Genre, item.Style...)
	nodes := Resolve(dag, terms)
	labels := make([]string, 0, len(nodes))
	for _, n := range nodes {
		labels = append(labels, n.Label)
	}
	return labels
}

// Resolve fuzzy-matches a slice of Discogs genre/style strings against the
// genre DAG and returns the best-matching node for each term.
// Compound terms (e.g. "Funk / Soul") are split and each part resolved
// independently. Terms with no match are logged and skipped.
func Resolve(dag *genres.DAG, terms []string) []*genres.Node {
	seen := make(map[string]bool)
	var result []*genres.Node

	add := func(n *genres.Node) {
		if n != nil && !seen[n.ID] {
			seen[n.ID] = true
			result = append(result, n)
		}
	}

	for _, term := range terms {
		n := resolveOne(dag, term)
		if n != nil {
			add(n)
			continue
		}

		// Try splitting into sub-terms for compound Discogs labels.
		parts := splitTerm(term)
		if len(parts) > 1 {
			matched := false
			for _, part := range parts {
				if sn := resolveOne(dag, part); sn != nil {
					add(sn)
					matched = true
				}
			}
			if matched {
				continue
			}
		}

		slog.Warn("no DAG match for discogs term", "term", term)
	}
	return result
}
