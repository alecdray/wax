package listeninghistory

import (
	"context"
	"fmt"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/spotify"

	"github.com/google/uuid"
	spotifylib "github.com/zmb3/spotify/v2"
)

type Service struct {
	db             *db.DB
	repo           *Repo
	spotifyService *spotify.Service
}

func NewService(d *db.DB, spotifyService *spotify.Service) *Service {
	return &Service{
		db:             d,
		repo:           NewRepo(d.Queries()),
		spotifyService: spotifyService,
	}
}

func (s *Service) upsertPlayHistory(ctx context.Context, userID string, items []spotifylib.RecentlyPlayedItem) error {
	for _, item := range items {
		track := item.Track
		album := track.Album

		albumImageURL := ""
		if len(album.Images) > 0 {
			albumImageURL = album.Images[0].URL
		}

		albumID, err := s.repo.GetOrCreateAlbum(ctx, AlbumInput{
			ID:        uuid.NewString(),
			SpotifyID: album.ID.String(),
			Title:     album.Name,
			ImageURL:  albumImageURL,
		})
		if err != nil {
			return fmt.Errorf("failed to get/create album %s: %w", album.ID, err)
		}

		trackID, err := s.repo.GetOrCreateTrack(ctx, TrackInput{
			ID:        uuid.NewString(),
			SpotifyID: track.ID.String(),
			Title:     track.Name,
		})
		if err != nil {
			return fmt.Errorf("failed to get/create track %s: %w", track.ID, err)
		}

		if err := s.repo.GetOrCreateAlbumTrack(ctx, albumID, trackID); err != nil {
			return fmt.Errorf("failed to get/create album track: %w", err)
		}

		for _, a := range album.Artists {
			artistID, err := s.repo.GetOrCreateArtist(ctx, ArtistInput{
				ID:        uuid.NewString(),
				SpotifyID: a.ID.String(),
				Name:      a.Name,
			})
			if err != nil {
				return fmt.Errorf("failed to get/create artist %s: %w", a.ID, err)
			}
			if err := s.repo.GetOrCreateAlbumArtist(ctx, albumID, artistID); err != nil {
				return fmt.Errorf("failed to get/create album artist: %w", err)
			}
		}

		err = s.repo.UpsertTrackPlay(ctx, TrackPlayInput{
			ID:       uuid.NewString(),
			UserID:   userID,
			TrackID:  trackID,
			AlbumID:  albumID,
			PlayedAt: item.PlayedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert track play: %w", err)
		}
	}
	return nil
}

func (s *Service) GetLastPlayedAtByAlbumIds(ctx context.Context, userID string, albumIDs []string) (map[string]time.Time, error) {
	result, err := s.repo.GetLastPlayedAtByAlbumIDs(ctx, userID, albumIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get last played at: %w", err)
	}
	return result, nil
}

func (s *Service) GetUserIDsWithSpotifyToken(ctx context.Context) ([]string, error) {
	return s.repo.GetUserIDsWithSpotifyToken(ctx)
}

func (s *Service) SyncUser(ctx contextx.ContextX, userID string) error {
	items, err := s.spotifyService.GetRecentlyPlayedTracks(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get recently played tracks: %w", err)
	}

	if len(items) > 0 {
		if err := s.upsertPlayHistory(ctx, userID, items); err != nil {
			return err
		}
	}

	return nil
}
