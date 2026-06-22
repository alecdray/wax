package feed

import (
	"context"
	"errors"
	"testing"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/spotify"
)

type fakeRadarSpotify struct {
	items     []spotify.PlaylistItem
	itemsErr  error
	removed   []string
	removeErr error
	// notInLibrary inverts the default so the zero value means "in library" and
	// existing tests keep ingesting; set true to simulate the user unfollowing.
	notInLibrary bool
}

func (f *fakeRadarSpotify) PlaylistInLibrary(_ contextx.ContextX, _, _ string) (bool, error) {
	return !f.notInLibrary, nil
}

func (f *fakeRadarSpotify) GetPlaylistItems(_ contextx.ContextX, _, _ string) ([]spotify.PlaylistItem, error) {
	return f.items, f.itemsErr
}

func (f *fakeRadarSpotify) RemovePlaylistTracks(_ contextx.ContextX, _, _ string, trackIDs []string) error {
	f.removed = append(f.removed, trackIDs...)
	return f.removeErr
}

type fakeRadarSink struct {
	results map[string]error // album spotify id -> error AddSpotifyAlbumToRadar returns
	calls   []string
}

func (f *fakeRadarSink) AddSpotifyAlbumToRadar(_ contextx.ContextX, _, spotifyID string) error {
	f.calls = append(f.calls, spotifyID)
	if f.results == nil {
		return nil
	}
	return f.results[spotifyID]
}

func bg() contextx.ContextX { return contextx.NewContextX(context.Background()) }

func TestIngestRadarPlaylist_AddsNewAlbumAndRemovesItsTrack(t *testing.T) {
	sp := &fakeRadarSpotify{items: []spotify.PlaylistItem{{TrackID: "t1", AlbumSpotifyID: "alb1"}}}
	sink := &fakeRadarSink{}

	removed, err := ingestRadarPlaylist(bg(), sp, sink, "u1", "pl1")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if len(sink.calls) != 1 || sink.calls[0] != "alb1" {
		t.Fatalf("expected one radar add for alb1, got %v", sink.calls)
	}
	if len(removed) != 1 || removed[0] != "t1" {
		t.Fatalf("expected t1 removed, got %v", removed)
	}
	if len(sp.removed) != 1 || sp.removed[0] != "t1" {
		t.Fatalf("expected t1 removed from playlist, got %v", sp.removed)
	}
}

func TestIngestRadarPlaylist_AlreadyDecidedAlbumIsRemovedNotFailed(t *testing.T) {
	sp := &fakeRadarSpotify{items: []spotify.PlaylistItem{{TrackID: "t1", AlbumSpotifyID: "owned"}}}
	sink := &fakeRadarSink{results: map[string]error{"owned": library.ErrAlbumAlreadyDecided}}

	removed, err := ingestRadarPlaylist(bg(), sp, sink, "u1", "pl1")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	// Already owned/wishlisted → handled: the track is removed even though no
	// radar entry is created.
	if len(removed) != 1 || removed[0] != "t1" {
		t.Fatalf("expected owned album's track removed, got %v", removed)
	}
}

func TestIngestRadarPlaylist_FailedAlbumLeavesItsTracks(t *testing.T) {
	sp := &fakeRadarSpotify{items: []spotify.PlaylistItem{
		{TrackID: "tok", AlbumSpotifyID: "ok"},
		{TrackID: "tbad", AlbumSpotifyID: "bad"},
	}}
	sink := &fakeRadarSink{results: map[string]error{"bad": errors.New("import blew up")}}

	removed, err := ingestRadarPlaylist(bg(), sp, sink, "u1", "pl1")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	// Only the successful album's track is removed; the failed one stays for retry.
	if len(removed) != 1 || removed[0] != "tok" {
		t.Fatalf("expected only tok removed, got %v", removed)
	}
}

func TestIngestRadarPlaylist_DeduplicatesTracksByAlbum(t *testing.T) {
	sp := &fakeRadarSpotify{items: []spotify.PlaylistItem{
		{TrackID: "t1", AlbumSpotifyID: "alb"},
		{TrackID: "t2", AlbumSpotifyID: "alb"},
		{TrackID: "t3", AlbumSpotifyID: "alb"},
	}}
	sink := &fakeRadarSink{}

	removed, err := ingestRadarPlaylist(bg(), sp, sink, "u1", "pl1")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if len(sink.calls) != 1 {
		t.Fatalf("expected one radar add for the deduped album, got %v", sink.calls)
	}
	if len(removed) != 3 {
		t.Fatalf("expected all three tracks of the album removed, got %v", removed)
	}
}

func TestIngestRadarPlaylist_IgnoresLocalTracksWithNoAlbum(t *testing.T) {
	sp := &fakeRadarSpotify{items: []spotify.PlaylistItem{
		{TrackID: "tlocal", AlbumSpotifyID: ""},
		{TrackID: "treal", AlbumSpotifyID: "alb"},
	}}
	sink := &fakeRadarSink{}

	removed, err := ingestRadarPlaylist(bg(), sp, sink, "u1", "pl1")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if len(sink.calls) != 1 || sink.calls[0] != "alb" {
		t.Fatalf("expected only the real album ingested, got %v", sink.calls)
	}
	for _, r := range removed {
		if r == "tlocal" {
			t.Fatalf("local track must be left in place, got removed %v", removed)
		}
	}
}

func TestIngestRadarPlaylist_UnfollowedPlaylistSignalsNotFound(t *testing.T) {
	// "Deleting" a playlist in Spotify only unfollows it (it stays readable), so
	// removal shows up as the playlist leaving the library, not as a 404. The
	// ingest must treat that as ErrPlaylistNotFound and not ingest anything.
	sp := &fakeRadarSpotify{
		notInLibrary: true,
		items:        []spotify.PlaylistItem{{TrackID: "t1", AlbumSpotifyID: "alb"}},
	}
	sink := &fakeRadarSink{}

	_, err := ingestRadarPlaylist(bg(), sp, sink, "u1", "pl1")
	if !errors.Is(err, spotify.ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound for an unfollowed playlist, got %v", err)
	}
	if len(sink.calls) != 0 {
		t.Fatalf("must not ingest from a playlist that left the library, got %v", sink.calls)
	}
}

func TestIngestRadarPlaylist_PropagatesPlaylistNotFound(t *testing.T) {
	sp := &fakeRadarSpotify{itemsErr: spotify.ErrPlaylistNotFound}
	sink := &fakeRadarSink{}

	_, err := ingestRadarPlaylist(bg(), sp, sink, "u1", "pl1")
	if !errors.Is(err, spotify.ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound to propagate, got %v", err)
	}
	if len(sink.calls) != 0 {
		t.Fatalf("expected no radar adds when the playlist is gone, got %v", sink.calls)
	}
}
