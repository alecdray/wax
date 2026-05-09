package library

import (
	"time"

	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/tags"
)

type AlbumSummaryDTO struct {
	ID        string
	SpotifyID string
	Title     string
	Artists   string
	ImageURL  string
	InLibrary bool
}

type ArtistDTO struct {
	ID        string
	SpotifyID string
	Name      string
}

type TrackDTO struct {
	ID        string
	SpotifyID string
	Title     string
}

type AlbumDTO struct {
	ID           string
	SpotifyID    string
	Title        string
	ImageURL     string
	Artists      []ArtistDTO
	Tracks       []TrackDTO
	Releases     ReleaseDTOs
	Rating       *review.AlbumRatingDTO
	RatingLog    []*review.AlbumRatingDTO
	Tags         []tags.TagDTO
	SleeveNote   *notes.AlbumNoteDTO
	LastPlayedAt *time.Time
	RatingState  *review.RatingStateDTO
}

type RerateAlbumDTO struct {
	ID          string
	SpotifyID   string
	Title       string
	Artists     string
	ImageURL    string
	Rating      *float64
	RatingState review.RatingState
}
