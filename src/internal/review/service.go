package review

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/google/uuid"
)

var ErrRatingStateNotFound = errors.New("rating state not found")

type AlbumRatingDTO struct {
	ID        string
	UserID    string
	AlbumID   string
	Rating    *float64
	Note      *string
	State     *RatingState
	CreatedAt time.Time
}

func NewAlbumRatingDTOFromModel(model sqlc.AlbumRatingLog) *AlbumRatingDTO {
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

func NewRatingStateDTOFromModel(model sqlc.AlbumRatingState) *RatingStateDTO {
	dto := &RatingStateDTO{
		ID:          model.ID,
		AlbumID:     model.AlbumID,
		UserID:      model.UserID,
		State:       RatingState(model.State),
		SnoozeCount: int(model.SnoozeCount),
		LastRatedAt: model.CreatedAt,
	}

	if model.NextRerateAt.Valid {
		t := model.NextRerateAt.Time
		dto.NextRerateAt = &t
	}

	return dto
}

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) AddRating(ctx context.Context, userID, albumID string, rating float64, note string, state RatingState) (*AlbumRatingDTO, error) {
	var noteParam sql.NullString
	if note != "" {
		noteParam = sql.NullString{String: note, Valid: true}
	}

	model, err := s.db.Queries().InsertAlbumRatingLogEntry(ctx, sqlc.InsertAlbumRatingLogEntryParams{
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

	return NewAlbumRatingDTOFromModel(model), nil
}

func (s *Service) DeleteRatingEntry(ctx context.Context, userID, entryID string) error {
	return s.db.Queries().DeleteAlbumRatingLogEntry(ctx, sqlc.DeleteAlbumRatingLogEntryParams{
		ID:     entryID,
		UserID: userID,
	})
}

func (s *Service) GetRatingLog(ctx context.Context, userID, albumID string) ([]*AlbumRatingDTO, error) {
	rows, err := s.db.Queries().GetUserAlbumRatingLog(ctx, sqlc.GetUserAlbumRatingLogParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, err
	}

	dtos := make([]*AlbumRatingDTO, len(rows))
	for i, row := range rows {
		dtos[i] = NewAlbumRatingDTOFromModel(row)
	}
	return dtos, nil
}

func (s *Service) GetRatingState(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	model, err := s.db.Queries().GetAlbumRatingState(ctx, sqlc.GetAlbumRatingStateParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return NewRatingStateDTOFromModel(model), nil
}

func (s *Service) GetAllRatingStates(ctx context.Context, userID string) (map[string]*RatingStateDTO, error) {
	rows, err := s.db.Queries().GetAllAlbumRatingStates(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*RatingStateDTO, len(rows))
	for _, row := range rows {
		dto := NewRatingStateDTOFromModel(row)
		result[dto.AlbumID] = dto
	}
	return result, nil
}

func (s *Service) CreateRatingState(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	model, err := s.db.Queries().InsertAlbumRatingState(ctx, sqlc.InsertAlbumRatingStateParams{
		ID:           uuid.NewString(),
		UserID:       userID,
		AlbumID:      albumID,
		State:        string(RatingStateProvisional),
		NextRerateAt: sql.NullTime{Time: time.Now().Add(RerateCycleDuration), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return NewRatingStateDTOFromModel(model), nil
}

func (s *Service) FinalizeRating(ctx context.Context, userID, albumID string, current *RatingStateDTO) (*RatingStateDTO, error) {
	model, err := s.db.Queries().UpdateAlbumRatingState(ctx, sqlc.UpdateAlbumRatingStateParams{
		State:        string(RatingStateFinalized),
		SnoozeCount:  int64(current.SnoozeCount),
		NextRerateAt: sql.NullTime{Time: time.Now().Add(RerateCycleDuration), Valid: true},
		UserID:       userID,
		AlbumID:      albumID,
	})
	if err != nil {
		return nil, err
	}
	return NewRatingStateDTOFromModel(model), nil
}

func (s *Service) SnoozeRating(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	current, err := s.GetRatingState(ctx, userID, albumID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrRatingStateNotFound
	}

	newSnooze := current.SnoozeCount + 1
	newState := StateAfterSnooze(*current)

	var nextRerateAt sql.NullTime
	if newState == RatingStateStalled {
		nextRerateAt = sql.NullTime{}
	} else {
		nextRerateAt = sql.NullTime{Time: time.Now().Add(SnoozeDuration), Valid: true}
	}

	model, err := s.db.Queries().UpdateAlbumRatingState(ctx, sqlc.UpdateAlbumRatingStateParams{
		State:        string(newState),
		SnoozeCount:  int64(newSnooze),
		NextRerateAt: nextRerateAt,
		UserID:       userID,
		AlbumID:      albumID,
	})
	if err != nil {
		return nil, err
	}
	return NewRatingStateDTOFromModel(model), nil
}
