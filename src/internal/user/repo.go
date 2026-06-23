package user

import (
	"context"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/alecdray/wax/src/internal/core/sqlx"

	"github.com/google/uuid"
)

// Repo is the user module's data access layer. It is the only file in
// package user that imports core/db/sqlc. Repo methods return user DTOs —
// never sqlc.* types.
type Repo struct {
	q *sqlc.Queries
}

// NewRepo binds a Repo to the given Queries. Callers can bind to db.Queries()
// for the global handle or to tx.Queries() inside a db.WithTx callback for
// transactional work.
func NewRepo(q *sqlc.Queries) *Repo {
	return &Repo{q: q}
}

// --- DTO conversion helpers (private — only repo.go touches sqlc types) ---

func userDTOFromModel(model sqlc.User) *UserDTO {
	dto := &UserDTO{
		ID:        model.ID,
		SpotifyID: model.SpotifyID,
	}

	if model.SpotifyRefreshToken.Valid {
		dto.spotifyRefreshToken = &model.SpotifyRefreshToken.String
	}

	if model.SpotifyAccessToken.Valid {
		dto.spotifyAccessToken = &model.SpotifyAccessToken.String
	}

	if model.SpotifyAccessTokenExpiresAt.Valid {
		dto.spotifyAccessTokenExpiresAt = &model.SpotifyAccessTokenExpiresAt.Time
	}

	return dto
}

// --- User lookups / mutations ---

func (r *Repo) GetUserByID(ctx context.Context, id string) (*UserDTO, error) {
	user, err := r.q.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return userDTOFromModel(user), nil
}

func (r *Repo) GetUserBySpotifyID(ctx context.Context, spotifyID string) (*UserDTO, error) {
	user, err := r.q.GetUserBySpotifyId(ctx, spotifyID)
	if err != nil {
		return nil, err
	}
	return userDTOFromModel(user), nil
}

func (r *Repo) SetSpotifyAccessToken(ctx context.Context, id, encryptedAccessToken string, expiresAt time.Time) error {
	return r.q.SetSpotifyAccessToken(ctx, sqlc.SetSpotifyAccessTokenParams{
		SpotifyAccessToken:          sqlx.NewNullString(encryptedAccessToken),
		SpotifyAccessTokenExpiresAt: sqlx.NewNullTime(&expiresAt),
		ID:                          id,
	})
}

func (r *Repo) UpsertSpotifyUser(ctx context.Context, spotifyID string, encryptedSpotifyRefreshToken string) (*UserDTO, error) {
	user, err := r.q.UpsertSpotifyUser(ctx, sqlc.UpsertSpotifyUserParams{
		ID:                  uuid.New().String(),
		SpotifyID:           spotifyID,
		SpotifyRefreshToken: sqlx.NewNullString(encryptedSpotifyRefreshToken),
	})
	if err != nil {
		return nil, err
	}
	return userDTOFromModel(user), nil
}
