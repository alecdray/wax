// Package spotify wraps the Spotify Web API. It exposes a Service for
// general user-scoped operations and an AuthService for the OAuth flow.
package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

// The playlist endpoints below target the post-February-2026 migration paths
// (POST /me/playlists, /playlists/{id}/items) which the vendor SDK does not yet
// use — it still calls the now-removed /users/{id}/playlists and
// /playlists/{id}/tracks paths, which 403 for Development-mode apps.

// PlaylistFollowed reports whether the playlist is in the user's library (i.e.
// followed) via a single GET /me/library/contains?uris=... call. Spotify has no
// true playlist delete — "deleting" only unfollows — so this is how the radar
// sync detects removal (it flips to false), without scanning every playlist.
func (c *Client) PlaylistFollowed(ctx contextx.ContextX, accessToken, playlistID string) (bool, error) {
	uri := url.QueryEscape("spotify:playlist:" + playlistID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiOrigin+"/v1/me/library/contains?uris="+uri, nil)
	if err != nil {
		return false, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status checking library membership: %s", resp.Status)
	}
	var results []bool
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return false, fmt.Errorf("failed to decode membership response: %w", err)
	}
	if len(results) == 0 {
		return false, nil
	}
	return results[0], nil
}

// CreatePlaylist creates a private playlist for the current user and returns its
// id, via POST /me/playlists.
func (c *Client) CreatePlaylist(ctx contextx.ContextX, accessToken, name, description string) (string, error) {
	body, err := json.Marshal(map[string]any{"name": name, "public": false, "description": description})
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiOrigin+"/v1/me/playlists", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status creating playlist: %s", resp.Status)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", fmt.Errorf("failed to decode playlist: %w", err)
	}
	return created.ID, nil
}

// GetPlaylistItems reads a playlist's items (paginated) via GET
// /playlists/{id}/items. Local/unavailable items (no track) are skipped.
// Returns ErrPlaylistNotFound on 404.
func (c *Client) GetPlaylistItems(ctx contextx.ContextX, accessToken, playlistID string) ([]PlaylistItem, error) {
	const limit = 100
	items := make([]PlaylistItem, 0)
	offset := 0
	for {
		url := fmt.Sprintf("%s/v1/playlists/%s/items?limit=%d&offset=%d", apiOrigin, playlistID, limit, offset)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, ErrPlaylistNotFound
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status reading playlist items: %s", resp.Status)
		}
		// The Feb-2026 migration renamed each element's "track" field to "item".
		var page struct {
			Items []struct {
				Item *struct {
					ID    string `json:"id"`
					Album struct {
						ID string `json:"id"`
					} `json:"album"`
				} `json:"item"`
			} `json:"items"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode playlist items: %w", err)
		}
		resp.Body.Close()

		for _, it := range page.Items {
			if it.Item == nil {
				continue // episode, local file, or unavailable track
			}
			items = append(items, PlaylistItem{TrackID: it.Item.ID, AlbumSpotifyID: it.Item.Album.ID})
		}
		if len(page.Items) < limit {
			break
		}
		offset += len(page.Items)
	}
	return items, nil
}

// RemovePlaylistItems removes tracks (by id) from a playlist via DELETE
// /playlists/{id}/items, batching at the Spotify per-request cap of 100.
func (c *Client) RemovePlaylistItems(ctx contextx.ContextX, accessToken, playlistID string, trackIDs []string) error {
	for start := 0; start < len(trackIDs); start += 100 {
		end := start + 100
		if end > len(trackIDs) {
			end = len(trackIDs)
		}
		uris := make([]map[string]string, 0, end-start)
		for _, id := range trackIDs[start:end] {
			uris = append(uris, map[string]string{"uri": "spotify:track:" + id})
		}
		body, err := json.Marshal(map[string]any{"items": uris})
		if err != nil {
			return fmt.Errorf("failed to encode request: %w", err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
			fmt.Sprintf("%s/v1/playlists/%s/items", apiOrigin, playlistID), bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status removing playlist items: %s", resp.Status)
		}
	}
	return nil
}
