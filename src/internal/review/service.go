package review

import (
	"context"
	"fmt"

	"github.com/alecdray/wax/src/internal/core/db"
)

// ErrFinalizeRequiresProvisional is returned by FinalizeWithRating when called
// for an album whose current rating-state is not provisional. The promotion to
// finalized is only meaningful from a provisional starting state — finalized
// albums stay finalized via the regular save path, and unrated albums cannot
// be finalized directly.
var ErrFinalizeRequiresProvisional = fmt.Errorf("finalize requires a provisional rating state")

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

// FinalizeWithRating writes a new rating-log entry and promotes the album's
// rating state from provisional to finalized in a single call. The current
// state must be provisional — calls on an unrated or already-finalized album
// return ErrFinalizeRequiresProvisional and write nothing.
func (s *Service) FinalizeWithRating(ctx context.Context, userID, albumID string, rating float64, note string) (*AlbumRatingDTO, *RatingStateDTO, error) {
	currentState, err := s.repo.GetAlbumRatingState(ctx, userID, albumID)
	if err != nil {
		return nil, nil, err
	}
	if currentState == nil || currentState.State != RatingStateProvisional {
		return nil, nil, ErrFinalizeRequiresProvisional
	}

	logEntry, err := s.repo.InsertAlbumRatingLogEntry(ctx, userID, albumID, rating, note, RatingStateFinalized)
	if err != nil {
		return nil, nil, err
	}

	newState, err := s.repo.UpdateAlbumRatingState(ctx, userID, albumID, RatingStateFinalized)
	if err != nil {
		return nil, nil, err
	}
	return logEntry, newState, nil
}
