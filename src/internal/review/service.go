package review

import (
	"context"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/db/sqlc"
	"shmoopicks/src/internal/core/sqlx"

	"github.com/google/uuid"
)

type AlbumRatingDTO struct {
	ID      string
	UserID  string
	AlbumID string
	Rating  *float64
	Review  *string
}

func NewAlbumRatingDTOFromModel(model sqlc.AlbumRating) *AlbumRatingDTO {
	dto := &AlbumRatingDTO{
		ID:      model.ID,
		UserID:  model.UserID,
		AlbumID: model.AlbumID,
	}

	if model.Rating.Valid {
		dto.Rating = &model.Rating.Float64
	}

	if model.Review.Valid {
		dto.Review = &model.Review.String
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

func (s *Service) UpdateRating(ctx context.Context, userId, albumId string, rating float64) (*AlbumRatingDTO, error) {
	model, err := s.db.Queries().UpsertAlbumRating(ctx, sqlc.UpsertAlbumRatingParams{
		ID:      uuid.NewString(),
		UserID:  userId,
		AlbumID: albumId,
		Rating:  sqlx.NewNullFloat64(rating),
	})
	if err != nil {
		return nil, err
	}

	return NewAlbumRatingDTOFromModel(model), nil
}

func (s *Service) UpdateReview(ctx context.Context, userId, albumId string, review string) (*AlbumRatingDTO, error) {
	model, err := s.db.Queries().UpsertAlbumReview(ctx, sqlc.UpsertAlbumReviewParams{
		ID:      uuid.NewString(),
		UserID:  userId,
		AlbumID: albumId,
		Review:  sqlx.NewNullString(review),
	})
	if err != nil {
		return nil, err
	}

	return NewAlbumRatingDTOFromModel(model), nil
}
