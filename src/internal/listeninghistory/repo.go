package listeninghistory

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/sqlc"
)

// Repo is the listeninghistory module's data access layer. It is the only
// file in package listeninghistory that imports core/db/sqlc. Repo methods
// return domain types — never sqlc.* types.
type Repo struct {
	q *sqlc.Queries
}

// NewRepo binds a Repo to the given Queries. Callers can bind to db.Queries()
// for the global handle or to tx.Queries() inside a db.WithTx callback for
// transactional work.
func NewRepo(q *sqlc.Queries) *Repo {
	return &Repo{q: q}
}

// AlbumInput is the data needed to upsert an album.
type AlbumInput struct {
	ID        string
	SpotifyID string
	Title     string
	ImageURL  string
}

// TrackInput is the data needed to upsert a track.
type TrackInput struct {
	ID        string
	SpotifyID string
	Title     string
}

// ArtistInput is the data needed to upsert an artist.
type ArtistInput struct {
	ID        string
	SpotifyID string
	Name      string
}

// TrackPlayInput is the data needed to record a track play.
type TrackPlayInput struct {
	ID       string
	UserID   string
	TrackID  string
	AlbumID  string
	PlayedAt time.Time
}

// GetOrCreateAlbum upserts the album by spotify_id and returns its row ID.
func (r *Repo) GetOrCreateAlbum(ctx context.Context, in AlbumInput) (string, error) {
	model, err := r.q.GetOrCreateAlbum(ctx, sqlc.GetOrCreateAlbumParams{
		ID:        in.ID,
		SpotifyID: in.SpotifyID,
		Title:     in.Title,
		ImageUrl:  sql.NullString{String: in.ImageURL, Valid: in.ImageURL != ""},
	})
	if err != nil {
		return "", err
	}
	return model.ID, nil
}

// GetOrCreateTrack upserts the track by spotify_id and returns its row ID.
func (r *Repo) GetOrCreateTrack(ctx context.Context, in TrackInput) (string, error) {
	model, err := r.q.GetOrCreateTrack(ctx, sqlc.GetOrCreateTrackParams{
		ID:        in.ID,
		SpotifyID: in.SpotifyID,
		Title:     in.Title,
	})
	if err != nil {
		return "", err
	}
	return model.ID, nil
}

// GetOrCreateAlbumTrack ensures the (album, track) link exists.
func (r *Repo) GetOrCreateAlbumTrack(ctx context.Context, albumID, trackID string) error {
	_, err := r.q.GetOrCreateAlbumTrack(ctx, sqlc.GetOrCreateAlbumTrackParams{
		AlbumID: albumID,
		TrackID: trackID,
	})
	return err
}

// GetOrCreateArtist upserts the artist by spotify_id and returns its row ID.
func (r *Repo) GetOrCreateArtist(ctx context.Context, in ArtistInput) (string, error) {
	model, err := r.q.GetOrCreateArtist(ctx, sqlc.GetOrCreateArtistParams{
		ID:        in.ID,
		SpotifyID: in.SpotifyID,
		Name:      in.Name,
	})
	if err != nil {
		return "", err
	}
	return model.ID, nil
}

// GetOrCreateAlbumArtist ensures the (album, artist) link exists.
func (r *Repo) GetOrCreateAlbumArtist(ctx context.Context, albumID, artistID string) error {
	_, err := r.q.GetOrCreateAlbumArtist(ctx, sqlc.GetOrCreateAlbumArtistParams{
		AlbumID:  albumID,
		ArtistID: artistID,
	})
	return err
}

// UpsertTrackPlay records a track play if not already present.
func (r *Repo) UpsertTrackPlay(ctx context.Context, in TrackPlayInput) error {
	return r.q.UpsertTrackPlay(ctx, sqlc.UpsertTrackPlayParams{
		ID:       in.ID,
		UserID:   in.UserID,
		TrackID:  in.TrackID,
		AlbumID:  in.AlbumID,
		PlayedAt: in.PlayedAt,
	})
}

// GetLastPlayedAtByAlbumIDs returns the latest play time per album for the user.
// Albums with unparseable timestamps are silently skipped.
func (r *Repo) GetLastPlayedAtByAlbumIDs(ctx context.Context, userID string, albumIDs []string) (map[string]time.Time, error) {
	if len(albumIDs) == 0 {
		return map[string]time.Time{}, nil
	}

	rows, err := r.q.GetLastPlayedAtByAlbumIds(ctx, sqlc.GetLastPlayedAtByAlbumIdsParams{
		UserID:   userID,
		AlbumIds: albumIDs,
	})
	if err != nil {
		return nil, err
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

// GetUserIDsWithSpotifyToken returns the IDs of all non-deleted users who have a Spotify refresh token.
func (r *Repo) GetUserIDsWithSpotifyToken(ctx context.Context) ([]string, error) {
	users, err := r.q.GetUsersWithSpotifyToken(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids, nil
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
