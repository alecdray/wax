package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"time"

	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	ErrFailedToGetToken = errors.New("failed to get token")
	ErrStateMismatch    = errors.New("state mismatch")
)

type AuthService struct {
	*spotifyauth.Authenticator
	// guard is the shared rate-limit guard. It is threaded into every SDK
	// client built here (via the oauth2 base client) and into the raw Client,
	// so all Spotify traffic draws on one per-app rate-limit view. See ADR 0006.
	guard *guard
}

func NewAuthService(clientID, clientSecret, redirectURI string, scopes ...string) *AuthService {
	return &AuthService{
		Authenticator: spotifyauth.New(
			spotifyauth.WithClientID(clientID),
			spotifyauth.WithClientSecret(clientSecret),
			spotifyauth.WithRedirectURL(redirectURI),
			spotifyauth.WithScopes(scopes...),
		),
		guard: newGuard(nil),
	}
}

// guardedCtx returns ctx carrying an http.Client whose transport is the shared
// guard. oauth2 uses this as the base for both the SDK's API calls and token
// refreshes, so every request the resulting *spotify.Client makes is paced and
// Retry-After-aware.
func (auth *AuthService) guardedCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: auth.guard})
}

// AuthURLForcingConsent returns the OAuth authorization URL with show_dialog=true,
// forcing Spotify to display the consent screen. Without this, Spotify silently
// re-uses a user's existing authorization and never grants scopes added since
// then (e.g. the playlist scopes the radar inbox needs), so the refreshed token
// keeps the old, narrower scope set.
func (auth *AuthService) AuthURLForcingConsent(state string) string {
	return auth.AuthURL(state, spotifyauth.ShowDialog)
}

func (auth *AuthService) GetClientFromCallback(ctx contextx.ContextX, state string, r *http.Request) (*spotify.Client, error) {
	token, err := auth.Token(ctx, state, r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetToken, err)
	}
	if st := r.FormValue("state"); st != state {
		return nil, ErrStateMismatch
	}

	gctx := auth.guardedCtx(ctx)
	return spotify.New(auth.Client(gctx, token)), nil
}

// RefreshAccessToken exchanges a refresh token for a fresh access token and
// returns it so the caller can cache it. The exchange runs through the guard.
func (auth *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	partialToken := oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Second),
	}

	token, err := auth.RefreshToken(auth.guardedCtx(ctx), &partialToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetToken, err)
	}

	return token, nil
}

// ClientFromToken builds a guarded SDK client from an existing token. The token
// source refreshes only if the token is expired, so passing a still-valid
// cached token issues no token-endpoint call.
func (auth *AuthService) ClientFromToken(ctx context.Context, token *oauth2.Token) *spotify.Client {
	return spotify.New(auth.Client(auth.guardedCtx(ctx), token))
}
