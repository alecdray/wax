package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/utils"
	"github.com/alecdray/wax/src/internal/listeninghistory"
	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/spotify"
	"github.com/alecdray/wax/src/internal/tags"
	"github.com/google/uuid"
	spotifylib "github.com/zmb3/spotify/v2"
)

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

	ratingsByAlbumId, err := s.reviewService.GetLatestRatings(ctx, userId)
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
			if err := txRepo.RemoveAlbumFromRadar(ctx, userId, album.ID); err != nil {
				return fmt.Errorf("failed to clear radar: %w", err)
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

	ratingDTO, err := s.reviewService.GetLatestRating(ctx, userId, album.ID)
	if errors.Is(err, sql.ErrNoRows) {
		ratingDTO = nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	ratingLog, err := s.reviewService.GetRatingLog(ctx, userId, album.ID)
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

	if err := s.repo.MarkReleasesRemovedByAlbumID(ctx, userId, albumId); err != nil {
		return fmt.Errorf("failed to mark releases removed: %w", err)
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
					if err := txRepo.RemoveAlbumFromRadar(ctx, userID, albumID); err != nil {
						return fmt.Errorf("failed to clear radar: %w", err)
					}
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
				if err := txRepo.MarkReleaseRemoved(ctx, userID, releaseID); err != nil {
					return fmt.Errorf("failed to mark release removed: %w", err)
				}
			}
		}
		return nil
	})
}

func (s *Service) GetUnratedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	dtos, err := s.repo.GetUnratedAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unrated albums: %w", err)
	}
	return dtos, nil
}

// AcquireFromWishlist transitions a user's wishlist release to 'owned' and clears
// the album from the user's radar in a single tx.
func (s *Service) AcquireFromWishlist(ctx context.Context, userID, albumID, releaseID string) error {
	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		if err := txRepo.MarkReleaseOwned(ctx, userID, releaseID); err != nil {
			return fmt.Errorf("failed to acquire from wishlist: %w", err)
		}
		if err := txRepo.RemoveAlbumFromRadar(ctx, userID, albumID); err != nil {
			return fmt.Errorf("failed to clear radar: %w", err)
		}
		return nil
	})
}

// AddReleaseToWishlist upserts a wishlist release for the user and clears the
// album from the user's radar in a single tx. If releaseID is empty, a new
// release row is created for (album, format).
func (s *Service) AddReleaseToWishlist(ctx context.Context, userID, albumID string, format models.ReleaseFormat, releaseID string) (string, error) {
	var resultID string
	err := s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		newID, err := txRepo.AddReleaseToWishlist(ctx, userID, albumID, format, releaseID)
		if err != nil {
			return fmt.Errorf("failed to add release to wishlist: %w", err)
		}
		if err := txRepo.RemoveAlbumFromRadar(ctx, userID, albumID); err != nil {
			return fmt.Errorf("failed to clear radar: %w", err)
		}
		resultID = newID
		return nil
	})
	return resultID, err
}

// ErrAlbumAlreadyDecided is returned by AddAlbumToRadar when the user has any
// user_release row (owned, wishlist, or removed) for the album. Radar is
// strictly pre-decision, so any existing decision disqualifies the album.
var ErrAlbumAlreadyDecided = errors.New("album already has a user release; cannot add to radar")

// AddAlbumToRadar adds an album to the user's radar (pre-decision queue).
// Refuses if the album already has any user_release row.
func (s *Service) AddAlbumToRadar(ctx context.Context, userID, albumID string) error {
	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		hasRelease, err := txRepo.HasAnyUserReleaseForAlbum(ctx, userID, albumID)
		if err != nil {
			return fmt.Errorf("failed to check user releases: %w", err)
		}
		if hasRelease {
			return ErrAlbumAlreadyDecided
		}
		if err := txRepo.AddAlbumToRadar(ctx, userID, albumID); err != nil {
			return fmt.Errorf("failed to add album to radar: %w", err)
		}
		return nil
	})
}

// GetRadarAlbums returns the caller's radar entries as fully-populated
// AlbumDTOs (artists set; tracks/releases left empty — radar entries have no
// release rows). Used by the discover page's radar carousel.
func (s *Service) GetRadarAlbums(ctx context.Context, userID string) ([]AlbumDTO, error) {
	_, albums, err := s.repo.GetRadarAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get radar albums: %w", err)
	}
	if len(albums) == 0 {
		return nil, nil
	}
	albumIDs := make([]string, len(albums))
	for i, a := range albums {
		albumIDs[i] = a.ID
	}
	artistsByAlbumID, err := s.repo.GetArtistsByAlbumIDs(ctx, albumIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get artists for radar albums: %w", err)
	}
	for i := range albums {
		albums[i].Artists = artistsByAlbumID[albums[i].ID]
	}
	return albums, nil
}

// RemoveAlbumFromRadar deletes the radar row. No-op if the user has no radar
// row for the album.
func (s *Service) RemoveAlbumFromRadar(ctx context.Context, userID, albumID string) error {
	return s.repo.RemoveAlbumFromRadar(ctx, userID, albumID)
}

// spotifyAlbumToDTO converts a Spotify FullAlbum into an AlbumDTO ready for
// EnsureAlbumWithMetadata. Releases is intentionally omitted — radar entries
// are pre-decision; the feed sync that populates Releases lives in
// feed/service.go and writes user_releases rows that radar must not create.
func spotifyAlbumToDTO(album *spotifylib.FullAlbum) AlbumDTO {
	var imageURL string
	if len(album.Images) > 0 {
		imageURL = album.Images[0].URL
	}
	dto := AlbumDTO{
		ID:        uuid.NewString(),
		SpotifyID: album.ID.String(),
		Title:     album.Name,
		ImageURL:  imageURL,
	}
	dto.Artists = make([]ArtistDTO, len(album.Artists))
	for i, a := range album.Artists {
		dto.Artists[i] = ArtistDTO{
			ID:        uuid.NewString(),
			SpotifyID: a.ID.String(),
			Name:      a.Name,
		}
	}
	dto.Tracks = make([]TrackDTO, 0, len(album.Tracks.Tracks))
	for _, t := range album.Tracks.Tracks {
		dto.Tracks = append(dto.Tracks, TrackDTO{
			ID:        uuid.NewString(),
			SpotifyID: t.ID.String(),
			Title:     t.Name,
		})
	}
	return dto
}

// AddSpotifyAlbumToRadar imports a Spotify album's metadata (album, artists,
// tracks) into wax and adds the album to the user's radar. Refuses with
// ErrAlbumAlreadyDecided if the album already has any user_releases row.
func (s *Service) AddSpotifyAlbumToRadar(ctx contextx.ContextX, userID, spotifyID string) error {
	spotifyAlbum, err := s.spotifyService.GetFullAlbum(ctx, userID, spotifyID)
	if err != nil {
		return fmt.Errorf("failed to fetch spotify album: %w", err)
	}
	dto := spotifyAlbumToDTO(spotifyAlbum)

	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		imported, err := txRepo.EnsureAlbumWithMetadata(ctx, dto)
		if err != nil {
			return fmt.Errorf("failed to import album metadata: %w", err)
		}
		hasRelease, err := txRepo.HasAnyUserReleaseForAlbum(ctx, userID, imported.ID)
		if err != nil {
			return fmt.Errorf("failed to check user releases: %w", err)
		}
		if hasRelease {
			return ErrAlbumAlreadyDecided
		}
		if err := txRepo.AddAlbumToRadar(ctx, userID, imported.ID); err != nil {
			return fmt.Errorf("failed to add album to radar: %w", err)
		}
		return nil
	})
}
