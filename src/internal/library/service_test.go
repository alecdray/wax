package library

import (
	"testing"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/genregraph"
	"github.com/alecdray/wax/src/internal/review"
	spotifylib "github.com/zmb3/spotify/v2"
)

// makeAlbumWithRelease creates an AlbumDTO with a single release format.
func makeAlbumWithRelease(id, title string, format models.ReleaseFormat) AlbumDTO {
	now := time.Now()
	return AlbumDTO{
		ID:       id,
		Title:    title,
		Releases: ReleaseDTOs{{Format: format, AddedAt: &now}},
	}
}

// helpers

func ptr[T any](v T) *T { return &v }

func makeAlbum(id, title string, artistName string, rating *float64, lastPlayed *time.Time) AlbumDTO {
	album := AlbumDTO{
		ID:           id,
		Title:        title,
		LastPlayedAt: lastPlayed,
	}
	if artistName != "" {
		album.Artists = []ArtistDTO{{ID: artistName, Name: artistName}}
	}
	if rating != nil {
		album.Rating = &review.AlbumRatingDTO{Rating: rating}
	}
	return album
}

// --- AlbumDTOs.Page ---

func TestPage_ReturnsFirstPage(t *testing.T) {
	albums := make(AlbumDTOs, 25)
	for i := range albums {
		albums[i].ID = string(rune('a' + i))
	}

	page := albums.Page(0)
	if len(page) != AlbumsPageSize {
		t.Fatalf("expected %d albums, got %d", AlbumsPageSize, len(page))
	}
}

func TestPage_ReturnsPartialLastPage(t *testing.T) {
	albums := make(AlbumDTOs, 25)
	page := albums.Page(20)
	if len(page) != 5 {
		t.Fatalf("expected 5 albums, got %d", len(page))
	}
}

func TestPage_ReturnsNilWhenOffsetBeyondLength(t *testing.T) {
	albums := make(AlbumDTOs, 10)
	page := albums.Page(20)
	if page != nil {
		t.Fatal("expected nil page")
	}
}

// --- SortByTitle ---

func TestSortByTitle_Ascending(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Ziggy", "", nil, nil),
		makeAlbum("2", "Abbey", "", nil, nil),
		makeAlbum("3", "Moon", "", nil, nil),
	}
	albums.SortByTitle(true)
	if albums[0].Title != "Abbey" || albums[2].Title != "Ziggy" {
		t.Fatalf("unexpected order: %v %v %v", albums[0].Title, albums[1].Title, albums[2].Title)
	}
}

func TestSortByTitle_Descending(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Ziggy", "", nil, nil),
		makeAlbum("2", "Abbey", "", nil, nil),
		makeAlbum("3", "Moon", "", nil, nil),
	}
	albums.SortByTitle(false)
	if albums[0].Title != "Ziggy" || albums[2].Title != "Abbey" {
		t.Fatalf("unexpected order: %v %v %v", albums[0].Title, albums[1].Title, albums[2].Title)
	}
}

// --- SortByArtist ---

func TestSortByArtist_Ascending(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "A", "Zeppelin", nil, nil),
		makeAlbum("2", "B", "Beatles", nil, nil),
		makeAlbum("3", "C", "Arcade Fire", nil, nil),
	}
	albums.SortByArtist(true)
	if albums[0].Artists[0].Name != "Arcade Fire" || albums[2].Artists[0].Name != "Zeppelin" {
		t.Fatalf("unexpected order: %v %v %v",
			albums[0].Artists[0].Name, albums[1].Artists[0].Name, albums[2].Artists[0].Name)
	}
}

func TestSortByArtist_NoArtistGoesFirst_Ascending(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "A", "", nil, nil),
		makeAlbum("2", "B", "Beatles", nil, nil),
	}
	albums.SortByArtist(true)
	// Album with no artist sorts first in ascending order (consistent with SortByArtist implementation)
	if albums[0].ID != "1" {
		t.Fatalf("expected album without artist first, got %v", albums[0].ID)
	}
}

// --- SortByRating ---

func TestSortByRating_Descending(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Low", "", ptr(6.0), nil),
		makeAlbum("2", "High", "", ptr(9.5), nil),
		makeAlbum("3", "Mid", "", ptr(7.5), nil),
	}
	albums.SortByRating(false)
	if *albums[0].Rating.Rating != 9.5 || *albums[2].Rating.Rating != 6.0 {
		t.Fatalf("unexpected order: %v %v %v",
			*albums[0].Rating.Rating, *albums[1].Rating.Rating, *albums[2].Rating.Rating)
	}
}

func TestSortByRating_NilRatingGoesLast_Descending(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Rated", "", ptr(8.0), nil),
		makeAlbum("2", "Unrated", "", nil, nil),
	}
	albums.SortByRating(false)
	if albums[0].ID != "1" {
		t.Fatalf("expected rated album first, got %v", albums[0].ID)
	}
}

// --- SortByLastPlayed ---

func TestSortByLastPlayed_Descending(t *testing.T) {
	now := time.Now()
	older := now.Add(-48 * time.Hour)
	albums := AlbumDTOs{
		makeAlbum("1", "Old", "", nil, &older),
		makeAlbum("2", "New", "", nil, &now),
	}
	albums.SortByLastPlayed(false)
	if albums[0].ID != "2" {
		t.Fatalf("expected most recently played first, got %v", albums[0].ID)
	}
}

func TestSortByLastPlayed_NilGoesLast_Descending(t *testing.T) {
	now := time.Now()
	albums := AlbumDTOs{
		makeAlbum("1", "Never played", "", nil, nil),
		makeAlbum("2", "Played", "", nil, &now),
	}
	albums.SortByLastPlayed(false)
	if albums[0].ID != "2" {
		t.Fatalf("expected played album first, got %v", albums[0].ID)
	}
}

// --- SortByDate ---

func TestSortByDate_Descending(t *testing.T) {
	now := time.Now()
	older := now.Add(-30 * 24 * time.Hour)

	albums := AlbumDTOs{
		{
			ID:    "1",
			Title: "Old",
			Releases: ReleaseDTOs{
				{Format: models.ReleaseFormatVinyl, AddedAt: &older},
			},
		},
		{
			ID:    "2",
			Title: "New",
			Releases: ReleaseDTOs{
				{Format: models.ReleaseFormatDigital, AddedAt: &now},
			},
		},
	}
	albums.SortByDate(false)
	if albums[0].ID != "2" {
		t.Fatalf("expected newer album first, got %v", albums[0].ID)
	}
}

// --- ReleaseDTOs ---

func TestOldestAddedAtDate(t *testing.T) {
	now := time.Now()
	older := now.Add(-7 * 24 * time.Hour)
	oldest := now.Add(-30 * 24 * time.Hour)

	releases := ReleaseDTOs{
		{AddedAt: &now},
		{AddedAt: &oldest},
		{AddedAt: &older},
	}

	result := releases.OldestAddedAtDate()
	if result == nil || !result.Equal(oldest) {
		t.Fatalf("expected oldest date, got %v", result)
	}
}

func TestOldestAddedAtDate_AllNil(t *testing.T) {
	releases := ReleaseDTOs{{}, {}}
	if releases.OldestAddedAtDate() != nil {
		t.Fatal("expected nil")
	}
}

func TestFindFormat(t *testing.T) {
	vinyl := ReleaseDTO{ID: "v", Format: models.ReleaseFormatVinyl}
	releases := ReleaseDTOs{
		{ID: "d", Format: models.ReleaseFormatDigital},
		vinyl,
	}
	result := releases.FindFormat(models.ReleaseFormatVinyl)
	if result == nil || result.ID != "v" {
		t.Fatalf("expected vinyl release, got %v", result)
	}
}

func TestFindFormat_NotFound(t *testing.T) {
	releases := ReleaseDTOs{
		{Format: models.ReleaseFormatDigital},
	}
	if releases.FindFormat(models.ReleaseFormatVinyl) != nil {
		t.Fatal("expected nil for missing format")
	}
}

// --- AlbumDTOs.Filter ---

func TestFilter_NoParams_ReturnsAll(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "A", "", nil, nil),
		makeAlbum("2", "B", "", ptr(7.0), nil),
	}
	result := albums.Filter(FilterParams{})
	if len(result) != 2 {
		t.Fatalf("expected 2 albums, got %d", len(result))
	}
}

func TestFilter_MinRating(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Low", "", ptr(5.0), nil),
		makeAlbum("2", "High", "", ptr(8.0), nil),
		makeAlbum("3", "Unrated", "", nil, nil),
	}
	result := albums.Filter(FilterParams{MinRating: ptr(7.0)})
	if len(result) != 1 || result[0].ID != "2" {
		t.Fatalf("expected only high-rated album, got %d albums", len(result))
	}
}

func TestFilter_MaxRating(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Low", "", ptr(4.0), nil),
		makeAlbum("2", "High", "", ptr(9.0), nil),
		makeAlbum("3", "Unrated", "", nil, nil),
	}
	result := albums.Filter(FilterParams{MaxRating: ptr(6.0)})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected only low-rated album, got %d albums", len(result))
	}
}

func TestFilter_MinAndMaxRating(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Low", "", ptr(3.0), nil),
		makeAlbum("2", "Mid", "", ptr(7.0), nil),
		makeAlbum("3", "High", "", ptr(10.0), nil),
	}
	result := albums.Filter(FilterParams{MinRating: ptr(6.0), MaxRating: ptr(8.0)})
	if len(result) != 1 || result[0].ID != "2" {
		t.Fatalf("expected only mid-rated album, got %d albums", len(result))
	}
}

func TestFilter_RatedOnly(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Rated", "", ptr(7.0), nil),
		makeAlbum("2", "Unrated", "", nil, nil),
	}
	result := albums.Filter(FilterParams{Rated: "only"})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected only rated album, got %d albums", len(result))
	}
}

func TestFilter_UnratedOnly(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Rated", "", ptr(7.0), nil),
		makeAlbum("2", "Unrated", "", nil, nil),
	}
	result := albums.Filter(FilterParams{Rated: "unrated"})
	if len(result) != 1 || result[0].ID != "2" {
		t.Fatalf("expected only unrated album, got %d albums", len(result))
	}
}

func makeAlbumWithPrimaries(id, title string, primaries ...genregraph.Primary) AlbumDTO {
	a := makeAlbum(id, title, "", nil, nil)
	a.Primaries = primaries
	return a
}

func TestFilter_PrimaryGenre_MatchesSelected(t *testing.T) {
	rock := genregraph.Primary{ID: "Q11399", Label: "rock"}
	pop := genregraph.Primary{ID: "Q37073", Label: "pop"}
	albums := AlbumDTOs{
		makeAlbumWithPrimaries("1", "Rock Album", rock),
		makeAlbumWithPrimaries("2", "Pop Album", pop),
		makeAlbumWithPrimaries("3", "Uncategorized Album"),
	}
	result := albums.Filter(FilterParams{PrimaryIDs: []string{"Q11399"}})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected only rock album, got %d albums", len(result))
	}
}

func TestFilter_PrimaryGenre_OrSemanticsAcrossSelections(t *testing.T) {
	rock := genregraph.Primary{ID: "Q11399", Label: "rock"}
	pop := genregraph.Primary{ID: "Q37073", Label: "pop"}
	jazz := genregraph.Primary{ID: "Q8341", Label: "jazz"}
	albums := AlbumDTOs{
		makeAlbumWithPrimaries("1", "Rock Album", rock),
		makeAlbumWithPrimaries("2", "Pop Album", pop),
		makeAlbumWithPrimaries("3", "Jazz Album", jazz),
	}
	result := albums.Filter(FilterParams{PrimaryIDs: []string{"Q11399", "Q37073"}})
	if len(result) != 2 {
		t.Fatalf("expected rock + pop albums, got %d", len(result))
	}
}

func TestFilter_PrimaryGenre_MultiPrimaryAlbumMatchesEither(t *testing.T) {
	pop := genregraph.Primary{ID: "Q37073", Label: "pop"}
	electronic := genregraph.Primary{ID: "Q9778", Label: "electronic"}
	hyperpop := makeAlbumWithPrimaries("1", "Hyperpop Album", pop, electronic)
	albums := AlbumDTOs{hyperpop}

	if got := albums.Filter(FilterParams{PrimaryIDs: []string{"Q9778"}}); len(got) != 1 {
		t.Fatalf("multi-primary album should match electronic, got %d", len(got))
	}
	if got := albums.Filter(FilterParams{PrimaryIDs: []string{"Q37073"}}); len(got) != 1 {
		t.Fatalf("multi-primary album should match pop, got %d", len(got))
	}
}

func TestFilter_PrimaryGenre_Uncategorized(t *testing.T) {
	rock := genregraph.Primary{ID: "Q11399", Label: "rock"}
	albums := AlbumDTOs{
		makeAlbumWithPrimaries("1", "Rock Album", rock),
		makeAlbumWithPrimaries("2", "Uncategorized Album"),
	}
	result := albums.Filter(FilterParams{PrimaryIDs: []string{UncategorizedPrimaryID}})
	if len(result) != 1 || result[0].ID != "2" {
		t.Fatalf("expected only uncategorized album, got %d albums", len(result))
	}
}

func TestFilter_Format(t *testing.T) {
	vinyl := makeAlbumWithRelease("1", "Vinyl Album", models.ReleaseFormatVinyl)
	digital := makeAlbumWithRelease("2", "Digital Album", models.ReleaseFormatDigital)
	albums := AlbumDTOs{vinyl, digital}

	result := albums.Filter(FilterParams{Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl}})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected only vinyl album, got %d albums", len(result))
	}
}

func TestFilter_Format_ExcludesNoAddedAt(t *testing.T) {
	// A release without AddedAt means it's not in the library for that format
	albums := AlbumDTOs{
		{
			ID:       "1",
			Releases: ReleaseDTOs{{Format: models.ReleaseFormatVinyl, AddedAt: nil}},
		},
	}
	result := albums.Filter(FilterParams{Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl}})
	if len(result) != 0 {
		t.Fatalf("expected 0 albums (no AddedAt), got %d", len(result))
	}
}

func TestFilter_SingleArtist(t *testing.T) {
	albums := AlbumDTOs{
		{ID: "1", Artists: []ArtistDTO{{ID: "artist-a", Name: "Artist A"}}},
		{ID: "2", Artists: []ArtistDTO{{ID: "artist-b", Name: "Artist B"}}},
	}
	result := albums.Filter(FilterParams{ArtistIDs: []string{"artist-a"}})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected album by artist-a only, got %d albums", len(result))
	}
}

func TestFilter_MultipleArtists(t *testing.T) {
	albums := AlbumDTOs{
		{ID: "1", Artists: []ArtistDTO{{ID: "artist-a", Name: "Artist A"}}},
		{ID: "2", Artists: []ArtistDTO{{ID: "artist-b", Name: "Artist B"}}},
		{ID: "3", Artists: []ArtistDTO{{ID: "artist-c", Name: "Artist C"}}},
	}
	result := albums.Filter(FilterParams{ArtistIDs: []string{"artist-a", "artist-b"}})
	if len(result) != 2 {
		t.Fatalf("expected 2 albums, got %d", len(result))
	}
}

func TestFilter_CombinedFormatAndRating(t *testing.T) {
	now := time.Now()
	albums := AlbumDTOs{
		{ID: "1", Releases: ReleaseDTOs{{Format: models.ReleaseFormatVinyl, AddedAt: &now}}, Rating: &review.AlbumRatingDTO{Rating: ptr(8.0)}},
		{ID: "2", Releases: ReleaseDTOs{{Format: models.ReleaseFormatVinyl, AddedAt: &now}}, Rating: &review.AlbumRatingDTO{Rating: ptr(5.0)}},
		{ID: "3", Releases: ReleaseDTOs{{Format: models.ReleaseFormatDigital, AddedAt: &now}}, Rating: &review.AlbumRatingDTO{Rating: ptr(9.0)}},
	}
	result := albums.Filter(FilterParams{
		Formats:   []models.ReleaseFormat{models.ReleaseFormatVinyl},
		MinRating: ptr(7.0),
	})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected only album 1, got %d albums", len(result))
	}
}

// --- AlbumDTOs.Filter — text query (Q) ---
//
// Contract for the Q field: case-insensitive substring match against album
// title and credited artist names; empty query returns the full slice;
// AND-composes with the other filter dimensions.

func makeAlbumTitled(id, title string, artistNames ...string) AlbumDTO {
	album := AlbumDTO{ID: id, Title: title}
	for _, name := range artistNames {
		album.Artists = append(album.Artists, ArtistDTO{ID: name, Name: name})
	}
	return album
}

func TestFilter_Q_EmptyReturnsAll(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbumTitled("1", "Abbey Road", "The Beatles"),
		makeAlbumTitled("2", "Kid A", "Radiohead"),
	}
	result := albums.Filter(FilterParams{Q: ""})
	if len(result) != 2 {
		t.Fatalf("empty Q must match every album; got %d / 2", len(result))
	}
}

func TestFilter_Q_WhitespaceOnlyReturnsAll(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbumTitled("1", "Abbey Road", "The Beatles"),
		makeAlbumTitled("2", "Kid A", "Radiohead"),
	}
	result := albums.Filter(FilterParams{Q: "   "})
	if len(result) != 2 {
		t.Fatalf("whitespace-only Q must match every album; got %d / 2", len(result))
	}
}

func TestFilter_Q_MatchesTitleSubstring(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbumTitled("1", "Abbey Road", "The Beatles"),
		makeAlbumTitled("2", "Kid A", "Radiohead"),
		makeAlbumTitled("3", "OK Computer", "Radiohead"),
	}
	result := albums.Filter(FilterParams{Q: "abbey"})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected only Abbey Road; got %+v", result)
	}
}

func TestFilter_Q_MatchesArtistNameSubstring(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbumTitled("1", "Abbey Road", "The Beatles"),
		makeAlbumTitled("2", "Kid A", "Radiohead"),
		makeAlbumTitled("3", "OK Computer", "Radiohead"),
	}
	result := albums.Filter(FilterParams{Q: "radio"})
	gotIDs := map[string]bool{}
	for _, a := range result {
		gotIDs[a.ID] = true
	}
	if len(result) != 2 || !gotIDs["2"] || !gotIDs["3"] {
		t.Fatalf("expected the two Radiohead albums; got %+v", result)
	}
}

func TestFilter_Q_CaseInsensitive(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbumTitled("1", "Abbey Road", "The Beatles"),
		makeAlbumTitled("2", "Kid A", "Radiohead"),
	}
	upper := albums.Filter(FilterParams{Q: "ABBEY"})
	mixed := albums.Filter(FilterParams{Q: "aBbEy"})
	if len(upper) != 1 || upper[0].ID != "1" {
		t.Fatalf("upper-case Q didn't match; got %+v", upper)
	}
	if len(mixed) != 1 || mixed[0].ID != "1" {
		t.Fatalf("mixed-case Q didn't match; got %+v", mixed)
	}
}

func TestFilter_Q_TitleOrArtistMatchIsSufficient(t *testing.T) {
	// "noise" matches album 1's title only. "head" matches the artist of
	// albums 2/3 only. Each side proves matching one is sufficient — neither
	// album has both sides matching.
	albums := AlbumDTOs{
		makeAlbumTitled("1", "White Noise", "The Beatles"),
		makeAlbumTitled("2", "Kid A", "Radiohead"),
		makeAlbumTitled("3", "OK Computer", "Radiohead"),
	}
	titleSide := albums.Filter(FilterParams{Q: "noise"})
	if len(titleSide) != 1 || titleSide[0].ID != "1" {
		t.Fatalf("title-only match: expected album 1; got %+v", titleSide)
	}
	artistSide := albums.Filter(FilterParams{Q: "head"})
	if len(artistSide) != 2 {
		t.Fatalf("artist-only match: expected the two Radiohead albums; got %+v", artistSide)
	}
}

func TestFilter_Q_NoMatchReturnsEmpty(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbumTitled("1", "Abbey Road", "The Beatles"),
		makeAlbumTitled("2", "Kid A", "Radiohead"),
	}
	result := albums.Filter(FilterParams{Q: "zzzznotpresent"})
	if len(result) != 0 {
		t.Fatalf("expected zero results; got %d", len(result))
	}
}

func TestFilter_Q_AndComposesWithFormat(t *testing.T) {
	now := time.Now()
	albums := AlbumDTOs{
		{
			ID:       "1",
			Title:    "Abbey Road",
			Artists:  []ArtistDTO{{ID: "beatles", Name: "The Beatles"}},
			Releases: ReleaseDTOs{{Format: models.ReleaseFormatVinyl, AddedAt: &now}},
		},
		{
			ID:       "2",
			Title:    "Abbey Road",
			Artists:  []ArtistDTO{{ID: "beatles", Name: "The Beatles"}},
			Releases: ReleaseDTOs{{Format: models.ReleaseFormatDigital, AddedAt: &now}},
		},
	}
	result := albums.Filter(FilterParams{
		Q:       "abbey",
		Formats: []models.ReleaseFormat{models.ReleaseFormatVinyl},
	})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected only the vinyl Abbey Road; got %+v", result)
	}
}

func TestFilter_Q_AndComposesWithRating(t *testing.T) {
	albums := AlbumDTOs{
		makeAlbum("1", "Radiohead Demo", "Radiohead", ptr(9.0), nil),
		makeAlbum("2", "Radiohead B-Sides", "Radiohead", ptr(5.0), nil),
		makeAlbum("3", "Different Band", "Other", ptr(9.5), nil),
	}
	result := albums.Filter(FilterParams{Q: "radio", MinRating: ptr(7.0)})
	if len(result) != 1 || result[0].ID != "1" {
		t.Fatalf("expected only the high-rated Radiohead album; got %+v", result)
	}
}

// --- SaveFormatInput ---

func TestSaveFormatInput_PhysicalFormats(t *testing.T) {
	t.Run("owned input has format set", func(t *testing.T) {
		input := SaveFormatInput{
			Format:    models.ReleaseFormatVinyl,
			Owned:     true,
			DiscogsID: "42",
			Label:     "Blue Note",
		}
		if input.Format != models.ReleaseFormatVinyl {
			t.Errorf("expected vinyl, got %v", input.Format)
		}
		if !input.Owned {
			t.Error("expected Owned = true")
		}
		if input.DiscogsID != "42" {
			t.Errorf("expected DiscogsID %q, got %q", "42", input.DiscogsID)
		}
	})

	t.Run("unowned input with empty discogs", func(t *testing.T) {
		input := SaveFormatInput{
			Format: models.ReleaseFormatCD,
			Owned:  false,
		}
		if input.Owned {
			t.Error("expected Owned = false")
		}
		if input.DiscogsID != "" {
			t.Errorf("expected empty DiscogsID, got %q", input.DiscogsID)
		}
	})
}

func TestMergeDiscoverState(t *testing.T) {
	t.Run("marks unknown albums as 'none' with empty AlbumID", func(t *testing.T) {
		results := []spotifylib.SimpleAlbum{
			simpleAlbumStub("sp-1", "Unknown One", "art-a", "Artist A"),
		}
		out := mergeDiscoverState(results, map[string]UserAlbumStateRow{})
		if len(out) != 1 {
			t.Fatalf("got %d results, want 1", len(out))
		}
		if out[0].State != DiscoverAlbumStateNone {
			t.Errorf("state = %q, want %q", out[0].State, DiscoverAlbumStateNone)
		}
		if out[0].AlbumID != "" {
			t.Errorf("AlbumID = %q, want empty", out[0].AlbumID)
		}
	})

	t.Run("attaches in_library state and AlbumID when known", func(t *testing.T) {
		results := []spotifylib.SimpleAlbum{
			simpleAlbumStub("sp-1", "Known", "art-a", "Artist A"),
		}
		states := map[string]UserAlbumStateRow{
			"sp-1": {AlbumID: "wax-1", State: DiscoverAlbumStateInLibrary},
		}
		out := mergeDiscoverState(results, states)
		if out[0].State != DiscoverAlbumStateInLibrary {
			t.Errorf("state = %q, want in_library", out[0].State)
		}
		if out[0].AlbumID != "wax-1" {
			t.Errorf("AlbumID = %q, want wax-1", out[0].AlbumID)
		}
	})

	t.Run("preserves input order and enriches each result independently", func(t *testing.T) {
		results := []spotifylib.SimpleAlbum{
			simpleAlbumStub("sp-1", "First", "art-a", "Artist A"),
			simpleAlbumStub("sp-2", "Second", "art-b", "Artist B"),
		}
		states := map[string]UserAlbumStateRow{
			"sp-2": {AlbumID: "wax-2", State: DiscoverAlbumStateOnRadar},
		}
		out := mergeDiscoverState(results, states)
		if out[0].SpotifyID != "sp-1" || out[1].SpotifyID != "sp-2" {
			t.Fatalf("order broken: %v", out)
		}
		if out[0].State != DiscoverAlbumStateNone {
			t.Errorf("first state = %q, want none", out[0].State)
		}
		if out[1].State != DiscoverAlbumStateOnRadar || out[1].AlbumID != "wax-2" {
			t.Errorf("second result = %+v, want on_radar/wax-2", out[1])
		}
	})

	t.Run("propagates artist metadata from Spotify result", func(t *testing.T) {
		results := []spotifylib.SimpleAlbum{
			simpleAlbumStub("sp-1", "Album", "art-a", "Artist A"),
		}
		out := mergeDiscoverState(results, map[string]UserAlbumStateRow{})
		if len(out[0].Artists) != 1 {
			t.Fatalf("got %d artists, want 1", len(out[0].Artists))
		}
		if out[0].Artists[0].SpotifyID != "art-a" {
			t.Errorf("artist SpotifyID = %q, want art-a", out[0].Artists[0].SpotifyID)
		}
		if out[0].Artists[0].Name != "Artist A" {
			t.Errorf("artist Name = %q, want Artist A", out[0].Artists[0].Name)
		}
	})

	t.Run("uses first image URL when present", func(t *testing.T) {
		stub := simpleAlbumStub("sp-1", "Album", "art-a", "Artist A")
		stub.Images = []spotifylib.Image{
			{URL: "https://example.test/cover.jpg"},
			{URL: "https://example.test/cover-small.jpg"},
		}
		out := mergeDiscoverState([]spotifylib.SimpleAlbum{stub}, map[string]UserAlbumStateRow{})
		if out[0].ImageURL != "https://example.test/cover.jpg" {
			t.Errorf("ImageURL = %q, want first image URL", out[0].ImageURL)
		}
	})
}

// simpleAlbumStub builds a minimal spotify SimpleAlbum for tests.
func simpleAlbumStub(id, name, artistID, artistName string) spotifylib.SimpleAlbum {
	return spotifylib.SimpleAlbum{
		ID:   spotifylib.ID(id),
		Name: name,
		Artists: []spotifylib.SimpleArtist{
			{ID: spotifylib.ID(artistID), Name: artistName},
		},
	}
}
