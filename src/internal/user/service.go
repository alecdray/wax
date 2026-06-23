package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/cryptox"
	"github.com/alecdray/wax/src/internal/core/db"
)

type Service struct {
	db   *db.DB
	repo *Repo
}

func NewService(d *db.DB) *Service {
	return &Service{
		db:   d,
		repo: NewRepo(d.Queries()),
	}
}

func (s *Service) GetUserById(ctx context.Context, id string) (*UserDTO, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) GetUserBySpotifyID(ctx context.Context, spotifyId string) (*UserDTO, error) {
	return s.repo.GetUserBySpotifyID(ctx, spotifyId)
}

func (s *Service) UpsertSpotifyUser(ctx contextx.ContextX, spotifyId string, spotifyRefreshToken string) (*UserDTO, error) {
	app, err := ctx.App()
	if err != nil {
		err = fmt.Errorf("failed to get app: %w", err)
		return nil, err
	}

	encryptedSpotifyRefreshToken, err := cryptox.SymmetricEncrypt(spotifyRefreshToken, app.Config().SpotifyTokenSecret)
	if err != nil {
		err = fmt.Errorf("failed to encrypt spotify refresh token: %w", err)
		return nil, err
	}

	return s.repo.UpsertSpotifyUser(ctx, spotifyId, encryptedSpotifyRefreshToken)
}

// SetSpotifyAccessToken encrypts and persists a user's Spotify access token and
// its expiry, so subsequent calls can reuse it until it expires instead of
// exchanging the refresh token every time.
func (s *Service) SetSpotifyAccessToken(ctx contextx.ContextX, userID, accessToken string, expiresAt time.Time) error {
	app, err := ctx.App()
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	encrypted, err := cryptox.SymmetricEncrypt(accessToken, app.Config().SpotifyTokenSecret)
	if err != nil {
		return fmt.Errorf("failed to encrypt spotify access token: %w", err)
	}

	return s.repo.SetSpotifyAccessToken(ctx, userID, encrypted, expiresAt)
}

func (s *Service) GetUserFromCtx(ctx contextx.ContextX) (*UserDTO, error) {
	userId, err := ctx.UserId()
	if errors.Is(err, contextx.ErrEmptyValue) {
		userId = ""
	} else if err != nil {
		err = fmt.Errorf("failed to get user id: %w", err)
		return nil, err
	}

	if userId == "" {
		app, err := ctx.App()
		if err != nil {
			err = fmt.Errorf("failed to get app: %w", err)
			return nil, err
		}

		userId = *app.Claims().UserID
	}

	if userId != "" {
		userDto, err := s.GetUserById(ctx, userId)
		if err != nil {
			err = fmt.Errorf("failed to get user by id: %w", err)
			return nil, err
		}
		return userDto, nil
	}

	return nil, errors.New("unauthorized")
}
