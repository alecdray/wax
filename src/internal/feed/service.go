package feed

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/spotify"

	"github.com/google/uuid"
)

type Service struct {
	db             *db.DB
	repo           *Repo
	spotifyService *spotify.Service
	libraryService *library.Service
}

func NewService(d *db.DB, spotifyService *spotify.Service, libraryService *library.Service) *Service {
	return &Service{
		db:             d,
		repo:           NewRepo(d.Queries()),
		spotifyService: spotifyService,
		libraryService: libraryService,
	}
}

func (s *Service) UpsertFeed(ctx context.Context, userID string, kind models.FeedKind) (*FeedDTO, error) {
	return s.repo.UpsertFeed(ctx, userID, kind)
}

func (s *Service) GetFeedByID(ctx context.Context, feedID, userID string) (*FeedDTO, error) {
	return s.repo.GetFeedByID(ctx, feedID, userID)
}

func (s *Service) GetUsersFeeds(ctx context.Context, userID string) ([]FeedDTO, error) {
	return s.repo.GetFeedsByUserID(ctx, userID)
}

func (s *Service) UpdateFeed(ctx context.Context, feed FeedDTO) (*FeedDTO, error) {
	return s.repo.UpdateFeed(ctx, feed)
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
			slog.Error("failed to parse added at time during syncSpotifyFeed", "error", err)
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
	feeds, err := s.repo.GetStaleFeedsBatch(ctx, models.FeedKindSpotify, MinStaleDuration)
	if err != nil {
		return nil, err
	}

	staleFeeds := make([]FeedDTO, 0, len(feeds))
	for _, f := range feeds {
		if f.Kind == models.FeedKindSpotify && f.IsSyncStale() {
			staleFeeds = append(staleFeeds, f)
		}
	}

	return staleFeeds, nil
}
