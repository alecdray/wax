package review

import (
	"context"
	"database/sql"

	"github.com/alecdray/wax/src/internal/core/db/sqlc"

	"github.com/google/uuid"
)

// Repo is the review module's data access layer. It is the only file in
// package review that imports core/db/sqlc. Repo methods return review DTOs —
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

// --- DTO conversion helpers ---

// albumRatingDTOFromModel converts a sqlc AlbumRatingLog row into an
// AlbumRatingDTO.
func albumRatingDTOFromModel(model sqlc.AlbumRatingLog) *AlbumRatingDTO {
	dto := &AlbumRatingDTO{
		ID:        model.ID,
		UserID:    model.UserID,
		AlbumID:   model.AlbumID,
		CreatedAt: model.CreatedAt,
	}

	dto.Rating = &model.Rating

	if model.Note.Valid {
		dto.Note = &model.Note.String
	}

	if model.State.Valid {
		s := RatingState(model.State.String)
		dto.State = &s
	}

	return dto
}

func ratingStateDTOFromModel(model sqlc.AlbumRatingState) *RatingStateDTO {
	return &RatingStateDTO{
		ID:          model.ID,
		AlbumID:     model.AlbumID,
		UserID:      model.UserID,
		State:       RatingState(model.State),
		LastRatedAt: model.CreatedAt,
	}
}

// --- Rating log mutations / lookups ---

func (r *Repo) InsertAlbumRatingLogEntry(ctx context.Context, userID, albumID string, rating float64, note string, state RatingState) (*AlbumRatingDTO, error) {
	var noteParam sql.NullString
	if note != "" {
		noteParam = sql.NullString{String: note, Valid: true}
	}

	model, err := r.q.InsertAlbumRatingLogEntry(ctx, sqlc.InsertAlbumRatingLogEntryParams{
		ID:      uuid.NewString(),
		UserID:  userID,
		AlbumID: albumID,
		Rating:  rating,
		Note:    noteParam,
		State:   sql.NullString{String: string(state), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return albumRatingDTOFromModel(model), nil
}

func (r *Repo) DeleteAlbumRatingLogEntry(ctx context.Context, userID, entryID string) error {
	return r.q.DeleteAlbumRatingLogEntry(ctx, sqlc.DeleteAlbumRatingLogEntryParams{
		ID:     entryID,
		UserID: userID,
	})
}

// GetLatestUserAlbumRating returns the latest rating for one user/album, or
// the underlying error (including sql.ErrNoRows) if no rating exists.
func (r *Repo) GetLatestUserAlbumRating(ctx context.Context, userID, albumID string) (*AlbumRatingDTO, error) {
	row, err := r.q.GetLatestUserAlbumRating(ctx, sqlc.GetLatestUserAlbumRatingParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, err
	}
	return albumRatingDTOFromModel(row), nil
}

// GetLatestUserAlbumRatings returns latest ratings keyed by album ID.
func (r *Repo) GetLatestUserAlbumRatings(ctx context.Context, userID string) (map[string]AlbumRatingDTO, error) {
	rows, err := r.q.GetLatestUserAlbumRatings(ctx, sqlc.GetLatestUserAlbumRatingsParams{
		UserID:   userID,
		UserID_2: userID,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]AlbumRatingDTO, len(rows))
	for _, row := range rows {
		out[row.AlbumID] = *albumRatingDTOFromModel(row)
	}
	return out, nil
}

// GetUserAlbumRatingLog returns the historical rating log for one user/album.
func (r *Repo) GetUserAlbumRatingLog(ctx context.Context, userID, albumID string) ([]*AlbumRatingDTO, error) {
	rows, err := r.q.GetUserAlbumRatingLog(ctx, sqlc.GetUserAlbumRatingLogParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, err
	}

	dtos := make([]*AlbumRatingDTO, len(rows))
	for i, row := range rows {
		dtos[i] = albumRatingDTOFromModel(row)
	}
	return dtos, nil
}

// --- Rating-state lookups / mutations ---

// GetAlbumRatingState returns the current rating state for one user/album, or
// nil when the row does not exist.
func (r *Repo) GetAlbumRatingState(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	model, err := r.q.GetAlbumRatingState(ctx, sqlc.GetAlbumRatingStateParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ratingStateDTOFromModel(model), nil
}

// GetAllAlbumRatingStates returns every rating state for the user keyed by
// album ID.
func (r *Repo) GetAllAlbumRatingStates(ctx context.Context, userID string) (map[string]*RatingStateDTO, error) {
	rows, err := r.q.GetAllAlbumRatingStates(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*RatingStateDTO, len(rows))
	for _, row := range rows {
		dto := ratingStateDTOFromModel(row)
		result[dto.AlbumID] = dto
	}
	return result, nil
}

// InsertAlbumRatingState creates a fresh rating state row with the given
// lifecycle value.
func (r *Repo) InsertAlbumRatingState(ctx context.Context, userID, albumID string, state RatingState) (*RatingStateDTO, error) {
	model, err := r.q.InsertAlbumRatingState(ctx, sqlc.InsertAlbumRatingStateParams{
		ID:      uuid.NewString(),
		UserID:  userID,
		AlbumID: albumID,
		State:   string(state),
	})
	if err != nil {
		return nil, err
	}
	return ratingStateDTOFromModel(model), nil
}

// UpdateAlbumRatingState writes the new lifecycle value for the user/album.
func (r *Repo) UpdateAlbumRatingState(ctx context.Context, userID, albumID string, state RatingState) (*RatingStateDTO, error) {
	model, err := r.q.UpdateAlbumRatingState(ctx, sqlc.UpdateAlbumRatingStateParams{
		State:   string(state),
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, err
	}
	return ratingStateDTOFromModel(model), nil
}
