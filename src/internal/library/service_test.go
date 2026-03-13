package library

import (
	"testing"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/review"
)

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
