package feed

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/alecdray/wax/src/internal/core/sqlx"
	"github.com/alecdray/wax/src/internal/core/utils"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/spotify"

	"github.com/google/uuid"
)

const (
	MinStaleDuration = 1 * time.Hour
)

type FeedDTO struct {
	ID                  string
	UserID              string
	Kind                models.FeedKind
	LastSyncStatus      models.FeedSyncStatus
	LastSyncCompletedAt *time.Time
	LastSyncStartedAt   *time.Time
}

func NewFeedDTOFromModel(model sqlc.Feed) *FeedDTO {
	dto := &FeedDTO{
		ID:             model.ID,
		UserID:         model.UserID,
		Kind:           model.Kind,
		LastSyncStatus: model.LastSyncStatus,
	}

	if model.LastSyncStartedAt.Valid {
		dto.LastSyncStartedAt = &model.LastSyncStartedAt.Time
	}

	if model.LastSyncCompletedAt.Valid {
		dto.LastSyncCompletedAt = &model.LastSyncCompletedAt.Time
	}

	return dto
}

func (f FeedDTO) IsSyncStale() bool {
	if f.LastSyncStatus.IsUnsyned() {
		return false
	}
	if f.LastSyncCompletedAt == nil {
		return true
	}
	minStaleTime := time.Now().Add(-MinStaleDuration)
	return f.LastSyncCompletedAt.Before(minStaleTime)
}

func (f *FeedDTO) SetSyncFailed() {
	f.LastSyncStatus = models.FeedSyncStatusFailure
}

func (f *FeedDTO) SetSyncSuccess() {
	f.LastSyncStatus = models.FeedSyncStatusSuccess
	f.LastSyncCompletedAt = utils.NewPointer(time.Now())
}

func (f *FeedDTO) SetSyncing() {
	f.LastSyncStatus = models.FeedSyncStatusPending
	f.LastSyncStartedAt = utils.NewPointer(time.Now())
}

type Service struct {
	db             *db.DB
	spotifyService *spotify.Service
	libraryService *library.Service
}

func NewService(db *db.DB, spotifyService *spotify.Service, libraryService *library.Service) *Service {
	return &Service{
		db:             db,
		spotifyService: spotifyService,
		libraryService: libraryService,
	}
}

func (s *Service) UpsertFeed(ctx context.Context, userID string, kind models.FeedKind) (*FeedDTO, error) {
	feed, err := s.db.Queries().UpsertFeed(ctx, sqlc.UpsertFeedParams{
		ID:     uuid.New().String(),
		UserID: userID,
		Kind:   kind,
	})
	if err != nil {
		return nil, err
	}
	return NewFeedDTOFromModel(feed), nil
}

func (s *Service) GetFeedByID(ctx context.Context, feedID, userID string) (*FeedDTO, error) {
	feed, err := s.db.Queries().GetFeedByID(ctx, sqlc.GetFeedByIDParams{
		ID:     feedID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return NewFeedDTOFromModel(feed), nil
}

func (s *Service) GetUsersFeeds(ctx context.Context, userID string) ([]FeedDTO, error) {
	feeds, err := s.db.Queries().GetFeedsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	var feedDTOs []FeedDTO
	for _, feed := range feeds {
		feedDTOs = append(feedDTOs, *NewFeedDTOFromModel(feed))
	}

	return feedDTOs, nil
}

func (s *Service) UpdateFeed(ctx context.Context, feed FeedDTO) (*FeedDTO, error) {
	feedModel, err := s.db.Queries().UpdateFeed(ctx, sqlc.UpdateFeedParams{
		ID:                  feed.ID,
		LastSyncStatus:      feed.LastSyncStatus,
		LastSyncStartedAt:   sqlx.NewNullTime(feed.LastSyncStartedAt),
		LastSyncCompletedAt: sqlx.NewNullTime(feed.LastSyncCompletedAt),
	})
	if err != nil {
		return nil, err
	}
	return NewFeedDTOFromModel(feedModel), nil
}

func (s *Service) syncAlbumsToLibrary(ctx contextx.ContextX, feed FeedDTO, syncWindow *time.Duration) error {
	var savedAlbums []spotify.SavedAlbum
	var err error
	if syncWindow == nil {
		savedAlbums, err = s.spotifyService.GetUsersSavedAlbums(ctx, feed.UserID)
		if err != nil {
			err = fmt.Errorf("failed to get user saved albums: %w", err)
			return err
		}
	} else {
		savedAlbums, err = s.spotifyService.GetRecentlySavedAlbums(ctx, feed.UserID, *syncWindow)
		if err != nil {
			err = fmt.Errorf("failed to get user saved albums: %w", err)
			return err
		}
	}

	albumsToSync := make([]library.AlbumDTO, len(savedAlbums))
	for i, album := range savedAlbums {
		var addedAt *time.Time = nil
		_addedAt, err := time.Parse(time.RFC3339, album.AddedAt)
		if err != nil {
			slog.Error("failed to parse added at time during syncSpotifyFeed", err)
		} else {
			addedAt = &_addedAt
		}

		var imageURL string
		if len(album.Images) > 0 {
			imageURL = album.Images[0].URL
		}

		lib := library.AlbumDTO{
			ID:        uuid.NewString(),
			SpotifyID: album.ID.String(),
			Title:     album.Name,
			ImageURL:  imageURL,
			Artists:   make([]library.ArtistDTO, len(album.Artists)),
			Tracks:    []library.TrackDTO{},
			Releases: []library.ReleaseDTO{
				{
					ID:      uuid.NewString(),
					Format:  models.ReleaseFormatDigital,
					AddedAt: addedAt,
				},
			},
		}

		for i, artist := range album.Artists {
			lib.Artists[i] = library.ArtistDTO{
				ID:        uuid.NewString(),
				SpotifyID: artist.ID.String(),
				Name:      artist.Name,
			}
		}

		for _, track := range album.Tracks.Tracks {
			lib.Tracks = append(lib.Tracks, library.TrackDTO{
				ID:        uuid.NewString(),
				SpotifyID: track.ID.String(),
				Title:     track.Name,
			})
		}

		albumsToSync[i] = lib
	}

	err = s.libraryService.AddAlbumsToLibrary(ctx, feed.UserID, albumsToSync)
	if err != nil {
		err = fmt.Errorf("failed to add albums to library: %w", err)
		return err
	}

	return nil
}

func (s *Service) SyncSpotifyFeed(ctx contextx.ContextX, feed FeedDTO) (*FeedDTO, error) {
	if feed.Kind != models.FeedKindSpotify {
		return nil, fmt.Errorf("feed kind must be spotify")
	}

	var syncWindow *time.Duration
	if feed.LastSyncCompletedAt != nil {
		lastSyncedAtPlusBuffer := (*feed.LastSyncCompletedAt).Add(-time.Hour)
		timeSinceLastSync := time.Now().Sub(lastSyncedAtPlusBuffer)
		syncWindow = &timeSinceLastSync
	} else {
		syncWindow = nil
	}

	feed.SetSyncing()
	_, err := s.UpdateFeed(ctx, feed)
	if err != nil {
		err = fmt.Errorf("failed to update feed on sync start: %w", err)
		return nil, err
	}

	err = s.syncAlbumsToLibrary(ctx, feed, syncWindow)
	if err != nil {
		err = fmt.Errorf("failed to sync albums to library: %w", err)

		feed.SetSyncFailed()
		_, updateErr := s.UpdateFeed(ctx, feed)
		if updateErr != nil {
			slog.Error("failed to update feed on sync error", "error", updateErr)
		}

		return nil, err
	}

	feed.SetSyncSuccess()
	_, err = s.UpdateFeed(ctx, feed)
	if err != nil {
		err = fmt.Errorf("failed to update feed on sync success: %w", err)
		return nil, err
	}

	return &feed, nil
}

func (s *Service) GetStaleSpotifyFeeds(ctx context.Context) ([]FeedDTO, error) {
	feeds, err := s.db.Queries().GetStaleFeedsBatch(ctx, sqlc.GetStaleFeedsBatchParams{
		Datetime: sqlx.DurationToSQLiteDatetime(MinStaleDuration),
		Kind:     models.FeedKindSpotify,
	})
	if err != nil {
		return nil, err
	}

	staleFeeds := make([]FeedDTO, 0, len(feeds))
	for _, f := range feeds {
		feed := NewFeedDTOFromModel(f)
		if feed.Kind == models.FeedKindSpotify && feed.IsSyncStale() {
			staleFeeds = append(staleFeeds, *feed)
		}
	}

	return staleFeeds, nil
}
