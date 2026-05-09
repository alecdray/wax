package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/utils"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/listeninghistory"
	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/spotify"
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

type ReleaseDTO struct {
	ID         string
	AlbumID    string
	Format     models.ReleaseFormat
	AddedAt    *time.Time
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
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

var allFormats = []models.ReleaseFormat{
	models.ReleaseFormatDigital,
	models.ReleaseFormatVinyl,
	models.ReleaseFormatCD,
	models.ReleaseFormatCassette,
}

type TrackDTO struct {
	ID        string
	SpotifyID string
	Title     string
}

type ArtistDTO struct {
	ID        string
	SpotifyID string
	Name      string
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

const AlbumsPageSize = 20

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
	MinRating *float64
	MaxRating *float64
	Rated     string // "only" | "unrated" | ""
	Formats   []models.ReleaseFormat
	ArtistIDs []string
}

func (albums AlbumDTOs) Filter(p FilterParams) AlbumDTOs {
	if p.MinRating == nil && p.MaxRating == nil && p.Rated == "" && len(p.Formats) == 0 && len(p.ArtistIDs) == 0 {
		return albums
	}
	result := make(AlbumDTOs, 0, len(albums))
	for _, album := range albums {
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

type Library struct {
	OwnerUserID string
	Albums      AlbumDTOs
	Artists     []ArtistDTO
	Tracks      []TrackDTO
}

func NewLibrary(ownerUserID string, albums []AlbumDTO) *Library {
	lib := &Library{
		OwnerUserID: ownerUserID,
		Albums:      albums,
	}

	lib.Artists = lib.artists()
	lib.Tracks = lib.tracks()

	return lib
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

type Service struct {
	db                      *db.DB
	repo                    *Repo
	spotifyService          *spotify.Service
	listeningHistoryService *listeninghistory.Service
	tagsService             *tags.Service
	notesService            *notes.Service
	reviewService           *review.Service
}

func NewService(d *db.DB, spotifyService *spotify.Service, listeningHistoryService *listeninghistory.Service, tagsService *tags.Service, notesService *notes.Service, reviewService *review.Service) *Service {
	return &Service{
		db:                      d,
		repo:                    NewRepo(d.Queries()),
		spotifyService:          spotifyService,
		listeningHistoryService: listeningHistoryService,
		tagsService:             tagsService,
		notesService:            notesService,
		reviewService:           reviewService,
	}
}

func (s *Service) GetReleasesInLibrary(ctx context.Context, userId string) ([]ReleaseDTO, error) {
	releaseDTOs, err := s.repo.GetUserReleases(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user releases: %w", err)
	}
	return releaseDTOs, nil
}

func (s *Service) GetAlbumsInLibrary(ctx context.Context, userId string) ([]AlbumDTO, error) {
	releases, err := s.GetReleasesInLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get releases: %w", err)
	}

	releasesByAlbumId := make(map[string][]ReleaseDTO, len(releases))
	albumIds := make([]string, 0, len(releases))
	for _, release := range releases {
		albumIds = append(albumIds, release.AlbumID)
		releasesByAlbumId[release.AlbumID] = append(releasesByAlbumId[release.AlbumID], release)
	}

	albums, err := s.repo.GetAlbumsByIDs(ctx, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums: %w", err)
	}

	artistsByAlbumId, err := s.repo.GetArtistsByAlbumIDs(ctx, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album artists: %w", err)
	}

	tracksByAlbumId, err := s.repo.GetTracksByAlbumIDs(ctx, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tracks: %w", err)
	}

	ratingsByAlbumId, err := s.repo.GetLatestUserAlbumRatings(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings: %w", err)
	}

	lastPlayedAtByAlbumId, err := s.listeningHistoryService.GetLastPlayedAtByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get last played at: %w", err)
	}

	tagsByAlbumId, err := s.tagsService.GetAlbumTagsByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}

	notesByAlbumId, err := s.notesService.GetAlbumNotesByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album notes: %w", err)
	}

	ratingStates, err := s.reviewService.GetAllRatingStates(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating states: %w", err)
	}

	var albumDTOs []AlbumDTO
	for _, album := range albums {
		dto := album
		dto.Artists = artistsByAlbumId[album.ID]
		dto.Tracks = tracksByAlbumId[album.ID]
		dto.Releases = releasesByAlbumId[album.ID]
		rating := ratingsByAlbumId[album.ID]
		dto.Rating = utils.NewPointer(rating)
		if t, ok := lastPlayedAtByAlbumId[album.ID]; ok {
			dto.LastPlayedAt = &t
		}
		dto.Tags = tagsByAlbumId[album.ID]
		dto.SleeveNote = notesByAlbumId[album.ID]
		dto.RatingState = ratingStates[album.ID]
		albumDTOs = append(albumDTOs, dto)
	}

	return albumDTOs, nil
}

func (s *Service) GetLibrary(ctx context.Context, userId string) (*Library, error) {
	albums, err := s.GetAlbumsInLibrary(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user albums: %w", err)
	}
	return NewLibrary(userId, albums), nil
}

func (s *Service) AddAlbumsToLibrary(ctx context.Context, userId string, albums []AlbumDTO) error {
	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		for _, album := range albums {
			if _, err := txRepo.AddAlbumToCollection(ctx, userId, album); err != nil {
				return fmt.Errorf("failed to add album to collection: %w", err)
			}
		}
		return nil
	})
}

func (s *Service) GetAlbumInLibrary(ctx context.Context, userId string, albumId string) (*AlbumDTO, error) {
	album, err := s.repo.GetAlbumByID(ctx, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums: %w", err)
	}

	releasesDtos, err := s.repo.GetUserReleasesByAlbumID(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get releases: %w", err)
	}

	if len(releasesDtos) < 1 {
		return nil, errors.New("album not in library")
	}

	artistDtos, err := s.repo.GetArtistsByAlbumID(ctx, album.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album artists: %w", err)
	}

	trackDtos, err := s.repo.GetTracksByAlbumID(ctx, album.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tracks: %w", err)
	}

	ratingDTO, err := s.repo.GetLatestUserAlbumRating(ctx, userId, album.ID)
	if errors.Is(err, sql.ErrNoRows) {
		ratingDTO = nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	ratingLog, err := s.repo.GetUserAlbumRatingLog(ctx, userId, album.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating log: %w", err)
	}

	album.Artists = artistDtos
	album.Tracks = trackDtos
	album.Releases = releasesDtos
	album.Rating = ratingDTO
	album.RatingLog = ratingLog

	albumTags, err := s.tagsService.GetAlbumTags(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}
	album.Tags = albumTags

	sleeveNote, err := s.notesService.GetAlbumNote(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get album note: %w", err)
	}
	album.SleeveNote = sleeveNote

	ratingState, err := s.reviewService.GetRatingState(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating state: %w", err)
	}
	album.RatingState = ratingState

	lastPlayedAtByAlbumId, err := s.listeningHistoryService.GetLastPlayedAtByAlbumIds(ctx, userId, []string{albumId})
	if err != nil {
		return nil, fmt.Errorf("failed to get last played at: %w", err)
	}
	if t, ok := lastPlayedAtByAlbumId[albumId]; ok {
		album.LastPlayedAt = &t
	}

	return album, nil
}

func (s *Service) GetRecentlyPlayedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	dtos, err := s.repo.GetRecentlyPlayedAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently played albums: %w", err)
	}
	return dtos, nil
}

func (s *Service) RemoveAlbumFromLibrary(ctx contextx.ContextX, userId, albumId string) error {
	spotifyID, err := s.repo.GetAlbumSpotifyID(ctx, albumId)
	if err != nil {
		return fmt.Errorf("failed to get album: %w", err)
	}

	if err := s.repo.SoftDeleteUserReleasesByAlbumID(ctx, userId, albumId); err != nil {
		return fmt.Errorf("failed to soft delete releases: %w", err)
	}

	if err := s.spotifyService.RemoveAlbumFromSavedLibrary(ctx, userId, spotifyID); err != nil {
		slog.WarnContext(ctx, "failed to remove album from spotify saved library", "error", err)
	}

	return nil
}

func (s *Service) GetRerateQueue(ctx context.Context, userID string) ([]RerateAlbumDTO, error) {
	dtos, err := s.repo.GetRerateQueue(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rerate queue: %w", err)
	}
	return dtos, nil
}

func (s *Service) GetAlbumFormats(ctx context.Context, userID, albumID string) ([]AlbumFormatDTO, error) {
	formats, err := s.repo.GetAlbumFormats(ctx, userID, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album formats: %w", err)
	}
	return formats, nil
}

type SaveFormatInput struct {
	Format     models.ReleaseFormat
	Owned      bool
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}

func (s *Service) SaveAlbumFormats(ctx context.Context, userID, albumID string, inputs []SaveFormatInput) error {
	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())

		currentFormats, err := txRepo.GetAlbumFormats(ctx, userID, albumID)
		if err != nil {
			return fmt.Errorf("failed to get current formats: %w", err)
		}
		currentByFormat := make(map[models.ReleaseFormat]AlbumFormatDTO, len(currentFormats))
		for _, f := range currentFormats {
			currentByFormat[f.Format] = f
		}

		for _, input := range inputs {
			if input.Format == models.ReleaseFormatDigital {
				continue // digital is managed by Spotify, never modified here
			}

			current := currentByFormat[input.Format]
			releaseID := current.ReleaseID

			if input.Owned {
				if !current.Owned {
					newReleaseID, err := txRepo.AddOwnedRelease(ctx, userID, albumID, input.Format, releaseID, time.Now())
					if err != nil {
						return fmt.Errorf("failed to add owned release: %w", err)
					}
					releaseID = newReleaseID
				}

				if releaseID != "" && input.DiscogsID != "" {
					if err := txRepo.UpdateReleaseDiscogsInfo(ctx, releaseID, input.DiscogsID, input.Label, input.ReleasedAt); err != nil {
						return fmt.Errorf("failed to update release discogs info: %w", err)
					}
				} else if releaseID != "" && input.DiscogsID == "" && current.DiscogsID != "" {
					if err := txRepo.ClearReleaseDiscogsInfo(ctx, releaseID); err != nil {
						return fmt.Errorf("failed to clear release discogs info: %w", err)
					}
				}
			} else if current.Owned && releaseID != "" {
				if err := txRepo.SoftDeleteUserRelease(ctx, userID, releaseID); err != nil {
					return fmt.Errorf("failed to soft delete user release: %w", err)
				}
			}
		}
		return nil
	})
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

func (s *Service) GetUnratedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	dtos, err := s.repo.GetUnratedAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unrated albums: %w", err)
	}
	return dtos, nil
}
