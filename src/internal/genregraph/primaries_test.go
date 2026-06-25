package genregraph

import (
	"testing"
)

// Real Wikidata Q-ids present in data.json, used as fixtures.
const (
	qHyperpop   = "Q104695865"
	qDeathMetal = "Q483251"
	qBebop      = "Q105513"
	qAmbient    = "Q193207"
	qGFunk      = "Q1045541"
	qPostPunk   = "Q598929"
	qDisco      = "Q58339"
)

func labels(ps []Primary) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Label
	}
	return out
}

func equalLabels(got []Primary, want ...string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i].Label != want[i] {
			return false
		}
	}
	return true
}

func TestPrimaries(t *testing.T) {
	dag, err := Load()
	if err != nil {
		t.Fatalf("failed to load genres: %v", err)
	}

	t.Run("full set is non-empty and every primary exists in the DAG", func(t *testing.T) {
		ps := dag.Primaries()
		if len(ps) == 0 {
			t.Fatal("expected a curated primary set, got none")
		}
		for _, p := range ps {
			if dag.Get(p.ID) == nil {
				t.Errorf("primary %s (%s) is not a node in the DAG", p.Label, p.ID)
			}
		}
	})

	t.Run("cross-branch sub-genre keeps one primary per branch", func(t *testing.T) {
		// hyperpop descends from both pop and electronic.
		got := dag.PrimariesOf(qHyperpop)
		if !equalLabels(got, "pop", "electronic") {
			t.Errorf("hyperpop primaries = %v, want [pop electronic]", labels(got))
		}
	})

	t.Run("nested sub-genre keeps only the most-specific primary", func(t *testing.T) {
		// death metal descends from metal, which descends from rock; metal wins.
		got := dag.PrimariesOf(qDeathMetal)
		if !equalLabels(got, "metal") {
			t.Errorf("death metal primaries = %v, want [metal] (not rock)", labels(got))
		}
	})

	t.Run("punk sub-genre maps to punk, not rock", func(t *testing.T) {
		// post-punk descends from punk, which descends from rock; punk wins.
		got := dag.PrimariesOf(qPostPunk)
		if !equalLabels(got, "punk") {
			t.Errorf("post-punk primaries = %v, want [punk] (not rock)", labels(got))
		}
	})

	t.Run("disco maps to disco, not R&B", func(t *testing.T) {
		// disco descends from rhythm and blues; as its own primary it wins.
		got := dag.PrimariesOf(qDisco)
		if !equalLabels(got, "disco") {
			t.Errorf("disco primaries = %v, want [disco] (not R&B)", labels(got))
		}
	})

	t.Run("single-branch sub-genre maps to its primary", func(t *testing.T) {
		got := dag.PrimariesOf(qBebop)
		if !equalLabels(got, "jazz") {
			t.Errorf("bebop primaries = %v, want [jazz]", labels(got))
		}
	})

	t.Run("genre under no primary is uncategorized", func(t *testing.T) {
		got := dag.PrimariesOf(qAmbient)
		if len(got) != 0 {
			t.Errorf("ambient primaries = %v, want none", labels(got))
		}
	})

	t.Run("unknown id has no primaries", func(t *testing.T) {
		if got := dag.PrimariesOf("Qdoesnotexist"); got != nil {
			t.Errorf("unknown id primaries = %v, want nil", labels(got))
		}
	})

	t.Run("primaries are returned in curated order", func(t *testing.T) {
		// g-funk spans hip-hop and funk; allowlist order is hip-hop before funk.
		got := dag.PrimariesOf(qGFunk)
		if !equalLabels(got, "hip-hop", "funk") {
			t.Errorf("g-funk primaries = %v, want [hip-hop funk]", labels(got))
		}
	})
}
