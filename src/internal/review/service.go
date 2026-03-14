package review

import (
	"context"
	"database/sql"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"time"

	"github.com/google/uuid"
)

type AlbumRatingDTO struct {
	ID        string
	UserID    string
	AlbumID   string
	Rating    *float64
	Note      *string
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

func (s *Service) AddRating(ctx context.Context, userId, albumId string, rating float64, note string) (*AlbumRatingDTO, error) {
	var noteParam sql.NullString
	if note != "" {
		noteParam = sql.NullString{String: note, Valid: true}
	}
	model, err := s.db.Queries().InsertAlbumRatingLogEntry(ctx, sqlc.InsertAlbumRatingLogEntryParams{
		ID:      uuid.NewString(),
		UserID:  userId,
		AlbumID: albumId,
		Rating:  rating,
		Note:    noteParam,
	})
	if err != nil {
		return nil, err
	}

	return NewAlbumRatingDTOFromModel(model), nil
}

func (s *Service) DeleteRatingEntry(ctx context.Context, userId, entryId string) error {
	return s.db.Queries().DeleteAlbumRatingLogEntry(ctx, sqlc.DeleteAlbumRatingLogEntryParams{
		ID:     entryId,
		UserID: userId,
	})
}

func (s *Service) GetRatingLog(ctx context.Context, userId, albumId string) ([]*AlbumRatingDTO, error) {
	rows, err := s.db.Queries().GetUserAlbumRatingLog(ctx, sqlc.GetUserAlbumRatingLogParams{
		UserID:  userId,
		AlbumID: albumId,
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
