package user

import (
	"github.com/alecdray/wax/src/internal/core/cryptox"
)

type UserDTO struct {
	ID                  string
	SpotifyID           string
	spotifyRefreshToken *string
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
