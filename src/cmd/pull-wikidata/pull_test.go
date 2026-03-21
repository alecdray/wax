package main

import (
	"strings"
	"testing"

	"github.com/alecdray/wax/src/internal/genres"
)

func TestParseBindings(t *testing.T) {
	body := `{"results":{"bindings":[
		{"genre":{"value":"http://www.wikidata.org/entity/Q11399"},"genreLabel":{"value":"rock music"},"parent":{"value":"http://www.wikidata.org/entity/Q9734"},"parentLabel":{"value":"pop music"}},
		{"genre":{"value":"http://www.wikidata.org/entity/Q43343"},"genreLabel":{"value":"folk"}}
	]}}`

	t.Run("returns all bindings", func(t *testing.T) {
		bindings, err := parseBindings(strings.NewReader(body))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bindings) != 2 {
			t.Fatalf("expected 2 bindings, got %d", len(bindings))
		}
	})

	t.Run("preserves raw values", func(t *testing.T) {
		bindings, _ := parseBindings(strings.NewReader(body))
		if bindings[0]["genre"].Value != "http://www.wikidata.org/entity/Q11399" {
			t.Errorf("unexpected genre value: %s", bindings[0]["genre"].Value)
		}
	})
}

func TestBindingsToEntries(t *testing.T) {
	body := `{"results":{"bindings":[
		{"genre":{"value":"http://www.wikidata.org/entity/Q11399"},"genreLabel":{"value":"rock music"},"parent":{"value":"http://www.wikidata.org/entity/Q9734"},"parentLabel":{"value":"pop music"}},
		{"genre":{"value":"http://www.wikidata.org/entity/Q43343"},"genreLabel":{"value":"folk"}}
	]}}`
	bindings, _ := parseBindings(strings.NewReader(body))

	t.Run("extracts ID from URI", func(t *testing.T) {
		entries := bindingsToEntries(bindings)
		if entries[0].Genre != "Q11399" {
			t.Errorf("expected Q11399, got %s", entries[0].Genre)
		}
		if entries[0].Parent != "Q9734" {
			t.Errorf("expected Q9734, got %s", entries[0].Parent)
		}
	})

	t.Run("strips ' music' suffix from labels", func(t *testing.T) {
		entries := bindingsToEntries(bindings)
		if entries[0].GenreLabel != "rock" {
			t.Errorf("expected 'rock', got %s", entries[0].GenreLabel)
		}
		if entries[0].ParentLabel != "pop" {
			t.Errorf("expected 'pop', got %s", entries[0].ParentLabel)
		}
	})

	t.Run("empty parent becomes '.'", func(t *testing.T) {
		entries := bindingsToEntries(bindings)
		if entries[1].Parent != "." {
			t.Errorf("expected '.', got %q", entries[1].Parent)
		}
	})
}

func TestPruneOrphans(t *testing.T) {
	t.Run("removes genres with no known parent", func(t *testing.T) {
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
			{Genre: "Q1", GenreLabel: "rock", Parent: rootGenreID},
			{Genre: "Q2", GenreLabel: "indie rock", Parent: "Q1"},
			{Genre: "Q3", GenreLabel: "orphan", Parent: "Qunknown"},
		}
		filtered, removed := pruneOrphans(entries)
		if removed != 1 {
			t.Errorf("expected 1 removed, got %d", removed)
		}
		for _, e := range filtered {
			if e.Genre == "Q3" {
				t.Error("orphan Q3 should have been removed")
			}
		}
	})

	t.Run("cascades across multiple rounds", func(t *testing.T) {
		// Q2 depends on Q1 which is an orphan — both should be removed after two rounds.
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
			{Genre: "Q1", GenreLabel: "orphan", Parent: "Qunknown"},
			{Genre: "Q2", GenreLabel: "child of orphan", Parent: "Q1"},
		}
		var totalRemoved int
		for {
			var n int
			entries, n = pruneOrphans(entries)
			if n == 0 {
				break
			}
			totalRemoved += n
		}
		if totalRemoved != 2 {
			t.Errorf("expected 2 total removed, got %d", totalRemoved)
		}
		if len(entries) != 1 || entries[0].Genre != rootGenreID {
			t.Errorf("expected only root to remain, got %v", entries)
		}
	})

	t.Run("never removes the root node", func(t *testing.T) {
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
		}
		filtered, removed := pruneOrphans(entries)
		if removed != 0 {
			t.Errorf("root should never be removed, got removed=%d", removed)
		}
		if len(filtered) != 1 {
			t.Error("expected root to remain")
		}
	})
}

func TestFindMissingParents(t *testing.T) {
	t.Run("reports parent not in dataset", func(t *testing.T) {
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
			{Genre: "Q1", GenreLabel: "rock", Parent: rootGenreID},
			{Genre: "Q2", GenreLabel: "latin", Parent: "Qmissing", ParentLabel: "missing genre"},
		}
		missing := findMissingParents(entries)
		if len(missing) != 1 {
			t.Fatalf("expected 1 missing parent, got %d", len(missing))
		}
		if missing[0].ParentID != "Qmissing" {
			t.Errorf("expected Qmissing, got %s", missing[0].ParentID)
		}
		if missing[0].ExampleGenre != "Q2" {
			t.Errorf("expected Q2 as example, got %s", missing[0].ExampleGenre)
		}
	})

	t.Run("skips genre if at least one parent is known", func(t *testing.T) {
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
			{Genre: "Q1", GenreLabel: "rock", Parent: rootGenreID},
			{Genre: "Q2", GenreLabel: "genre", Parent: "Q1"},
			{Genre: "Q2", GenreLabel: "genre", Parent: "Qmissing"},
		}
		missing := findMissingParents(entries)
		if len(missing) != 0 {
			t.Errorf("expected no missing parents, got %d", len(missing))
		}
	})
}

func TestRemoveStaleParentRefs(t *testing.T) {
	t.Run("removes rows whose parent is not in the set", func(t *testing.T) {
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
			{Genre: "Q1", GenreLabel: "rock", Parent: rootGenreID},
			{Genre: "Q1", GenreLabel: "rock", Parent: "Qgone"},
		}
		result := removeStaleParentRefs(entries)
		if len(result) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(result))
		}
		for _, e := range result {
			if e.Parent == "Qgone" {
				t.Error("stale parent ref should have been removed")
			}
		}
	})

	t.Run("keeps rows with no parent", func(t *testing.T) {
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
		}
		result := removeStaleParentRefs(entries)
		if len(result) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(result))
		}
	})
}

func TestFindOrphans(t *testing.T) {
	t.Run("returns genres with no known parent", func(t *testing.T) {
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
			{Genre: "Q1", GenreLabel: "rock", Parent: rootGenreID},
			{Genre: "Q2", GenreLabel: "latin", Parent: "Qunknown"},
		}
		orphans := findOrphans(entries)
		if len(orphans) != 1 {
			t.Fatalf("expected 1 orphan, got %d", len(orphans))
		}
		if orphans[0].ID != "Q2" {
			t.Errorf("expected Q2, got %s", orphans[0].ID)
		}
	})

	t.Run("does not count root as orphan", func(t *testing.T) {
		entries := []genres.Entry{
			{Genre: rootGenreID, GenreLabel: "music"},
		}
		orphans := findOrphans(entries)
		if len(orphans) != 0 {
			t.Errorf("expected no orphans, got %d", len(orphans))
		}
	})
}
