// Package spotify wraps the Spotify Web API. It exposes a Service for
// general user-scoped operations and an AuthService for the OAuth flow.
package spotify

import (
	"fmt"
	"net/http"

	"github.com/alecdray/wax/src/internal/core/contextx"
)

const apiOrigin = "https://api.spotify.com"

// Client owns low-level HTTP calls to the Spotify Web API for endpoints
// not covered by the vendor SDK. SDK-backed calls flow through the
// per-user *spotify.Client built by AuthService.
type Client struct{}

func NewClient() *Client {
	return &Client{}
}

// RemoveAlbum removes a single album from the authenticated user's saved
// library. The vendor SDK does not expose this endpoint, so it is issued
// as a direct HTTP request using the supplied bearer token.
func (c *Client) RemoveAlbum(ctx contextx.ContextX, accessToken, spotifyId string) error {
	uri := fmt.Sprintf("spotify:album:%s", spotifyId)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		apiOrigin+"/v1/me/library?uris="+uri, nil)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}
