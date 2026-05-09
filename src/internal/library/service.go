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

func (s *Service) GetUnratedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	dtos, err := s.repo.GetUnratedAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unrated albums: %w", err)
	}
	return dtos, nil
}
