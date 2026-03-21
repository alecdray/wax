package main

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/alecdray/wax/src/internal/genres"
)

const (
	sparqlURL   = "https://query.wikidata.org/sparql"
	sparqlQuery = `SELECT ?genre ?genreLabel ?parent ?parentLabel WHERE { ?genre wdt:P31 wd:Q188451 . OPTIONAL { ?genre wdt:P279 ?parent } SERVICE wikibase:label { bd:serviceParam wikibase:language "en" } }`
	outputFile  = "src/internal/genres/data.json"
)

type sparqlBinding struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sparqlResult struct {
	Bindings []map[string]sparqlBinding `json:"bindings"`
}

type sparqlResponse struct {
	Results sparqlResult `json:"results"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	slog.Info("Fetching genres from Wikidata")
	bindings, err := fetch(sparqlQuery)
	if err != nil {
		slog.Error("Request failed", "error", err)
		os.Exit(1)
	}

	entries := bindingsToEntries(bindings)
	entries = append(entries, genres.Entry{Genre: rootGenreID, GenreLabel: "music"})
	slog.Info("Fetched genres", "count", len(entries))

	for round := 1; ; round++ {
		filtered, removed := pruneOrphans(entries)
		if removed == 0 {
			slog.Info("No more orphans", "round", round, "remaining", len(entries))
			break
		}
		slog.Info("Removed orphans", "round", round, "removed", removed, "remaining", len(filtered))
		entries = filtered
	}

	before := len(entries)
	entries = removeStaleParentRefs(entries)
	if removed := before - len(entries); removed > 0 {
		slog.Info("Removed stale parent refs", "removed", removed)
	}

	if orphans := findOrphans(entries); len(orphans) > 0 {
		slog.Info("Genres with no known parents", "count", len(orphans))
		writeJSON("tmp/wikidata_orphans.json", orphans)
	}

	if missing := findMissingParents(entries); len(missing) > 0 {
		slog.Warn("Parents not in dataset", "count", len(missing))
		writeJSON("tmp/wikidata_missing_parents.json", missing)
	}

	if err := os.MkdirAll("tmp", 0755); err != nil {
		slog.Error("Failed to create tmp dir", "error", err)
		os.Exit(1)
	}

	dag := genres.Build(entries)
	if errs := dag.Validate(); len(errs) > 0 {
		for _, e := range errs {
			slog.Error("DAG validation failed", "detail", e)
		}
		os.Exit(1)
	}
	slog.Info("DAG valid", "nodes", len(dag.Nodes()), "roots", len(dag.Roots()))

	if err := writeJSON(outputFile, entries); err != nil {
		slog.Error("Failed to write output", "error", err)
		os.Exit(1)
	}

	slog.Info("Done", "output", outputFile)
}

func writeJSON(file string, v any) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
