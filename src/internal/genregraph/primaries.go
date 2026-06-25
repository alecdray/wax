package genregraph

// Primary is a curated top-level genre bucket: one of a small, fixed set of
// broad genres an album is assigned to for coarse filtering.
type Primary struct {
	ID    string
	Label string
}

// primaryAllowlist is the curated, ordered set of primary genres, pinned by
// Wikidata Q-id. Order is the display order for filter chips and for the
// per-album result. The set is intentionally small and broad; nesting is fine
// (metal descends from rock) because assignment keeps only the most-specific
// primary per ancestor path. See ADR 0009.
var primaryAllowlist = []Primary{
	{"Q11399", "rock"},
	{"Q12326776", "metal"},
	{"Q25554481", "punk"},
	{"Q37073", "pop"},
	{"Q9778", "electronic"},
	{"Q11401", "hip-hop"},
	{"Q8341", "jazz"},
	{"Q9730", "classical"},
	{"Q9759", "blues"},
	{"Q83440", "country"},
	{"Q43343", "folk"},
	{"Q9794", "reggae"},
	{"Q131272", "soul"},
	{"Q45981", "R&B"},
	{"Q164444", "funk"},
}

// PrimaryGenres returns the full curated primary set in display order. It needs
// no DAG — the allowlist is static — so view/filter code can list the options
// without loading the graph.
func PrimaryGenres() []Primary {
	out := make([]Primary, len(primaryAllowlist))
	copy(out, primaryAllowlist)
	return out
}

// Primaries returns the full curated primary set, in display order.
func (d *DAG) Primaries() []Primary {
	return PrimaryGenres()
}

// PrimariesOf returns the most-specific primary genres for a genre node, in
// curated order. A node maps to a primary if that primary is the node itself
// or one of its ancestors; where matched primaries nest, only the deepest on
// each path is kept (death metal → metal, not rock), while primaries on
// unrelated branches are all kept (hyperpop → pop + electronic). Returns nil
// for an unknown node or one under no primary.
func (d *DAG) PrimariesOf(genreID string) []Primary {
	return d.primaries[genreID]
}

// computePrimaries precomputes PrimariesOf for every node. Called once at Build.
func (d *DAG) computePrimaries() {
	allow := make(map[string]bool, len(primaryAllowlist))
	for _, p := range primaryAllowlist {
		allow[p.ID] = true
	}

	d.primaries = make(map[string][]Primary, len(d.nodes))
	for id := range d.nodes {
		if ps := d.computePrimariesFor(id, allow); ps != nil {
			d.primaries[id] = ps
		}
	}
}

func (d *DAG) computePrimariesFor(id string, allow map[string]bool) []Primary {
	// Candidate primaries: the node itself and any ancestor that is allowlisted.
	candidates := make(map[string]bool)
	if allow[id] {
		candidates[id] = true
	}
	for _, a := range d.Ancestors(id) {
		if allow[a.ID] {
			candidates[a.ID] = true
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	// Keep only the most-specific: drop any candidate that is a (strict)
	// ancestor of another candidate.
	ancestorOf := make(map[string]map[string]bool, len(candidates))
	for c := range candidates {
		set := make(map[string]bool)
		for _, a := range d.Ancestors(c) {
			set[a.ID] = true
		}
		ancestorOf[c] = set
	}

	var result []Primary
	for _, p := range primaryAllowlist { // preserve curated order
		if !candidates[p.ID] {
			continue
		}
		broader := false
		for other := range candidates {
			if other != p.ID && ancestorOf[other][p.ID] {
				broader = true
				break
			}
		}
		if !broader {
			result = append(result, p)
		}
	}
	return result
}
