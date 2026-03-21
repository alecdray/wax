package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/alecdray/wax/src/internal/genres"
)

const rootGenreID = "Q115484611"

// fetch fetches raw SPARQL bindings from Wikidata.
func fetch(query string) ([]map[string]sparqlBinding, error) {
	req, err := http.NewRequest(http.MethodGet, sparqlURL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("format", "json")
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/sparql-results+json")
	req.Header.Set("User-Agent", "wax-genre-puller/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	return parseBindings(resp.Body)
}

// parseBindings decodes SPARQL JSON bindings from r.
func parseBindings(r io.Reader) ([]map[string]sparqlBinding, error) {
	var sr sparqlResponse
	if err := json.NewDecoder(r).Decode(&sr); err != nil {
		return nil, err
	}
	return sr.Results.Bindings, nil
}

// bindingsToEntries converts raw SPARQL bindings to genre entries.
func bindingsToEntries(bindings []map[string]sparqlBinding) []genres.Entry {
	entries := make([]genres.Entry, 0, len(bindings))
	for _, b := range bindings {
		entries = append(entries, genres.Entry{
			Genre:       path.Base(b["genre"].Value),
			GenreLabel:  strings.TrimSuffix(b["genreLabel"].Value, " music"),
			Parent:      path.Base(b["parent"].Value),
			ParentLabel: strings.TrimSuffix(b["parentLabel"].Value, " music"),
		})
	}
	return entries
}

// pruneOrphans removes entries with no known parent, preserving the root.
// Returns the pruned slice and the number of entries removed.
// A single call removes one generation; call repeatedly until 0 is returned.
func pruneOrphans(entries []genres.Entry) ([]genres.Entry, int) {
	known := make(map[string]bool, len(entries))
	for _, e := range entries {
		known[e.Genre] = true
	}

	parents := make(map[string][]string, len(entries))
	for _, e := range entries {
		if e.Parent != "" && e.Parent != "." {
			parents[e.Genre] = append(parents[e.Genre], e.Parent)
		}
	}

	toRemove := make(map[string]bool)
	for _, e := range entries {
		if e.Genre == rootGenreID {
			continue
		}
		hasKnown := false
		for _, p := range parents[e.Genre] {
			if known[p] {
				hasKnown = true
				break
			}
		}
		if !hasKnown {
			toRemove[e.Genre] = true
		}
	}

	if len(toRemove) == 0 {
		return entries, 0
	}

	filtered := make([]genres.Entry, 0, len(entries)-len(toRemove))
	for _, e := range entries {
		if !toRemove[e.Genre] {
			filtered = append(filtered, e)
		}
	}
	return filtered, len(toRemove)
}

// removeStaleParentRefs removes entry rows whose parent is not in the final set.
func removeStaleParentRefs(entries []genres.Entry) []genres.Entry {
	known := make(map[string]bool, len(entries))
	for _, e := range entries {
		known[e.Genre] = true
	}
	filtered := make([]genres.Entry, 0, len(entries))
	for _, e := range entries {
		if e.Parent != "" && e.Parent != "." && !known[e.Parent] {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

type missingParent struct {
	ParentID     string `json:"parent_id"`
	ParentLabel  string `json:"parent_label"`
	ExampleGenre string `json:"example_genre"`
	ExampleLabel string `json:"example_label"`
}

// findMissingParents returns parent IDs referenced by genres that have no
// known parent in the dataset.
func findMissingParents(entries []genres.Entry) []missingParent {
	known := make(map[string]bool, len(entries))
	for _, e := range entries {
		known[e.Genre] = true
	}

	type pe struct{ id, label string }
	genreParents := make(map[string][]pe)
	genreLabel := make(map[string]string)
	for _, e := range entries {
		genreLabel[e.Genre] = e.GenreLabel
		if e.Parent != "" && e.Parent != "." {
			genreParents[e.Genre] = append(genreParents[e.Genre], pe{e.Parent, e.ParentLabel})
		}
	}

	seen := make(map[string]bool)
	var missing []missingParent
	for genre, parents := range genreParents {
		for _, p := range parents {
			if known[p.id] {
				goto next
			}
		}
		for _, p := range parents {
			if !seen[p.id] {
				seen[p.id] = true
				missing = append(missing, missingParent{
					ParentID:     p.id,
					ParentLabel:  p.label,
					ExampleGenre: genre,
					ExampleLabel: genreLabel[genre],
				})
			}
		}
	next:
	}
	return missing
}

type orphan struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// findOrphans returns genres with no parent present in the dataset.
func findOrphans(entries []genres.Entry) []orphan {
	known := make(map[string]bool, len(entries))
	for _, e := range entries {
		known[e.Genre] = true
	}

	type pe struct{ id, label string }
	genreParents := make(map[string][]pe)
	for _, e := range entries {
		if e.Parent != "" && e.Parent != "." {
			genreParents[e.Genre] = append(genreParents[e.Genre], pe{e.Parent, e.ParentLabel})
		}
	}

	seen := make(map[string]bool)
	var orphans []orphan
	for _, e := range entries {
		if seen[e.Genre] || e.Genre == rootGenreID {
			continue
		}
		for _, p := range genreParents[e.Genre] {
			if known[p.id] {
				goto next
			}
		}
		seen[e.Genre] = true
		orphans = append(orphans, orphan{e.Genre, e.GenreLabel})
	next:
	}
	return orphans
}
