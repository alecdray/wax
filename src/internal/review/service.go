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

// setRatingState upserts the album's rating-state row to the given lifecycle
// value: insert a fresh row when none exists, update in place otherwise. The
// resulting state is therefore a pure function of the value passed in,
// independent of the prior state.
func (s *Service) setRatingState(ctx context.Context, userID, albumID string, state RatingState) (*RatingStateDTO, error) {
	current, err := s.repo.GetAlbumRatingState(ctx, userID, albumID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return s.repo.InsertAlbumRatingState(ctx, userID, albumID, state)
	}
	return s.repo.UpdateAlbumRatingState(ctx, userID, albumID, state)
}

// SaveRating writes a new rating-log entry with the supplied score and sets the
// album's rating state to provisional, from any prior state. Saving a finalized
// album demotes it to provisional — the save action is the only un-finalize
// path. The resulting state is always provisional.
func (s *Service) SaveRating(ctx context.Context, userID, albumID string, rating float64, note string) (*AlbumRatingDTO, *RatingStateDTO, error) {
	logEntry, err := s.repo.InsertAlbumRatingLogEntry(ctx, userID, albumID, rating, note, RatingStateProvisional)
	if err != nil {
		return nil, nil, err
	}
	newState, err := s.setRatingState(ctx, userID, albumID, RatingStateProvisional)
	if err != nil {
		return nil, nil, err
	}
	return logEntry, newState, nil
}

// FinalizeWithRating writes a new rating-log entry with the supplied score and
// sets the album's rating state to finalized, from any prior state — unrated,
// provisional, or already finalized. The resulting state is always finalized.
func (s *Service) FinalizeWithRating(ctx context.Context, userID, albumID string, rating float64, note string) (*AlbumRatingDTO, *RatingStateDTO, error) {
	logEntry, err := s.repo.InsertAlbumRatingLogEntry(ctx, userID, albumID, rating, note, RatingStateFinalized)
	if err != nil {
		return nil, nil, err
	}
	newState, err := s.setRatingState(ctx, userID, albumID, RatingStateFinalized)
	if err != nil {
		return nil, nil, err
	}
	return logEntry, newState, nil
}
