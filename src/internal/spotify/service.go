package spotify

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/user"

	spotify "github.com/zmb3/spotify/v2"
)

// Radar inbox playlist identity. The playlist is private and Wax-managed.
const (
	radarPlaylistName        = "wax radar"
	radarPlaylistDescription = "Add albums here to send them to your wax radar."
)

// ErrPlaylistNotFound is returned by GetPlaylistItems when the playlist no
// longer exists on Spotify (HTTP 404) — e.g. the user deleted it.
var ErrPlaylistNotFound = errors.New("spotify playlist not found")

// ErrInsufficientScope is returned when Spotify rejects a request for lack of a
// granted scope (HTTP 403) — e.g. a user connected before playlist scopes were
// requested tries to create the radar playlist. The caller re-authenticates.
var ErrInsufficientScope = errors.New("spotify request denied for insufficient scope")

// spotifyStatus reports the HTTP status of a Spotify API error, if it is one.
func spotifyStatus(err error) (int, bool) {
	var spotifyErr spotify.Error
	if errors.As(err, &spotifyErr) {
		return spotifyErr.Status, true
	}
	return 0, false
}

// PlaylistItem is a minimal view of one track in a playlist: the track's own id
// (used to remove it) and the spotify id of the album it belongs to. AlbumSpotifyID
// is empty for local files or tracks unavailable in the market.
type PlaylistItem struct {
	TrackID        string
	AlbumSpotifyID string
}

type Service struct {
	client             *Client
	spotifyAuthService *AuthService
	userService        *user.Service
}

func NewService(userService *user.Service, spotifyAuthService *AuthService) *Service {
	return &Service{
		client:             NewClient(),
		userService:        userService,
		spotifyAuthService: spotifyAuthService,
	}
}

func (s *Service) Client(ctx contextx.ContextX, userId string) (*spotify.Client, error) {
	user, err := s.userService.GetUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	app, err := ctx.App()
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	userRefreshToken := user.SpotifyRefreshToken(app.Config().SpotifyTokenSecret)
	if userRefreshToken == nil {
		return nil, fmt.Errorf("user has no spotify refresh token")
	}

	client, err := s.spotifyAuthService.GetClientFromRefreshToken(ctx, *userRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetToken, err)
	}

	return client, nil
}

func (s *Service) GetUser(ctx contextx.ContextX, userId string) (*spotify.PrivateUser, error) {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return nil, err
	}

	return client.CurrentUser(ctx)
}

func (s *Service) GetRecentlySavedAlbums(ctx contextx.ContextX, userId string, window time.Duration) ([]spotify.SavedAlbum, error) {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return nil, err
	}

	var userAlbums []spotify.SavedAlbum = make([]spotify.SavedAlbum, 0, 50)
	minTime := time.Now().Add(-window)
	maxTime := time.Now()
	offset := 0
	for maxTime.After(minTime) {
		albums, err := client.CurrentUsersAlbums(ctx, spotify.Limit(50), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		offset += len(albums.Albums)

		for _, album := range albums.Albums {
			addedAt, err := time.Parse(time.RFC3339, album.AddedAt)
			if err != nil {
				return nil, err
			}

			if addedAt.After(minTime) {
				userAlbums = append(userAlbums, album)
			}

			if addedAt.Before(maxTime) {
				maxTime = addedAt
			}
		}

		if len(albums.Albums) < 50 {
			break
		}
	}

	return userAlbums, nil
}

func (s *Service) GetRecentlyPlayedTracks(ctx contextx.ContextX, userId string) ([]spotify.RecentlyPlayedItem, error) {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return nil, err
	}

	return client.PlayerRecentlyPlayedOpt(ctx, &spotify.RecentlyPlayedOptions{
		Limit: 50,
	})
}

func (s *Service) GetUsersSavedAlbums(ctx contextx.ContextX, userId string) ([]spotify.SavedAlbum, error) {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return nil, err
	}

	var collectedAlbums []spotify.SavedAlbum = make([]spotify.SavedAlbum, 0)
	limit := 50
	offset := 0
	for offset < 1_000 {
		albums, err := client.CurrentUsersAlbums(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		if len(albums.Albums) == 0 {
			break
		}

		collectedAlbums = append(collectedAlbums, albums.Albums...)

		offset += len(albums.Albums)
	}
	return collectedAlbums, nil
}

func (s *Service) RemoveAlbumFromSavedLibrary(ctx contextx.ContextX, userId, spotifyId string) error {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return fmt.Errorf("failed to get spotify client: %w", err)
	}

	token, err := client.Token()
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	return s.client.RemoveAlbum(ctx, token.AccessToken, spotifyId)
}

// AddAlbumToSavedLibrary saves an album to the user's Spotify saved library.
// Mirrors RemoveAlbumFromSavedLibrary; uses the SDK directly (no raw HTTP).
func (s *Service) AddAlbumToSavedLibrary(ctx contextx.ContextX, userId, spotifyId string) error {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return fmt.Errorf("failed to get spotify client: %w", err)
	}
	if err := client.AddAlbumsToLibrary(ctx, spotify.ID(spotifyId)); err != nil {
		return fmt.Errorf("failed to add album to spotify saved library: %w", err)
	}
	return nil
}

func (s *Service) GetUsersSavedTracks(ctx contextx.ContextX, userId string) ([]spotify.SavedTrack, error) {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return nil, err
	}

	var collectedTracks []spotify.SavedTrack = make([]spotify.SavedTrack, 0)
	limit := 50
	offset := 0
	for offset < 1_000 {
		tracks, err := client.CurrentUsersTracks(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		if len(tracks.Tracks) == 0 {
			break
		}

		collectedTracks = append(collectedTracks, tracks.Tracks...)

		offset += len(tracks.Tracks)
	}
	return collectedTracks, nil
}

// CreateRadarPlaylist creates the user's private "wax radar" inbox playlist and
// returns its id. Requires playlist-modify scope.
func (s *Service) CreateRadarPlaylist(ctx contextx.ContextX, userId string) (string, error) {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return "", err
	}
	me, err := client.CurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current spotify user: %w", err)
	}
	playlist, err := client.CreatePlaylistForUser(ctx, me.ID, radarPlaylistName, radarPlaylistDescription, false, false)
	if err != nil {
		if status, ok := spotifyStatus(err); ok && status == http.StatusForbidden {
			return "", ErrInsufficientScope
		}
		return "", fmt.Errorf("failed to create radar playlist: %w", err)
	}
	return playlist.ID.String(), nil
}

// GetPlaylistItems returns the playlist's tracks as minimal PlaylistItems,
// paginated. Local files and unavailable tracks are skipped. Returns
// ErrPlaylistNotFound if the playlist no longer exists.
func (s *Service) GetPlaylistItems(ctx contextx.ContextX, userId, playlistID string) ([]PlaylistItem, error) {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return nil, err
	}
	const limit = 100
	items := make([]PlaylistItem, 0)
	offset := 0
	for {
		page, err := client.GetPlaylistItems(ctx, spotify.ID(playlistID), spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			var spotifyErr spotify.Error
			if errors.As(err, &spotifyErr) && spotifyErr.Status == http.StatusNotFound {
				return nil, ErrPlaylistNotFound
			}
			return nil, fmt.Errorf("failed to get playlist items: %w", err)
		}
		for _, it := range page.Items {
			if it.Track.Track == nil {
				continue // episode, local file, or unavailable track
			}
			items = append(items, PlaylistItem{
				TrackID:        it.Track.Track.ID.String(),
				AlbumSpotifyID: it.Track.Track.Album.ID.String(),
			})
		}
		if len(page.Items) < limit {
			break
		}
		offset += len(page.Items)
	}
	return items, nil
}

// RemovePlaylistTracks removes the given tracks (by track id) from the playlist,
// batching at the Spotify per-request cap of 100. A no-op for an empty list.
func (s *Service) RemovePlaylistTracks(ctx contextx.ContextX, userId, playlistID string, trackIDs []string) error {
	if len(trackIDs) == 0 {
		return nil
	}
	client, err := s.Client(ctx, userId)
	if err != nil {
		return err
	}
	ids := make([]spotify.ID, len(trackIDs))
	for i, id := range trackIDs {
		ids[i] = spotify.ID(id)
	}
	for start := 0; start < len(ids); start += 100 {
		end := start + 100
		if end > len(ids) {
			end = len(ids)
		}
		if _, err := client.RemoveTracksFromPlaylist(ctx, spotify.ID(playlistID), ids[start:end]...); err != nil {
			return fmt.Errorf("failed to remove tracks from playlist: %w", err)
		}
	}
	return nil
}

// GetFullAlbum returns one Spotify album by ID, including artists and tracks.
func (s *Service) GetFullAlbum(ctx contextx.ContextX, userId, spotifyId string) (*spotify.FullAlbum, error) {
	client, err := s.Client(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get spotify client: %w", err)
	}
	album, err := client.GetAlbum(ctx, spotify.ID(spotifyId))
	if err != nil {
		return nil, fmt.Errorf("failed to get spotify album: %w", err)
	}
	return album, nil
}

// SearchAlbums runs a Spotify catalog search restricted to albums.
// limit is clamped to the Spotify API max of 10 for the search endpoint.
func (s *Service) SearchAlbums(ctx contextx.ContextX, userId, query string, limit int) ([]spotify.SimpleAlbum, error) {
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 10 {
		limit = 10
	}
	client, err := s.Client(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get spotify client: %w", err)
	}
	result, err := client.Search(ctx, query, spotify.SearchTypeAlbum, spotify.Limit(limit))
	if err != nil {
		return nil, fmt.Errorf("spotify album search failed: %w", err)
	}
	if result == nil || result.Albums == nil {
		return nil, nil
	}
	return result.Albums.Albums, nil
}
