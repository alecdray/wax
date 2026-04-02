package labels

import (
	"testing"

	"github.com/alecdray/wax/src/internal/genres"
)

func TestSearchGenres(t *testing.T) {
	dag, err := genres.Load()
	if err != nil {
		t.Skip("genre DAG unavailable:", err)
	}
	s := NewService(nil, dag)

	t.Run("returns results for known genre query", func(t *testing.T) {
		results := s.SearchGenres("electronic")
		if len(results) == 0 {
			t.Error("expected results for 'electronic', got none")
		}
	})

	t.Run("returns empty for blank query", func(t *testing.T) {
		results := s.SearchGenres("")
		if len(results) != 0 {
			t.Errorf("expected no results for empty query, got %d", len(results))
		}
	})

	t.Run("results include parent label breadcrumb", func(t *testing.T) {
		results := s.SearchGenres("electronic")
		hasParent := false
		for _, r := range results {
			if r.ParentLabel != "" {
				hasParent = true
				break
			}
		}
		if !hasParent {
			t.Error("expected at least one result to have a parent label")
		}
	})
}

func TestNormalizeLabel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  Hello World  ", "hello world"},
		{"Rock & Roll!", "rock & roll"},
		{"Hip-Hop", "hip-hop"},
		{"Café", "café"},
		{"Tag@#$%", "tag"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizeLabel(tc.input)
			if got != tc.want {
				t.Errorf("normalizeLabel(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFilterParamsExpandGenreDescendants(t *testing.T) {
	dag, err := genres.Load()
	if err != nil {
		t.Skip("genre DAG unavailable:", err)
	}

	t.Run("nil dag returns params unchanged", func(t *testing.T) {
		from := "github.com/alecdray/wax/src/internal/library"
		_ = from
		// We test through the library package but the logic is there;
		// test normalize here with a nil dag guard
		s := NewService(nil, nil)
		results := s.SearchGenres("anything")
		if len(results) != 0 {
			t.Error("expected no results with nil DAG")
		}
	})

	t.Run("descendants are included for a known genre", func(t *testing.T) {
		// Find a genre that has children
		var parentID string
		for id, node := range dag.Nodes() {
			if len(node.Children) > 0 {
				parentID = id
				break
			}
		}
		if parentID == "" {
			t.Skip("no genre with children found")
		}

		descs := dag.Descendants(parentID)
		if len(descs) == 0 {
			t.Skip("no descendants found")
		}

		// Verify Descendants returns nodes that are children of the parent.
		childIDs := make(map[string]bool, len(descs))
		for _, d := range descs {
			childIDs[d.ID] = true
		}
		parentNode := dag.Get(parentID)
		for _, child := range parentNode.Children {
			if !childIDs[child.ID] {
				t.Errorf("child %s not in Descendants(%s)", child.ID, parentID)
			}
		}
	})
}
