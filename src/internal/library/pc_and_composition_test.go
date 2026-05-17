package library

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/review"
)

// Integration tests for the wax-unified-search build's product criteria.
//
// These tests assert PC1 and PC3 as whole-system invariants of the assembled
// search/filter/sort pipeline, distinct from per-task tests. They drive
// realistic combinations of (q, filters, sort) through Library.Filter +
// SortBy*, then compare the result against an independently-computed
// reference implementation (`referenceFilteredSorted` below). The reference
// is deliberately naive — a predicate loop and a stable sort.Slice with
// hand-written less functions — so the two implementations agree only when
// the production pipeline correctly composes its parts.
//
// PC1 — AND composition across text, filters, and sort. For any (q, filter,
//   sort) combination, the rendered set equals the AND of all dimensions in
//   the active sort order. Tested by `TestPC1_AndCompositionMatrix` which
//   walks a Cartesian product of representative parameter values.
//
// PC3 — No regressions to pre-existing filter/sort behavior. With q="",
//   every filter × sort combination produces what the unchanged pipeline
//   produces. Tested by `TestPC3_NoRegressionWhenQEmpty`. (The reference
//   for "before this build" is the same Filter/Sort pipeline minus the Q
//   field — Q matching is the only addition to Filter, and SortBy* is
//   untouched.)
//
// PC2 (URL round-trip) is verified at the e2e layer — see
// `e2e/spec/library.spec.ts` PC2 tests, which require a browser.

// fixtureAlbums returns a small but representative library: titles and
// artist names sampled to exercise Q substring matching (case, partial,
// whitespace, artist-vs-title), formats covering both physical and
// digital, ratings spanning rated/unrated, and AddedAt dates spread far
// enough apart to give SortByDate a deterministic order.
func fixtureAlbums(t *testing.T) AlbumDTOs {
	t.Helper()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	mk := func(idx int, title string, artist string, ratingVal *float64, format models.ReleaseFormat, lastPlayedOffsetDays int) AlbumDTO {
		addedAt := base.AddDate(0, 0, idx)
		a := AlbumDTO{
			ID:    fmt.Sprintf("album-%d", idx),
			Title: title,
			Artists: []ArtistDTO{
				{ID: "artist-" + artist, Name: artist},
			},
			Releases: ReleaseDTOs{{Format: format, AddedAt: &addedAt}},
		}
		if ratingVal != nil {
			a.Rating = &review.AlbumRatingDTO{Rating: ratingVal}
		}
		if lastPlayedOffsetDays >= 0 {
			lp := base.AddDate(0, 0, lastPlayedOffsetDays)
			a.LastPlayedAt = &lp
		}
		return a
	}
	r := func(v float64) *float64 { return &v }
	// Titles deliberately include shared substrings ("the", "love"),
	// case differences (Brian vs. brian), and an artist whose name
	// matches another album's title ("Beatles" / "Beatles Forever").
	return AlbumDTOs{
		mk(1, "Abbey Road", "The Beatles", r(9.5), models.ReleaseFormatVinyl, 5),
		mk(2, "Kid A", "Radiohead", r(9.0), models.ReleaseFormatDigital, -1),
		mk(3, "OK Computer", "Radiohead", r(8.5), models.ReleaseFormatVinyl, 12),
		mk(4, "Love Supreme", "John Coltrane", r(9.0), models.ReleaseFormatCD, -1),
		mk(5, "Love Forever", "Brian Eno", nil, models.ReleaseFormatCassette, 30),
		mk(6, "Sea Change", "Beck", r(7.5), models.ReleaseFormatVinyl, 22),
		mk(7, "In Rainbows", "Radiohead", nil, models.ReleaseFormatDigital, -1),
		mk(8, "The Wall", "Pink Floyd", r(8.0), models.ReleaseFormatVinyl, 8),
		mk(9, "Plastic Ono Band", "John Lennon", r(7.0), models.ReleaseFormatCD, -1),
		mk(10, "Beatles Forever", "Cover Band", nil, models.ReleaseFormatDigital, 1),
	}
}

// referenceFilteredSorted is an independent implementation of the assembled
// pipeline. It applies the Q + filter predicates directly (no shortcut for
// the all-defaults case) and sorts via sort.Slice with explicit less
// functions. It must agree with `albums.Filter(fp)` followed by the
// matching SortBy* call.
//
// The reference is structurally distinct from the production code: it uses
// the explicit predicate form rather than the early-return guard, and it
// shares no helpers with Library.Filter / SortBy*. Agreement therefore
// demonstrates that the production pipeline composes correctly across all
// inputs.
func referenceFilteredSorted(albums AlbumDTOs, fp FilterParams, sortBy, dir string) AlbumDTOs {
	q := strings.TrimSpace(strings.ToLower(fp.Q))

	out := make(AlbumDTOs, 0, len(albums))
albumLoop:
	for _, a := range albums {
		// Q: case-insensitive substring against title OR any artist name.
		if q != "" {
			titleMatch := strings.Contains(strings.ToLower(a.Title), q)
			artistMatch := false
			for _, ar := range a.Artists {
				if strings.Contains(strings.ToLower(ar.Name), q) {
					artistMatch = true
					break
				}
			}
			if !titleMatch && !artistMatch {
				continue
			}
		}
		// MinRating
		if fp.MinRating != nil {
			if a.Rating == nil || a.Rating.Rating == nil || *a.Rating.Rating < *fp.MinRating {
				continue
			}
		}
		// MaxRating
		if fp.MaxRating != nil {
			if a.Rating == nil || a.Rating.Rating == nil || *a.Rating.Rating > *fp.MaxRating {
				continue
			}
		}
		// Rated
		switch fp.Rated {
		case "only":
			if a.Rating == nil || a.Rating.Rating == nil {
				continue
			}
		case "unrated":
			if a.Rating != nil && a.Rating.Rating != nil {
				continue
			}
		}
		// Formats: any release in any of the requested formats with AddedAt set.
		if len(fp.Formats) > 0 {
			ok := false
			for _, want := range fp.Formats {
				for _, rel := range a.Releases {
					if rel.Format == want && rel.AddedAt != nil {
						ok = true
						break
					}
				}
				if ok {
					break
				}
			}
			if !ok {
				continue
			}
		}
		// ArtistIDs: at least one of the album's artists is in the set.
		if len(fp.ArtistIDs) > 0 {
			ok := false
			for _, want := range fp.ArtistIDs {
				for _, ar := range a.Artists {
					if ar.ID == want {
						ok = true
						break
					}
				}
				if ok {
					break
				}
			}
			if !ok {
				continue albumLoop
			}
		}
		out = append(out, a)
	}

	asc := dir == "asc"
	// Default sort field is "date"; empty sortBy resolves to that.
	field := sortBy
	if field == "" {
		field = "date"
	}
	switch field {
	case "album":
		sort.SliceStable(out, func(i, j int) bool {
			if asc {
				return out[i].Title < out[j].Title
			}
			return out[i].Title > out[j].Title
		})
	case "artist":
		sort.SliceStable(out, func(i, j int) bool {
			ni, nj := "", ""
			if len(out[i].Artists) > 0 {
				ni = out[i].Artists[0].Name
			}
			if len(out[j].Artists) > 0 {
				nj = out[j].Artists[0].Name
			}
			// Mirror the production tie-break for missing artists.
			if ni == "" && nj == "" {
				return false
			}
			if ni == "" {
				return asc
			}
			if nj == "" {
				return !asc
			}
			if asc {
				return ni < nj
			}
			return ni > nj
		})
	case "rating":
		sort.SliceStable(out, func(i, j int) bool {
			var ri, rj *float64
			if out[i].Rating != nil {
				ri = out[i].Rating.Rating
			}
			if out[j].Rating != nil {
				rj = out[j].Rating.Rating
			}
			if ri == nil && rj == nil {
				return false
			}
			if ri == nil {
				return asc
			}
			if rj == nil {
				return !asc
			}
			if asc {
				return *ri < *rj
			}
			return *ri > *rj
		})
	case "lastPlayed":
		sort.SliceStable(out, func(i, j int) bool {
			li, lj := out[i].LastPlayedAt, out[j].LastPlayedAt
			if li == nil && lj == nil {
				return false
			}
			if li == nil {
				return asc
			}
			if lj == nil {
				return !asc
			}
			if asc {
				return li.Before(*lj)
			}
			return li.After(*lj)
		})
	default: // "date"
		sort.SliceStable(out, func(i, j int) bool {
			di := out[i].Releases.OldestAddedAtDate()
			dj := out[j].Releases.OldestAddedAtDate()
			if di == nil && dj == nil {
				return false
			}
			if di == nil {
				return asc
			}
			if dj == nil {
				return !asc
			}
			if asc {
				return di.Before(*dj)
			}
			return di.After(*dj)
		})
	}
	return out
}

// productionFilteredSorted runs the assembled production pipeline: sort the
// in-memory slice with the same dispatch the HTTP handler uses, then apply
// Filter with the given params. The handler in adapters/http.go does sort
// first then filter; this helper mirrors that order so the comparison is
// against what the dashboard actually renders.
func productionFilteredSorted(albums AlbumDTOs, fp FilterParams, sortBy, dir string) AlbumDTOs {
	// Copy the slice so the in-place sort doesn't mutate the caller.
	cp := make(AlbumDTOs, len(albums))
	copy(cp, albums)

	asc := dir == "asc"
	switch sortBy {
	case "album":
		cp.SortByTitle(asc)
	case "artist":
		cp.SortByArtist(asc)
	case "rating":
		cp.SortByRating(asc)
	case "lastPlayed":
		cp.SortByLastPlayed(asc)
	default:
		cp.SortByDate(asc)
	}
	return cp.Filter(fp)
}

func assertSameIDsInOrder(t *testing.T, label string, got, want AlbumDTOs) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: length mismatch got=%d want=%d\ngot IDs=%v\nwant IDs=%v",
			label, len(got), len(want), idsOf(got), idsOf(want))
	}
	for i := range got {
		if got[i].ID != want[i].ID {
			t.Fatalf("%s: order/identity mismatch at position %d got=%s want=%s\nfull got=%v\nfull want=%v",
				label, i, got[i].ID, want[i].ID, idsOf(got), idsOf(want))
		}
	}
}

func idsOf(albums AlbumDTOs) []string {
	out := make([]string, len(albums))
	for i, a := range albums {
		out[i] = a.ID
	}
	return out
}

// --- PC1: AND composition across text, filters, and sort ---

// TestPC1_AndCompositionMatrix walks a representative Cartesian product of
// (q, filter, sort) and asserts that the assembled pipeline produces the
// same set in the same order as the reference. This is the PC1 invariant:
// the rendered list IS the AND of text + every active filter, ordered by
// the active sort, for every reachable combination.
func TestPC1_AndCompositionMatrix(t *testing.T) {
	albums := fixtureAlbums(t)

	queries := []string{
		"",         // empty matches everything
		"love",     // matches by title (Love Supreme, Love Forever)
		"radio",    // matches by artist (Radiohead × 3)
		"BEATLES",  // mixed-case, matches title (Beatles Forever) AND artist (The Beatles)
		"the",      // very broad — title ("The Wall"), artist ("The Beatles")
		"xxxxnone", // zero-result
	}
	filters := []FilterParams{
		{}, // no filter
		{MinRating: ptr(8.0)},
		{MaxRating: ptr(8.0)},
		{Rated: "only"},
		{Rated: "unrated"},
		{Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl}},
		{Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl, models.ReleaseFormatDigital}},
		{ArtistIDs: []string{"artist-Radiohead"}},
		{ArtistIDs: []string{"artist-The Beatles", "artist-Pink Floyd"}},
		// Combined: rating + format — neither alone identifies the same set
		{MinRating: ptr(8.0), Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl}},
	}
	type sortKey struct{ by, dir string }
	sorts := []sortKey{
		{"", ""},              // default — date desc
		{"date", "desc"},      // explicit default
		{"date", "asc"},
		{"album", "asc"},
		{"album", "desc"},
		{"artist", "asc"},
		{"rating", "desc"},
		{"rating", "asc"},
		{"lastPlayed", "desc"},
		{"lastPlayed", "asc"},
	}

	for _, q := range queries {
		for fi, f := range filters {
			for _, s := range sorts {
				fp := f
				fp.Q = q
				label := fmt.Sprintf("q=%q filter#%d sort=%s/%s", q, fi, s.by, s.dir)
				t.Run(label, func(t *testing.T) {
					got := productionFilteredSorted(albums, fp, s.by, s.dir)
					want := referenceFilteredSorted(albums, fp, s.by, s.dir)
					assertSameIDsInOrder(t, label, got, want)
				})
			}
		}
	}
}

// TestPC1_DefaultsYieldFullLibraryInDefaultOrder pins the second sentence
// of PC1: empty text + default filters + default sort yields the full
// library in the default order (date desc).
func TestPC1_DefaultsYieldFullLibraryInDefaultOrder(t *testing.T) {
	albums := fixtureAlbums(t)

	// Default: empty FilterParams, empty sortBy/dir (defaults to date desc).
	got := productionFilteredSorted(albums, FilterParams{}, "", "")

	if len(got) != len(albums) {
		t.Fatalf("default view must contain the full library; got %d, want %d", len(got), len(albums))
	}

	// Verify the order is date desc — newest AddedAt first.
	for i := 1; i < len(got); i++ {
		prev := got[i-1].Releases.OldestAddedAtDate()
		cur := got[i].Releases.OldestAddedAtDate()
		if prev == nil || cur == nil {
			t.Fatalf("fixture must have AddedAt set on all releases (i=%d)", i)
		}
		if cur.After(*prev) {
			t.Fatalf("default order should be date desc; row %d (%s, %v) is newer than row %d (%s, %v)",
				i, got[i].ID, *cur, i-1, got[i-1].ID, *prev)
		}
	}
}

// --- PC3: No regressions to pre-existing filter/sort behavior ---

// TestPC3_NoRegressionWhenQEmpty asserts the pre-build invariant: with no
// Q, the assembled pipeline produces the same list (in the same order) as
// it would have produced before this build. The reference IS the
// pre-build behavior, because:
//
//   - SortBy* is structurally unchanged by this build (verified by git
//     blame on library.go in the build commits).
//   - Filter gained one new field (Q) and one early-return guard for
//     Q==""; with Q="" the predicate path is identical to before.
//
// The test enumerates every filter dimension × every sort field × both
// directions, and asserts that production matches the reference. Failure
// would indicate that the build changed pre-existing filter/sort behavior.
func TestPC3_NoRegressionWhenQEmpty(t *testing.T) {
	albums := fixtureAlbums(t)

	filters := []FilterParams{
		{},
		{MinRating: ptr(7.0)},
		{MaxRating: ptr(8.5)},
		{MinRating: ptr(7.0), MaxRating: ptr(9.0)},
		{Rated: "only"},
		{Rated: "unrated"},
		{Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl}},
		{Formats: []models.ReleaseFormat{models.ReleaseFormatCD, models.ReleaseFormatCassette}},
		{ArtistIDs: []string{"artist-Radiohead"}},
		{ArtistIDs: []string{"artist-Radiohead", "artist-The Beatles"}},
		{MinRating: ptr(7.0), Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl}},
	}
	sortFields := []string{"", "date", "album", "artist", "rating", "lastPlayed"}
	dirs := []string{"desc", "asc"}

	for fi, f := range filters {
		for _, sb := range sortFields {
			for _, d := range dirs {
				fp := f // Q stays empty
				label := fmt.Sprintf("filter#%d sortBy=%q dir=%s", fi, sb, d)
				t.Run(label, func(t *testing.T) {
					got := productionFilteredSorted(albums, fp, sb, d)
					want := referenceFilteredSorted(albums, fp, sb, d)
					assertSameIDsInOrder(t, label, got, want)
				})
			}
		}
	}
}

// TestPC3_BareURLPipelineIdentity is a property-style check: for every
// (filter, sort) the dashboard handler is reachable from a pre-existing
// URL (no `q` param), the assembled pipeline output equals the reference.
// This is the data-layer half of "pre-existing URLs render identically to
// before" — the URL parsing half is exercised in e2e.
func TestPC3_BareURLPipelineIdentity(t *testing.T) {
	albums := fixtureAlbums(t)

	// All combinations a pre-existing URL could have produced — q is
	// implicitly absent (empty string post-trim).
	for _, sortBy := range []string{"", "date", "album", "artist", "rating", "lastPlayed"} {
		for _, dir := range []string{"", "desc", "asc"} {
			for _, fp := range []FilterParams{
				{},
				{Rated: "only"},
				{Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl}},
				{ArtistIDs: []string{"artist-Radiohead"}},
				{MinRating: ptr(8.0)},
			} {
				got := productionFilteredSorted(albums, fp, sortBy, dir)
				want := referenceFilteredSorted(albums, fp, sortBy, dir)
				if len(got) != len(want) {
					t.Fatalf("bare URL parity failed: sortBy=%q dir=%q filter=%+v: len got=%d want=%d",
						sortBy, dir, fp, len(got), len(want))
				}
				for i := range got {
					if got[i].ID != want[i].ID {
						t.Fatalf("bare URL parity failed: sortBy=%q dir=%q filter=%+v: pos %d got=%s want=%s",
							sortBy, dir, fp, i, got[i].ID, want[i].ID)
					}
				}
			}
		}
	}
}
