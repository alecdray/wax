package review

import (
	"context"

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

func (s *Service) AddRating(ctx context.Context, userID, albumID string, rating float64, note string, state RatingState) (*AlbumRatingDTO, error) {
	return s.repo.InsertAlbumRatingLogEntry(ctx, userID, albumID, rating, note, state)
}

func (s *Service) DeleteRatingEntry(ctx context.Context, userID, entryID string) error {
	return s.repo.DeleteAlbumRatingLogEntry(ctx, userID, entryID)
}

func (s *Service) GetRatingLog(ctx context.Context, userID, albumID string) ([]*AlbumRatingDTO, error) {
	return s.repo.GetUserAlbumRatingLog(ctx, userID, albumID)
}

func (s *Service) GetLatestRating(ctx context.Context, userID, albumID string) (*AlbumRatingDTO, error) {
	return s.repo.GetLatestUserAlbumRating(ctx, userID, albumID)
}

func (s *Service) GetLatestRatings(ctx context.Context, userID string) (map[string]AlbumRatingDTO, error) {
	return s.repo.GetLatestUserAlbumRatings(ctx, userID)
}

func (s *Service) GetRatingState(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	return s.repo.GetAlbumRatingState(ctx, userID, albumID)
}

func (s *Service) GetAllRatingStates(ctx context.Context, userID string) (map[string]*RatingStateDTO, error) {
	return s.repo.GetAllAlbumRatingStates(ctx, userID)
}

func (s *Service) CreateRatingState(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	return s.repo.InsertAlbumRatingState(ctx, userID, albumID, RatingStateProvisional)
}

func (s *Service) FinalizeRating(ctx context.Context, userID, albumID string) (*RatingStateDTO, error) {
	return s.repo.UpdateAlbumRatingState(ctx, userID, albumID, RatingStateFinalized)
}
