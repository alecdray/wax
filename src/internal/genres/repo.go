package genres

import (
	"context"

	"github.com/alecdray/wax/src/internal/core/db/sqlc"
)

// Repo is the genres module's data access layer — the only file in package
// genres that imports core/db/sqlc. Repo methods return genre DTOs, never
// sqlc.* types.
type Repo struct {
	q *sqlc.Queries
}

// NewRepo binds a Repo to the given Queries. Callers bind to db.Queries() for
// the global handle or to tx.Queries() inside a db.WithTx callback.
func NewRepo(q *sqlc.Queries) *Repo {
	return &Repo{q: q}
}

// GetAlbumGenresByAlbumIDs returns resolved leaf genres grouped by album ID.
func (r *Repo) GetAlbumGenresByAlbumIDs(ctx context.Context, albumIDs []string) (map[string][]AlbumGenreDTO, error) {
	result := make(map[string][]AlbumGenreDTO)
	if len(albumIDs) == 0 {
		return result, nil
	}
	rows, err := r.q.GetAlbumGenresByAlbumIds(ctx, albumIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.AlbumID] = append(result[row.AlbumID], AlbumGenreDTO{
			GenreID: row.GenreID,
			Label:   row.GenreLabel,
		})
	}
	return result, nil
}

// EnrichedAlbumIDs returns the subset of the given album IDs already enriched.
func (r *Repo) EnrichedAlbumIDs(ctx context.Context, albumIDs []string) (map[string]bool, error) {
	enriched := make(map[string]bool)
	if len(albumIDs) == 0 {
		return enriched, nil
	}
	ids, err := r.q.GetEnrichedAlbumIds(ctx, albumIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		enriched[id] = true
	}
	return enriched, nil
}

// ReplaceAlbumGenres clears an album's genres and inserts the given ones.
func (r *Repo) ReplaceAlbumGenres(ctx context.Context, albumID string, genres []AlbumGenreDTO) error {
	if err := r.q.DeleteAlbumGenresByAlbumId(ctx, albumID); err != nil {
		return err
	}
	for _, g := range genres {
		if err := r.q.UpsertAlbumGenre(ctx, sqlc.UpsertAlbumGenreParams{
			ID:         newID(),
			AlbumID:    albumID,
			GenreID:    g.GenreID,
			GenreLabel: g.Label,
		}); err != nil {
			return err
		}
	}
	return nil
}

// MarkEnriched records that an album's genres were resolved.
func (r *Repo) MarkEnriched(ctx context.Context, albumID string) error {
	return r.q.MarkAlbumGenreEnriched(ctx, albumID)
}
