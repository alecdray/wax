package discogs

import (
	"testing"

	"github.com/alecdray/wax/src/internal/genres"
)

func TestSplitTerm(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"Funk / Soul", []string{"Funk", "Soul"}},
		{"Folk, World, & Country", []string{"Folk", "World", "Country"}},
		{"RnB/Swing", []string{"RnB", "Swing"}},
		{"Rock", []string{"Rock"}},
		{"Jazz-Rock", []string{"Jazz-Rock"}}, // hyphen is not a split char
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := splitTerm(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("splitTerm(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitTerm(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// minimalDAG builds a small DAG with a few well-known genre nodes for testing.
func minimalDAG() *genres.DAG {
	return genres.Build([]genres.Entry{
		{Genre: "Q11399", GenreLabel: "Rock"},
		{Genre: "Q9778", GenreLabel: "Electronic music", Parent: "Q638", ParentLabel: "Music"},
		{Genre: "Q8341", GenreLabel: "Jazz", Parent: "Q638", ParentLabel: "Music"},
		{Genre: "Q842328", GenreLabel: "Soul music", Parent: "Q638", ParentLabel: "Music"},
		{Genre: "Q638", GenreLabel: "Music"},
	})
}

func resolveItem(dag *genres.DAG, item *SearchItem) []*genres.Node {
	if item == nil {
		return nil
	}
	return Resolve(dag, append(item.Genre, item.Style...))
}

func TestResolveItemGenres(t *testing.T) {
	dag := minimalDAG()

	t.Run("returns nodes for matched genres and styles", func(t *testing.T) {
		item := &SearchItem{
			Genre: []string{"Rock"},
			Style: []string{"Jazz"},
		}
		got := resolveItem(dag, item)
		if len(got) != 2 {
			t.Fatalf("expected 2 nodes, got %v", got)
		}
	})

	t.Run("combines genre and style fields", func(t *testing.T) {
		item := &SearchItem{
			Genre: []string{"Rock"},
			Style: []string{"Soul music"},
		}
		got := resolveItem(dag, item)
		if len(got) != 2 {
			t.Fatalf("expected 2 nodes, got %v", got)
		}
	})

	t.Run("deduplicates when genre and style resolve to the same node", func(t *testing.T) {
		item := &SearchItem{
			Genre: []string{"Rock"},
			Style: []string{"Rock"},
		}
		got := resolveItem(dag, item)
		if len(got) != 1 {
			t.Fatalf("expected 1 node after dedup, got %v", got)
		}
		if got[0].Label != "Rock" {
			t.Errorf("expected Rock, got %q", got[0].Label)
		}
	})

	t.Run("skips unmatched terms without error", func(t *testing.T) {
		item := &SearchItem{
			Genre: []string{"xyzzy nonsense genre"},
			Style: []string{"Rock"},
		}
		got := resolveItem(dag, item)
		if len(got) != 1 {
			t.Fatalf("expected 1 node, got %v", got)
		}
		if got[0].Label != "Rock" {
			t.Errorf("expected Rock, got %q", got[0].Label)
		}
	})

	t.Run("returns empty slice when no terms match", func(t *testing.T) {
		item := &SearchItem{
			Genre: []string{"xyzzy1"},
			Style: []string{"xyzzy2"},
		}
		got := resolveItem(dag, item)
		if len(got) != 0 {
			t.Fatalf("expected empty, got %v", got)
		}
	})

	t.Run("returns empty slice for empty item", func(t *testing.T) {
		item := &SearchItem{}
		got := resolveItem(dag, item)
		if len(got) != 0 {
			t.Fatalf("expected empty, got %v", got)
		}
	})
}
