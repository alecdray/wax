package spotify

import (
	"fmt"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/user"

	spotify "github.com/zmb3/spotify/v2"
)

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

// SearchAlbums runs a Spotify catalog search restricted to albums.
// limit is clamped to the Spotify API max of 50.
func (s *Service) SearchAlbums(ctx contextx.ContextX, userId, query string, limit int) ([]spotify.SimpleAlbum, error) {
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
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
