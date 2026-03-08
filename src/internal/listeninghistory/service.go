package listeninghistory

import (
	"context"
	"database/sql"
	"fmt"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/db/sqlc"
	"shmoopicks/src/internal/spotify"
	"time"

	spotifylib "github.com/zmb3/spotify/v2"

	"github.com/google/uuid"
)

type Service struct {
	db             *db.DB
	spotifyService *spotify.Service
}

func NewService(db *db.DB, spotifyService *spotify.Service) *Service {
	return &Service{
		db:             db,
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

		albumModel, err := s.db.Queries().GetOrCreateAlbum(ctx, sqlc.GetOrCreateAlbumParams{
			ID:        uuid.NewString(),
			SpotifyID: album.ID.String(),
			Title:     album.Name,
			ImageUrl:  sql.NullString{String: albumImageURL, Valid: albumImageURL != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to get/create album %s: %w", album.ID, err)
		}

		trackModel, err := s.db.Queries().GetOrCreateTrack(ctx, sqlc.GetOrCreateTrackParams{
			ID:        uuid.NewString(),
			SpotifyID: track.ID.String(),
			Title:     track.Name,
		})
		if err != nil {
			return fmt.Errorf("failed to get/create track %s: %w", track.ID, err)
		}

		_, err = s.db.Queries().GetOrCreateAlbumTrack(ctx, sqlc.GetOrCreateAlbumTrackParams{
			AlbumID: albumModel.ID,
			TrackID: trackModel.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to get/create album track: %w", err)
		}

		for _, a := range album.Artists {
			artistModel, err := s.db.Queries().GetOrCreateArtist(ctx, sqlc.GetOrCreateArtistParams{
				ID:        uuid.NewString(),
				SpotifyID: a.ID.String(),
				Name:      a.Name,
			})
			if err != nil {
				return fmt.Errorf("failed to get/create artist %s: %w", a.ID, err)
			}
			_, err = s.db.Queries().GetOrCreateAlbumArtist(ctx, sqlc.GetOrCreateAlbumArtistParams{
				AlbumID:  albumModel.ID,
				ArtistID: artistModel.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to get/create album artist: %w", err)
			}
		}

		err = s.db.Queries().UpsertTrackPlay(ctx, sqlc.UpsertTrackPlayParams{
			ID:       uuid.NewString(),
			UserID:   userID,
			TrackID:  trackModel.ID,
			AlbumID:  albumModel.ID,
			PlayedAt: item.PlayedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert track play: %w", err)
		}
	}
	return nil
}

func (s *Service) GetLastPlayedAtByAlbumIds(ctx context.Context, userID string, albumIDs []string) (map[string]time.Time, error) {
	if len(albumIDs) == 0 {
		return map[string]time.Time{}, nil
	}

	rows, err := s.db.Queries().GetLastPlayedAtByAlbumIds(ctx, sqlc.GetLastPlayedAtByAlbumIdsParams{
		UserID:   userID,
		AlbumIds: albumIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get last played at: %w", err)
	}

	result := make(map[string]time.Time, len(rows))
	for _, row := range rows {
		t, err := parseInterfaceTime(row.LastPlayedAt)
		if err != nil {
			continue
		}
		result[row.AlbumID] = t
	}
	return result, nil
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

// parseInterfaceTime converts a SQLite interface{} datetime value to time.Time.
// SQLite returns datetime aggregates (like MAX) as strings.
func parseInterfaceTime(v interface{}) (time.Time, error) {
	if v == nil {
		return time.Time{}, fmt.Errorf("nil time value")
	}
	switch val := v.(type) {
	case string:
		formats := []string{
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05",
			time.RFC3339Nano,
			time.RFC3339,
		}
		for _, format := range formats {
			t, err := time.Parse(format, val)
			if err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("could not parse time string: %s", val)
	case []byte:
		return parseInterfaceTime(string(val))
	case time.Time:
		return val, nil
	default:
		return time.Time{}, fmt.Errorf("unexpected time type: %T", v)
	}
}
