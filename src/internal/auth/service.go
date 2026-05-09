package auth

import (
	"fmt"
	"net/http"

	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/feed"
	"github.com/alecdray/wax/src/internal/spotify"
	"github.com/alecdray/wax/src/internal/user"
)

// Service owns the auth domain: login orchestration, the Spotify OAuth
// callback flow, and JWT-claims helpers used by the HTTP adapter. It
// composes peer module services (`spotify.AuthService`, `user.Service`,
// `feed.Service`) to bootstrap a user on first login.
type Service struct {
	spotifyAuth *spotify.AuthService
	userService *user.Service
	feedService *feed.Service
}

func NewService(spotifyAuth *spotify.AuthService, userService *user.Service, feedService *feed.Service) *Service {
	return &Service{
		spotifyAuth: spotifyAuth,
		userService: userService,
		feedService: feedService,
	}
}

// SpotifyAuthURL returns the Spotify OAuth authorization URL the login page
// links to. The state code comes from app config and is verified on
// callback.
func (s *Service) SpotifyAuthURL(stateCode string) string {
	return s.spotifyAuth.AuthURL(stateCode)
}

// LoginRedirect returns a non-empty redirect URL when the request already
// carries a valid session for a user with a stored Spotify refresh token —
// i.e., already logged in. It returns "" when the caller should render the
// login page instead.
func (s *Service) LoginRedirect(ctx contextx.ContextX, claims *app.Claims) (string, error) {
	if claims == nil || claims.UserID == nil {
		return "", nil
	}

	a, err := ctx.App()
	if err != nil {
		return "", fmt.Errorf("failed to get app: %w", err)
	}

	u, err := s.userService.GetUserById(ctx, *claims.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if u.SpotifyRefreshToken(a.Config().SpotifyTokenSecret) == nil {
		return "", nil
	}

	return "/app/library/dashboard", nil
}

// CompleteSpotifyLogin runs the Spotify OAuth callback flow: exchanges the
// authorization code for a token, fetches the Spotify user, upserts the
// user record (encrypting the refresh token), and ensures a feed exists.
// Returns the local user ID so the caller can update JWT claims.
func (s *Service) CompleteSpotifyLogin(ctx contextx.ContextX, r *http.Request) (string, error) {
	a, err := ctx.App()
	if err != nil {
		return "", fmt.Errorf("failed to get app: %w", err)
	}

	client, err := s.spotifyAuth.GetClientFromCallback(ctx, a.Config().StateCode, r)
	if err != nil {
		return "", fmt.Errorf("failed to get spotify client: %w", err)
	}

	token, err := client.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get spotify token: %w", err)
	}

	spotifyUser, err := client.CurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get spotify user: %w", err)
	}

	u, err := s.userService.UpsertSpotifyUser(ctx, spotifyUser.ID, token.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to upsert spotify user: %w", err)
	}

	if _, err := s.feedService.UpsertFeed(ctx, u.ID, models.FeedKindSpotify); err != nil {
		return "", fmt.Errorf("failed to upsert feed: %w", err)
	}

	return u.ID, nil
}
