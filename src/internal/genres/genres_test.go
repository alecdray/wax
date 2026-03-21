package genres

import (
	"testing"
)

func TestNormalizeLabel(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Soul-Jazz", "Soul Jazz"},
		{"Jazz-Rock", "Jazz Rock"},
		{"Rhythm & Blues", "Rhythm and Blues"},
		{"Jazzy Hip-Hop", "Jazzy Hip Hop"},
		{"Rock", "Rock"},
		{"Post Rock", "Post Rock"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := normalizeLabel(tt.in)
			if got != tt.want {
				t.Errorf("normalizeLabel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRoot(t *testing.T) {
	dag, err := Load()
	if err != nil {
		t.Fatalf("failed to load genres: %v", err)
	}

	t.Run("is not nil", func(t *testing.T) {
		if dag.Root == nil {
			t.Fatal("expected root node, got nil")
		}
	})

	t.Run("has no parents", func(t *testing.T) {
		if len(dag.Root.Parents) != 0 {
			t.Errorf("root should have no parents, got %d", len(dag.Root.Parents))
		}
	})

	t.Run("has children", func(t *testing.T) {
		if len(dag.Root.Children) == 0 {
			t.Error("root should have children")
		}
	})
}

func TestSearch(t *testing.T) {
	dag, err := Load()
	if err != nil {
		t.Fatalf("failed to load genres: %v", err)
	}

	t.Run("returns results for a known genre", func(t *testing.T) {
		results := dag.Search("rock")
		if len(results) == 0 {
			t.Fatal("expected results for 'rock', got none")
		}
	})

	t.Run("is case insensitive", func(t *testing.T) {
		lower := dag.Search("rock")
		upper := dag.Search("ROCK")
		if len(lower) != len(upper) {
			t.Errorf("case sensitivity mismatch: %d vs %d results", len(lower), len(upper))
		}
	})

	t.Run("returns empty for nonsense query", func(t *testing.T) {
		results := dag.Search("zzzzzzzzzzzzz")
		if len(results) != 0 {
			t.Errorf("expected no results for nonsense query, got %d", len(results))
		}
	})

	t.Run("hyphenated label matches space-separated query ahead of longer labels", func(t *testing.T) {
		results := dag.Search("Hip Hop")
		if len(results) == 0 {
			t.Skip("Hip-Hop not in dataset")
		}
		if results[0].Label != "hip-hop" {
			t.Errorf("expected top result to be 'hip-hop', got %q", results[0].Label)
		}
	})

	t.Run("ampersand in query matches 'and' label", func(t *testing.T) {
		results := dag.Search("rock & roll")
		if len(results) == 0 {
			t.Skip("rock and roll not in dataset")
		}
		if results[0].Label != "rock and roll" {
			t.Errorf("expected top result to be 'rock and roll', got %q", results[0].Label)
		}
	})

	t.Run("closer matches rank higher", func(t *testing.T) {
		results := dag.Search("punk")
		if len(results) < 2 {
			t.Skip("not enough results to compare ranking")
		}
		// The first result should contain "punk" in its label.
		found := false
		for i, n := range results {
			if n.Label == "punk" {
				if i > 3 {
					t.Errorf("exact match 'punk' ranked too low at position %d", i)
				}
				found = true
				break
			}
		}
		if !found {
			t.Log("no exact 'punk' match in results (may not exist in dataset)")
		}
	})
}
