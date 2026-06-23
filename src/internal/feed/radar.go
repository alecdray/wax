package feed

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/spotify"
)

// radarSpotifyPort is the slice of spotify.Service the radar inbox sync needs.
// Defined here (consumer-side) so the sync can be tested with a fake.
type radarSpotifyPort interface {
	PlaylistFollowed(ctx contextx.ContextX, userID, playlistID string) (bool, error)
	GetPlaylistItems(ctx contextx.ContextX, userID, playlistID string) ([]spotify.PlaylistItem, error)
	RemovePlaylistTracks(ctx contextx.ContextX, userID, playlistID string, trackIDs []string) error
}

// radarSink persists an ingested album onto the user's radar.
type radarSink interface {
	AddSpotifyAlbumToRadar(ctx contextx.ContextX, userID, spotifyID string) error
}

// ingestRadarPlaylist reads the inbox playlist, adds each distinct album to the
// radar, and returns the track ids that were handled and removed. A track is
// handled when its album was added, or was already owned/wishlisted (returned as
// library.ErrAlbumAlreadyDecided) — both mean "deal with it, clear the track".
// Albums that fail to ingest for any other reason leave their tracks in place
// for the next cycle. Local/unknown tracks (no album id) are ignored and left.
func ingestRadarPlaylist(ctx contextx.ContextX, sp radarSpotifyPort, sink radarSink, userID, playlistID string) ([]string, error) {
	// Spotify "delete" only unfollows the playlist (it stays readable, no 404),
	// so detect removal with a one-call library-membership check and signal it as
	// ErrPlaylistNotFound to reuse the delete-the-feed recovery.
	followed, err := sp.PlaylistFollowed(ctx, userID, playlistID)
	if err != nil {
		return nil, err
	}
	if !followed {
		return nil, spotify.ErrPlaylistNotFound
	}

	items, err := sp.GetPlaylistItems(ctx, userID, playlistID)
	if err != nil {
		return nil, err
	}

	// Deduplicate tracks by album, preserving first-seen order.
	tracksByAlbum := map[string][]string{}
	albumOrder := make([]string, 0)
	for _, item := range items {
		if item.AlbumSpotifyID == "" {
			continue
		}
		if _, seen := tracksByAlbum[item.AlbumSpotifyID]; !seen {
			albumOrder = append(albumOrder, item.AlbumSpotifyID)
		}
		tracksByAlbum[item.AlbumSpotifyID] = append(tracksByAlbum[item.AlbumSpotifyID], item.TrackID)
	}

	toRemove := make([]string, 0)
	for _, albumID := range albumOrder {
		err := sink.AddSpotifyAlbumToRadar(ctx, userID, albumID)
		switch {
		case err == nil, errors.Is(err, library.ErrAlbumAlreadyDecided):
			toRemove = append(toRemove, tracksByAlbum[albumID]...)
		default:
			slog.Error("radar inbox: failed to ingest album; leaving tracks for retry",
				"user", userID, "album", albumID, "error", err)
		}
	}

	if err := sp.RemovePlaylistTracks(ctx, userID, playlistID, toRemove); err != nil {
		return toRemove, fmt.Errorf("failed to remove ingested tracks: %w", err)
	}
	return toRemove, nil
}

// SyncSpotifyRadarFeed ingests the user's radar inbox playlist. If the playlist
// has been deleted on Spotify (ErrPlaylistNotFound), the feed's handle is cleared
// and it stops syncing until the user re-enables — never silently recreated.
func (s *Service) SyncSpotifyRadarFeed(ctx contextx.ContextX, feed FeedDTO) (*FeedDTO, error) {
	if feed.Kind != models.FeedKindSpotifyRadar {
		return nil, fmt.Errorf("feed kind must be spotify_radar")
	}
	if feed.SourceRef == nil || *feed.SourceRef == "" {
		return nil, fmt.Errorf("radar feed %s has no playlist handle", feed.ID)
	}
	playlistID := *feed.SourceRef

	feed.SetSyncing()
	if _, err := s.UpdateFeed(ctx, feed); err != nil {
		return nil, fmt.Errorf("failed to update feed on sync start: %w", err)
	}

	_, ingestErr := ingestRadarPlaylist(ctx, s.spotifyService, s.libraryService, feed.UserID, playlistID)

	if errors.Is(ingestErr, spotify.ErrPlaylistNotFound) {
		// The user removed the playlist (opt-out). That is not a failure — delete
		// the feed so it disappears cleanly rather than lingering as a failed
		// entry; the radar inbox control returns to Enable, and re-enabling recreates.
		if err := s.repo.DeleteFeed(ctx, feed.ID); err != nil {
			return nil, fmt.Errorf("failed to delete radar feed after playlist removal: %w", err)
		}
		slog.Info("radar inbox: playlist removed by user, feed disabled", "feed", feed.ID)
		return nil, nil
	}

	if ingestErr != nil {
		feed.SetSyncFailed()
		if _, err := s.UpdateFeed(ctx, feed); err != nil {
			slog.Error("failed to update feed on sync error", "feed", feed.ID, "error", err)
		}
		return nil, ingestErr
	}

	feed.SetSyncSuccess()
	if _, err := s.UpdateFeed(ctx, feed); err != nil {
		return nil, fmt.Errorf("failed to update feed on sync success: %w", err)
	}
	return &feed, nil
}

// GetSyncableRadarFeeds returns the radar inbox feeds eligible to sync.
func (s *Service) GetSyncableRadarFeeds(ctx context.Context) ([]FeedDTO, error) {
	return s.repo.GetSyncableRadarFeeds(ctx)
}

// RadarInboxPlaylistID returns the user's radar inbox playlist id, or "" if the
// user has not enabled the inbox (no radar feed, or no playlist created yet).
func (s *Service) RadarInboxPlaylistID(ctx context.Context, userID string) (string, error) {
	feeds, err := s.repo.GetFeedsByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	for _, f := range feeds {
		if f.Kind == models.FeedKindSpotifyRadar && f.SourceRef != nil {
			return *f.SourceRef, nil
		}
	}
	return "", nil
}

// EnableRadarInbox opts the user into the Spotify radar inbox: it ensures a
// radar feed row exists and that a Wax-managed playlist has been created for it,
// returning the playlist id. Idempotent — if the playlist already exists, its id
// is returned without creating another. Creating the playlist requires playlist
// scope; the caller surfaces a missing-scope failure as a re-authentication.
func (s *Service) EnableRadarInbox(ctx contextx.ContextX, userID string) (string, error) {
	feed, err := s.repo.UpsertFeed(ctx, userID, models.FeedKindSpotifyRadar)
	if err != nil {
		return "", fmt.Errorf("failed to upsert radar feed: %w", err)
	}
	slog.Info("radar inbox enable: start", "user", userID, "feed", feed.ID, "existing_source_ref", feed.SourceRef)
	if feed.SourceRef != nil && *feed.SourceRef != "" {
		return *feed.SourceRef, nil
	}

	// Reuse an existing "wax radar" playlist if the user already has one
	// (idempotent re-enable / recovery); only create when none exists.
	playlistID, err := s.spotifyService.FindRadarPlaylist(ctx, userID)
	if err != nil {
		slog.Error("radar inbox enable: find playlist failed", "feed", feed.ID, "error", err)
		return "", fmt.Errorf("failed to find radar playlist: %w", err)
	}
	if playlistID == "" {
		playlistID, err = s.spotifyService.CreateRadarPlaylist(ctx, userID)
		if err != nil {
			slog.Error("radar inbox enable: create playlist failed", "feed", feed.ID, "error", err)
			return "", fmt.Errorf("failed to create radar playlist: %w", err)
		}
		slog.Info("radar inbox enable: created playlist", "feed", feed.ID, "playlist", playlistID)
	} else {
		slog.Info("radar inbox enable: reusing existing playlist", "feed", feed.ID, "playlist", playlistID)
	}
	if playlistID == "" {
		return "", fmt.Errorf("spotify returned an empty radar playlist id")
	}
	if err := s.repo.SetFeedSourceRef(ctx, feed.ID, playlistID); err != nil {
		return "", fmt.Errorf("failed to store radar playlist handle: %w", err)
	}
	slog.Info("radar inbox enable: stored handle", "feed", feed.ID, "playlist", playlistID)
	return playlistID, nil
}
