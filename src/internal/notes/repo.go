package notes

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alecdray/wax/src/internal/core/db/sqlc"

	"github.com/google/uuid"
)

// Repo is the notes module's data access layer. It is the only file in
// package notes that imports core/db/sqlc. Repo methods return notes DTOs —
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

func albumNoteDTOFromModel(m sqlc.AlbumNote) *AlbumNoteDTO {
	return &AlbumNoteDTO{
		ID:        m.ID,
		UserID:    m.UserID,
		AlbumID:   m.AlbumID,
		Content:   m.Content,
		UpdatedAt: m.UpdatedAt,
	}
}

// --- Album note lookups / mutations ---

// UpsertAlbumNote inserts or updates the sleeve note for an album.
func (r *Repo) UpsertAlbumNote(ctx context.Context, userID, albumID, content string) (*AlbumNoteDTO, error) {
	model, err := r.q.UpsertAlbumNote(ctx, sqlc.UpsertAlbumNoteParams{
		ID:      uuid.NewString(),
		UserID:  userID,
		AlbumID: albumID,
		Content: content,
	})
	if err != nil {
		return nil, err
	}
	return albumNoteDTOFromModel(model), nil
}

// GetAlbumNote returns the sleeve note for an album, or nil if none exists.
func (r *Repo) GetAlbumNote(ctx context.Context, userID, albumID string) (*AlbumNoteDTO, error) {
	model, err := r.q.GetAlbumNote(ctx, sqlc.GetAlbumNoteParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return albumNoteDTOFromModel(model), nil
}

// GetAlbumNotesByAlbumIDs returns notes keyed by album ID for bulk fetching.
func (r *Repo) GetAlbumNotesByAlbumIDs(ctx context.Context, userID string, albumIDs []string) (map[string]*AlbumNoteDTO, error) {
	rows, err := r.q.GetAlbumNotesByAlbumIds(ctx, sqlc.GetAlbumNotesByAlbumIdsParams{
		UserID:   userID,
		AlbumIds: albumIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make(map[string]*AlbumNoteDTO, len(rows))
	for _, row := range rows {
		result[row.AlbumID] = albumNoteDTOFromModel(row)
	}
	return result, nil
}
