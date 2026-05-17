package library

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/tags"
)

const AlbumsPageSize = 20

type UserReleaseStatus string

const (
	UserReleaseStatusWishlist UserReleaseStatus = "wishlist"
	UserReleaseStatusOwned    UserReleaseStatus = "owned"
	UserReleaseStatusRemoved  UserReleaseStatus = "removed"
)

// ---------- Album aggregate types ----------

type AlbumSummaryDTO struct {
	ID        string
	SpotifyID string
	Title     string
	Artists   string
	ImageURL  string
	InLibrary bool
	OnRadar   bool
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

type ProvisionalAlbumDTO struct {
	ID          string
	SpotifyID   string
	Title       string
	Artists     string
	ImageURL    string
	Rating      *float64
	RatingState review.RatingState
}

// ---------- Releases & formats ----------

type ReleaseDTO struct {
	ID              string
	AlbumID         string
	Format          models.ReleaseFormat
	Status          UserReleaseStatus
	AddedAt         *time.Time   // alias of StatusUpdatedAt while Status == "owned"; kept for existing UI callers
	CreatedAt       *time.Time
	StatusUpdatedAt *time.Time
	DiscogsID       string
	Label           string
	ReleasedAt      *time.Time
}

type ReleaseDTOs []ReleaseDTO

func (releases ReleaseDTOs) OldestAddedAtDate() *time.Time {
	var oldest *time.Time
	for _, r := range releases {
		if r.AddedAt != nil {
			if oldest == nil || r.AddedAt.Before(*oldest) {
				oldest = r.AddedAt
			}
		}
	}
	return oldest
}

func (releases ReleaseDTOs) FindFormat(format models.ReleaseFormat) *ReleaseDTO {
	for _, r := range releases {
		if r.Format == format {
			return &r
		}
	}
	return nil
}

type RadarDTO struct {
	AlbumID   string
	CreatedAt time.Time
}

// AlbumFormatDTO represents one format row in the formats modal.
// It exists for all 4 formats regardless of whether the user owns that format.
type AlbumFormatDTO struct {
	Format     models.ReleaseFormat
	ReleaseID  string // empty if this format has never been added for this album
	Owned      bool
	AddedAt    *time.Time
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}

type SaveFormatInput struct {
	Format     models.ReleaseFormat
	Owned      bool
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}

var allFormats = []models.ReleaseFormat{
	models.ReleaseFormatDigital,
	models.ReleaseFormatVinyl,
	models.ReleaseFormatCD,
	models.ReleaseFormatCassette,
}

// PhysicalFormats lists the physical release formats wax tracks for a library album.
var PhysicalFormats = []models.ReleaseFormat{
	models.ReleaseFormatVinyl,
	models.ReleaseFormatCD,
	models.ReleaseFormatCassette,
}

// IsPhysicalFormat reports whether the given format is one wax tracks as a physical release.
func IsPhysicalFormat(f models.ReleaseFormat) bool {
	for _, pf := range PhysicalFormats {
		if pf == f {
			return true
		}
	}
	return false
}

// NormalizeDiscogsReleasedDate produces a YYYY-MM-DD date string from a Discogs Release plus
// fallback values. Discogs's Release.Released can be "YYYY-MM-DD", "YYYY", or empty; this
// function picks the most precise option and pads bare years to YYYY-01-01. release may be nil.
func NormalizeDiscogsReleasedDate(release *discogs.Release, fallbackYear string) string {
	releasedDate := fallbackYear
	if release != nil {
		if len(release.Released) >= 10 {
			releasedDate = release.Released
		} else if len(release.Released) >= 4 {
			releasedDate = release.Released[:4] + "-01-01"
		} else if release.Year > 0 {
			releasedDate = strconv.Itoa(release.Year) + "-01-01"
		}
	}
	if len(releasedDate) == 4 {
		releasedDate += "-01-01"
	}
	return releasedDate
}

// ---------- AlbumDTOs slice: pagination, sorting, filtering ----------

type AlbumDTOs []AlbumDTO

func (albums AlbumDTOs) Page(offset int) AlbumDTOs {
	if offset >= len(albums) {
		return nil
	}
	end := offset + AlbumsPageSize
	if end > len(albums) {
		end = len(albums)
	}
	return albums[offset:end]
}

func (albums AlbumDTOs) SortByTitle(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		if ascending {
			return albums[i].Title < albums[j].Title
		}
		return albums[i].Title > albums[j].Title
	})
}

func (albums AlbumDTOs) SortByArtist(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		if len(albums[i].Artists) == 0 && len(albums[j].Artists) == 0 {
			return false
		}
		if len(albums[i].Artists) == 0 {
			return ascending
		}
		if len(albums[j].Artists) == 0 {
			return !ascending
		}
		if ascending {
			return albums[i].Artists[0].Name < albums[j].Artists[0].Name
		}
		return albums[i].Artists[0].Name > albums[j].Artists[0].Name
	})
}

func (albums AlbumDTOs) SortByRating(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		var ratingI, ratingJ *float64
		if albums[i].Rating != nil {
			ratingI = albums[i].Rating.Rating
		}
		if albums[j].Rating != nil {
			ratingJ = albums[j].Rating.Rating
		}
		if ratingI == nil && ratingJ == nil {
			return false
		}
		if ratingI == nil {
			return ascending
		}
		if ratingJ == nil {
			return !ascending
		}
		if ascending {
			return *ratingI < *ratingJ
		}
		return *ratingI > *ratingJ
	})
}

func (albums AlbumDTOs) SortByLastPlayed(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		if albums[i].LastPlayedAt == nil && albums[j].LastPlayedAt == nil {
			return false
		}
		if albums[i].LastPlayedAt == nil {
			return ascending
		}
		if albums[j].LastPlayedAt == nil {
			return !ascending
		}
		if ascending {
			return albums[i].LastPlayedAt.Before(*albums[j].LastPlayedAt)
		}
		return albums[i].LastPlayedAt.After(*albums[j].LastPlayedAt)
	})
}

func (albums AlbumDTOs) SortByDate(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		dateI := albums[i].Releases.OldestAddedAtDate()
		dateJ := albums[j].Releases.OldestAddedAtDate()
		if dateI == nil && dateJ == nil {
			return false
		}
		if dateI == nil {
			return ascending
		}
		if dateJ == nil {
			return !ascending
		}
		if ascending {
			return dateI.Before(*dateJ)
		}
		return dateI.After(*dateJ)
	})
}

type FilterParams struct {
	Q         string // case-insensitive substring against album title and credited artist names
	MinRating *float64
	MaxRating *float64
	Rated     string // "only" | "unrated" | ""
	Formats   []models.ReleaseFormat
	ArtistIDs []string
}

// matchesQ returns true if the album's title or any credited artist's name
// contains q as a case-insensitive substring. Empty q matches everything.
func matchesQ(album AlbumDTO, q string) bool {
	if q == "" {
		return true
	}
	needle := strings.ToLower(q)
	if strings.Contains(strings.ToLower(album.Title), needle) {
		return true
	}
	for _, artist := range album.Artists {
		if strings.Contains(strings.ToLower(artist.Name), needle) {
			return true
		}
	}
	return false
}

func (albums AlbumDTOs) Filter(p FilterParams) AlbumDTOs {
	q := strings.TrimSpace(p.Q)
	if q == "" && p.MinRating == nil && p.MaxRating == nil && p.Rated == "" && len(p.Formats) == 0 && len(p.ArtistIDs) == 0 {
		return albums
	}
	result := make(AlbumDTOs, 0, len(albums))
	for _, album := range albums {
		if !matchesQ(album, q) {
			continue
		}
		if p.MinRating != nil {
			if album.Rating == nil || album.Rating.Rating == nil || *album.Rating.Rating < *p.MinRating {
				continue
			}
		}
		if p.MaxRating != nil {
			if album.Rating == nil || album.Rating.Rating == nil || *album.Rating.Rating > *p.MaxRating {
				continue
			}
		}
		switch p.Rated {
		case "only":
			if album.Rating == nil || album.Rating.Rating == nil {
				continue
			}
		case "unrated":
			if album.Rating != nil && album.Rating.Rating != nil {
				continue
			}
		}
		if len(p.Formats) > 0 {
			hasFormat := false
			for _, format := range p.Formats {
				if release := album.Releases.FindFormat(format); release != nil && release.AddedAt != nil {
					hasFormat = true
					break
				}
			}
			if !hasFormat {
				continue
			}
		}
		if len(p.ArtistIDs) > 0 {
			hasArtist := false
		outer:
			for _, artistID := range p.ArtistIDs {
				for _, artist := range album.Artists {
					if artist.ID == artistID {
						hasArtist = true
						break outer
					}
				}
			}
			if !hasArtist {
				continue
			}
		}
		result = append(result, album)
	}
	return result
}

// ---------- Library aggregate (dashboard view) ----------

// Library is the aggregate the library dashboard view binds to: a user's
// albums plus derived collections (unique artists and tracks across those albums).
type Library struct {
	OwnerUserID string
	Albums      AlbumDTOs
	Artists     []ArtistDTO
	Tracks      []TrackDTO
}

func NewLibrary(ownerUserID string, albums []AlbumDTO) *Library {
	l := &Library{
		OwnerUserID: ownerUserID,
		Albums:      albums,
	}

	l.Artists = l.artists()
	l.Tracks = l.tracks()

	return l
}

func (l *Library) artists() []ArtistDTO {
	artistsSet := make(map[string]ArtistDTO)
	for _, album := range l.Albums {
		for _, artist := range album.Artists {
			artistsSet[artist.ID] = artist
		}
	}

	artists := make([]ArtistDTO, 0, len(artistsSet))
	for _, artist := range artistsSet {
		artists = append(artists, artist)
	}

	return artists
}

func (l *Library) tracks() []TrackDTO {
	tracksSet := make(map[string]TrackDTO)
	for _, album := range l.Albums {
		for _, track := range album.Tracks {
			tracksSet[track.ID] = track
		}
	}

	tracks := make([]TrackDTO, 0, len(tracksSet))
	for _, track := range tracksSet {
		tracks = append(tracks, track)
	}

	return tracks
}

// DiscoverAlbumState describes whether the caller already has a relationship
// with an album (used to render Spotify search results in /discover).
type DiscoverAlbumState string

const (
	DiscoverAlbumStateNone      DiscoverAlbumState = "none"
	DiscoverAlbumStateInLibrary DiscoverAlbumState = "in_library"
	DiscoverAlbumStateOnRadar   DiscoverAlbumState = "on_radar"
	DiscoverAlbumStateRemoved   DiscoverAlbumState = "removed"
)

// UserAlbumStateRow is the per-album result of GetUserAlbumStateBySpotifyIDs.
// AlbumID is the wax album row's primary key (always populated when present
// in the map).
type UserAlbumStateRow struct {
	AlbumID string
	State   DiscoverAlbumState
}

// DiscoverResultDTO is one row in the /discover page's search results.
// AlbumID is empty when State == "none" (the album has no wax row yet).
type DiscoverResultDTO struct {
	SpotifyID string
	Title     string
	Artists   []ArtistDTO
	ImageURL  string
	State     DiscoverAlbumState
	AlbumID   string
}
