package crates

import (
	"context"

	"github.com/alecdray/wax/src/internal/core/db/sqlc"
)

// Repo is the crates module's data access layer. It is the only file in
// package crates that imports core/db/sqlc. Repo methods return crate DTOs —
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

// ListCrates returns all crates owned by the user with their album counts.
func (r *Repo) ListCrates(ctx context.Context, userID string) ([]CrateDTO, error) {
	rows, err := r.q.ListCrates(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]CrateDTO, len(rows))
	for i, row := range rows {
		out[i] = CrateDTO{
			ID:         row.ID,
			Name:       row.Name,
			AlbumCount: int(row.AlbumCount),
			CreatedAt:  row.CreatedAt,
		}
	}
	return out, nil
}

// GetCrate returns the base CrateDTO for the given crate. Album hydration is
// handled at the service layer.
func (r *Repo) GetCrate(ctx context.Context, id, userID string) (CrateDTO, error) {
	row, err := r.q.GetCrate(ctx, sqlc.GetCrateParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return CrateDTO{}, err
	}
	return CrateDTO{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
	}, nil
}

// GetCrateAlbumIDs returns the album IDs in a crate ordered by added_at DESC.
func (r *Repo) GetCrateAlbumIDs(ctx context.Context, crateID string) ([]string, error) {
	return r.q.GetCrateAlbumIDs(ctx, crateID)
}

// InsertCrate creates a new crate and returns its DTO.
func (r *Repo) InsertCrate(ctx context.Context, id, userID, name string) (CrateDTO, error) {
	crate, err := r.q.InsertCrate(ctx, sqlc.InsertCrateParams{
		ID:     id,
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		return CrateDTO{}, err
	}
	return CrateDTO{
		ID:        crate.ID,
		Name:      crate.Name,
		CreatedAt: crate.CreatedAt,
	}, nil
}

// DeleteCrate removes a crate owned by the given user.
func (r *Repo) DeleteCrate(ctx context.Context, id, userID string) error {
	return r.q.DeleteCrate(ctx, sqlc.DeleteCrateParams{
		ID:     id,
		UserID: userID,
	})
}

// InsertCrateAlbum adds an album to a crate. Idempotent — ON CONFLICT DO NOTHING.
func (r *Repo) InsertCrateAlbum(ctx context.Context, id, crateID, albumID string) error {
	return r.q.InsertCrateAlbum(ctx, sqlc.InsertCrateAlbumParams{
		ID:      id,
		CrateID: crateID,
		AlbumID: albumID,
	})
}

// DeleteCrateAlbum removes an album from a crate.
func (r *Repo) DeleteCrateAlbum(ctx context.Context, crateID, albumID string) error {
	return r.q.DeleteCrateAlbum(ctx, sqlc.DeleteCrateAlbumParams{
		CrateID: crateID,
		AlbumID: albumID,
	})
}
