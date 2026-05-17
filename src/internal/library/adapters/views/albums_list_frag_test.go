package views

import (
	"net/url"
	"strings"
	"testing"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/library"
)

// URL builder tests pin the URL contract: default sort/dir is omitted from
// the URL so stale params don't accumulate as the user changes their mind.
// Default filters were already dropped by the builder; the extension this
// file pins is that defaultSort = (date, desc) is also bare.

func ptr[T any](v T) *T { return &v }

func parseQuery(t *testing.T, raw string) url.Values {
	t.Helper()
	idx := strings.Index(raw, "?")
	if idx == -1 {
		return url.Values{}
	}
	v, err := url.ParseQuery(raw[idx+1:])
	if err != nil {
		t.Fatalf("parse query %q: %v", raw, err)
	}
	return v
}

func TestBuildAlbumsTableURL_BareWhenDefaults(t *testing.T) {
	got := buildAlbumsTableURL("date", "desc", library.FilterParams{})
	if got != "/app/library/dashboard/albums-table" {
		t.Fatalf("default view URL must be bare; got %q", got)
	}
	got = buildAlbumsTableURL("", "", library.FilterParams{})
	if got != "/app/library/dashboard/albums-table" {
		t.Fatalf("empty sort/dir must be bare; got %q", got)
	}
}

func TestBuildAlbumsTableURL_DropsDefaultSort(t *testing.T) {
	got := buildAlbumsTableURL("date", "desc", library.FilterParams{Q: "beatles"})
	q := parseQuery(t, got)
	if q.Has("sortBy") || q.Has("dir") {
		t.Fatalf("default sort must be dropped; got %q", got)
	}
	if q.Get("q") != "beatles" {
		t.Fatalf("q must round-trip; got %q", got)
	}
}

func TestBuildAlbumsTableURL_KeepsNonDefaultSort(t *testing.T) {
	got := buildAlbumsTableURL("artist", "asc", library.FilterParams{})
	q := parseQuery(t, got)
	if q.Get("sortBy") != "artist" || q.Get("dir") != "asc" {
		t.Fatalf("non-default sort must round-trip; got %q", got)
	}
}

func TestBuildAlbumsTableURL_KeepsNonDefaultDirOnly(t *testing.T) {
	// Direction set to non-default (asc) with default sort field still needs
	// to round-trip — date asc is a meaningfully different sort.
	got := buildAlbumsTableURL("date", "asc", library.FilterParams{})
	q := parseQuery(t, got)
	if q.Get("dir") != "asc" {
		t.Fatalf("non-default dir must round-trip even with default sortBy; got %q", got)
	}
}

func TestBuildAlbumsTableURL_RepeatableMultiSelect(t *testing.T) {
	fp := library.FilterParams{
		Formats:   []models.ReleaseFormat{models.ReleaseFormatVinyl, models.ReleaseFormatCD},
		ArtistIDs: []string{"a-1", "a-2"},
	}
	got := buildAlbumsTableURL("date", "desc", fp)
	q := parseQuery(t, got)
	if len(q["format"]) != 2 {
		t.Fatalf("format must be repeatable; got %q", got)
	}
	if len(q["artist"]) != 2 {
		t.Fatalf("artist must be repeatable; got %q", got)
	}
}

func TestBuildAlbumsTableURL_RatingRoundTrip(t *testing.T) {
	fp := library.FilterParams{MinRating: ptr(7.0), MaxRating: ptr(9.5), Rated: "only"}
	got := buildAlbumsTableURL("date", "desc", fp)
	q := parseQuery(t, got)
	if q.Get("minRating") != "7" || q.Get("maxRating") != "9.5" || q.Get("rated") != "only" {
		t.Fatalf("rating params must round-trip; got %q", got)
	}
}

func TestBuildAlbumsTableURLForInput_OmitsQ(t *testing.T) {
	// The input's hx-get URL omits q so the input's own name=value pair is the
	// sole source of q in the request.
	got := buildAlbumsTableURLForInput("date", "desc", library.FilterParams{Q: "beatles"})
	q := parseQuery(t, got)
	if q.Has("q") {
		t.Fatalf("input URL must not include q; got %q", got)
	}
}

func TestBuildAlbumsPageURL_CarriesAllActiveParams(t *testing.T) {
	// Pagination sentinel must carry every active dimension so the appended
	// rows belong to the same narrowed set.
	fp := library.FilterParams{
		Q:         "radio",
		MinRating: ptr(7.0),
		Rated:     "only",
		Formats:   []models.ReleaseFormat{models.ReleaseFormatVinyl},
		ArtistIDs: []string{"a-1"},
	}
	got := buildAlbumsPageURL(20, "artist", "asc", fp)
	q := parseQuery(t, got)

	if q.Get("offset") != "20" {
		t.Fatalf("offset missing; got %q", got)
	}
	if q.Get("q") != "radio" {
		t.Fatalf("q missing; got %q", got)
	}
	if q.Get("sortBy") != "artist" || q.Get("dir") != "asc" {
		t.Fatalf("sort missing; got %q", got)
	}
	if q.Get("minRating") != "7" || q.Get("rated") != "only" {
		t.Fatalf("rating missing; got %q", got)
	}
	if len(q["format"]) != 1 || q["format"][0] != "vinyl" {
		t.Fatalf("format missing; got %q", got)
	}
	if len(q["artist"]) != 1 || q["artist"][0] != "a-1" {
		t.Fatalf("artist missing; got %q", got)
	}
}

func TestBuildAlbumsPageURL_DropsDefaultSort(t *testing.T) {
	got := buildAlbumsPageURL(20, "date", "desc", library.FilterParams{Q: "beatles"})
	q := parseQuery(t, got)
	if q.Has("sortBy") || q.Has("dir") {
		t.Fatalf("default sort must be dropped from pagination URL; got %q", got)
	}
	if q.Get("offset") != "20" || q.Get("q") != "beatles" {
		t.Fatalf("non-default params must round-trip; got %q", got)
	}
}

func TestIsDefaultView(t *testing.T) {
	cases := []struct {
		name    string
		sortBy  string
		sortDir string
		fp      library.FilterParams
		want    bool
	}{
		{"empty everything", "", "", library.FilterParams{}, true},
		{"explicit defaults", "date", "desc", library.FilterParams{}, true},
		{"q set", "date", "desc", library.FilterParams{Q: "x"}, false},
		{"non-default sort field", "artist", "desc", library.FilterParams{}, false},
		{"non-default sort dir", "date", "asc", library.FilterParams{}, false},
		{"min rating set", "date", "desc", library.FilterParams{MinRating: ptr(7.0)}, false},
		{"rated unrated", "date", "desc", library.FilterParams{Rated: "unrated"}, false},
		{"format selected", "date", "desc", library.FilterParams{Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl}}, false},
		{"artist selected", "date", "desc", library.FilterParams{ArtistIDs: []string{"a"}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isDefaultView(c.sortBy, c.sortDir, c.fp); got != c.want {
				t.Fatalf("isDefaultView(%q, %q, %+v) = %v; want %v", c.sortBy, c.sortDir, c.fp, got, c.want)
			}
		})
	}
}
