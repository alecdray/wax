package user

import (
	"time"

	"github.com/alecdray/wax/src/internal/core/cryptox"
)

type UserDTO struct {
	ID                          string
	SpotifyID                   string
	spotifyRefreshToken         *string
	spotifyAccessToken          *string
	spotifyAccessTokenExpiresAt *time.Time
}

func (u *UserDTO) SpotifyRefreshToken(secret string) *string {
	if u.spotifyRefreshToken == nil {
		return nil
	}

	decrypted, err := cryptox.SymmetricDecrypt(*u.spotifyRefreshToken, secret)
	if err != nil {
		return nil
	}

	return &decrypted
}

// CachedSpotifyAccessToken returns the decrypted cached access token and its
// expiry. ok is false when no cached token is stored or it fails to decrypt —
// the caller then refreshes. Freshness (is the expiry far enough in the future)
// is the caller's decision, since it owns the safety buffer.
func (u *UserDTO) CachedSpotifyAccessToken(secret string) (token string, expiresAt time.Time, ok bool) {
	if u.spotifyAccessToken == nil || u.spotifyAccessTokenExpiresAt == nil {
		return "", time.Time{}, false
	}

	decrypted, err := cryptox.SymmetricDecrypt(*u.spotifyAccessToken, secret)
	if err != nil {
		return "", time.Time{}, false
	}

	return decrypted, *u.spotifyAccessTokenExpiresAt, true
}
