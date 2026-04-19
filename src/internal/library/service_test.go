package library

import (
	"database/sql"
	"testing"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/alecdray/wax/src/internal/review"
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

// --- NewReleaseDTOFromModel ---

func TestNewReleaseDTOFromModel_MapsDiscogsFields(t *testing.T) {
	now := time.Now()
	release := sqlc.Release{
		ID:        "r1",
		AlbumID:   "a1",
		Format:    models.ReleaseFormatVinyl,
		CreatedAt: now,
		DiscogsID: sql.NullString{String: "12345", Valid: true},
		Label:     sql.NullString{String: "Warner Bros.", Valid: true},
		ReleasedAt: sql.NullTime{Time: now, Valid: true},
	}
	userRelease := &sqlc.UserRelease{AddedAt: now}

	dto := NewReleaseDTOFromModel(release, userRelease)

	if dto.DiscogsID != "12345" {
		t.Errorf("expected DiscogsID %q, got %q", "12345", dto.DiscogsID)
	}
	if dto.Label != "Warner Bros." {
		t.Errorf("expected Label %q, got %q", "Warner Bros.", dto.Label)
	}
	if dto.ReleasedAt == nil || !dto.ReleasedAt.Equal(now) {
		t.Errorf("expected ReleasedAt %v, got %v", now, dto.ReleasedAt)
	}
}

// --- AlbumFormatDTO helpers ---

func TestAlbumFormatDTO_OwnedFormat_HasDiscogsData(t *testing.T) {
	now := time.Now()
	release := sqlc.Release{
		ID:         "r1",
		AlbumID:    "a1",
		Format:     models.ReleaseFormatVinyl,
		DiscogsID:  sql.NullString{String: "99", Valid: true},
		Label:      sql.NullString{String: "ECM", Valid: true},
		ReleasedAt: sql.NullTime{Time: now, Valid: true},
	}
	userRelease := sqlc.UserRelease{AddedAt: now}

	dto := albumFormatDTOFromRelease(release, &userRelease)

	t.Run("owned is true", func(t *testing.T) {
		if !dto.Owned {
			t.Error("expected Owned = true")
		}
	})
	t.Run("release ID set", func(t *testing.T) {
		if dto.ReleaseID != "r1" {
			t.Errorf("expected ReleaseID %q, got %q", "r1", dto.ReleaseID)
		}
	})
	t.Run("discogs ID mapped", func(t *testing.T) {
		if dto.DiscogsID != "99" {
			t.Errorf("expected DiscogsID %q, got %q", "99", dto.DiscogsID)
		}
	})
	t.Run("label mapped", func(t *testing.T) {
		if dto.Label != "ECM" {
			t.Errorf("expected Label %q, got %q", "ECM", dto.Label)
		}
	})
	t.Run("released_at mapped", func(t *testing.T) {
		if dto.ReleasedAt == nil || !dto.ReleasedAt.Equal(now) {
			t.Errorf("expected ReleasedAt %v, got %v", now, dto.ReleasedAt)
		}
	})
}

func TestAlbumFormatDTO_UnownedReleaseExists(t *testing.T) {
	release := sqlc.Release{
		ID:      "r2",
		AlbumID: "a1",
		Format:  models.ReleaseFormatCD,
	}

	dto := albumFormatDTOFromRelease(release, nil)

	t.Run("owned is false", func(t *testing.T) {
		if dto.Owned {
			t.Error("expected Owned = false")
		}
	})
	t.Run("release ID still set", func(t *testing.T) {
		if dto.ReleaseID != "r2" {
			t.Errorf("expected ReleaseID %q, got %q", "r2", dto.ReleaseID)
		}
	})
}

// --- buildAlbumFormats ---

func TestBuildAlbumFormats(t *testing.T) {
	now := time.Now()

	t.Run("owned format is marked owned with Discogs data", func(t *testing.T) {
		releases := []sqlc.Release{
			{ID: "r-vinyl", AlbumID: "a1", Format: models.ReleaseFormatVinyl,
				DiscogsID: sql.NullString{String: "42", Valid: true},
				Label:     sql.NullString{String: "ECM", Valid: true},
				ReleasedAt: sql.NullTime{Time: now, Valid: true},
			},
		}
		userReleases := []sqlc.GetUserReleasesByAlbumIdRow{
			{Release: releases[0], UserRelease: sqlc.UserRelease{AddedAt: now}},
		}

		result := buildAlbumFormats(releases, userReleases)

		var vinyl *AlbumFormatDTO
		for i := range result {
			if result[i].Format == models.ReleaseFormatVinyl {
				vinyl = &result[i]
			}
		}
		if vinyl == nil {
			t.Fatal("vinyl format not found in result")
		}
		if !vinyl.Owned {
			t.Error("expected vinyl to be owned")
		}
		if vinyl.ReleaseID != "r-vinyl" {
			t.Errorf("expected ReleaseID %q, got %q", "r-vinyl", vinyl.ReleaseID)
		}
		if vinyl.DiscogsID != "42" {
			t.Errorf("expected DiscogsID %q, got %q", "42", vinyl.DiscogsID)
		}
		if vinyl.Label != "ECM" {
			t.Errorf("expected Label %q, got %q", "ECM", vinyl.Label)
		}
	})

	t.Run("release exists but not owned has ReleaseID and Owned=false", func(t *testing.T) {
		releases := []sqlc.Release{
			{ID: "r-cd", AlbumID: "a1", Format: models.ReleaseFormatCD},
		}
		result := buildAlbumFormats(releases, nil)

		var cd *AlbumFormatDTO
		for i := range result {
			if result[i].Format == models.ReleaseFormatCD {
				cd = &result[i]
			}
		}
		if cd == nil {
			t.Fatal("CD format not found in result")
		}
		if cd.Owned {
			t.Error("expected Owned = false")
		}
		if cd.ReleaseID != "r-cd" {
			t.Errorf("expected ReleaseID %q, got %q", "r-cd", cd.ReleaseID)
		}
	})

	t.Run("format with no release row gets empty DTO placeholder", func(t *testing.T) {
		result := buildAlbumFormats(nil, nil)

		if len(result) != len(allFormats) {
			t.Fatalf("expected %d formats, got %d", len(allFormats), len(result))
		}
		for _, dto := range result {
			if dto.Owned {
				t.Errorf("expected Owned = false for format %v with no release", dto.Format)
			}
			if dto.ReleaseID != "" {
				t.Errorf("expected empty ReleaseID for format %v with no release", dto.Format)
			}
		}
	})

	t.Run("result always has all four formats in order", func(t *testing.T) {
		result := buildAlbumFormats(nil, nil)

		if len(result) != 4 {
			t.Fatalf("expected 4 formats, got %d", len(result))
		}
		expected := []models.ReleaseFormat{
			models.ReleaseFormatDigital,
			models.ReleaseFormatVinyl,
			models.ReleaseFormatCD,
			models.ReleaseFormatCassette,
		}
		for i, f := range expected {
			if result[i].Format != f {
				t.Errorf("position %d: expected %v, got %v", i, f, result[i].Format)
			}
		}
	})
}

func TestNewReleaseDTOFromModel_NullDiscogsFieldsAreEmpty(t *testing.T) {
	release := sqlc.Release{
		ID:      "r1",
		AlbumID: "a1",
		Format:  models.ReleaseFormatVinyl,
	}

	dto := NewReleaseDTOFromModel(release, nil)

	if dto.DiscogsID != "" {
		t.Errorf("expected empty DiscogsID, got %q", dto.DiscogsID)
	}
	if dto.Label != "" {
		t.Errorf("expected empty Label, got %q", dto.Label)
	}
	if dto.ReleasedAt != nil {
		t.Errorf("expected nil ReleasedAt, got %v", dto.ReleasedAt)
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
