package feed

import (
	"context"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/alecdray/wax/src/internal/core/sqlx"

	"github.com/google/uuid"
)

// Repo is the feed module's data access layer. It is the only file in
// package feed that imports core/db/sqlc. Repo methods return feed DTOs —
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

func feedDTOFromModel(model sqlc.Feed) *FeedDTO {
	dto := &FeedDTO{
		ID:             model.ID,
		UserID:         model.UserID,
		Kind:           model.Kind,
		LastSyncStatus: model.LastSyncStatus,
	}

	if model.LastSyncStartedAt.Valid {
		dto.LastSyncStartedAt = &model.LastSyncStartedAt.Time
	}

	if model.LastSyncCompletedAt.Valid {
		dto.LastSyncCompletedAt = &model.LastSyncCompletedAt.Time
	}

	if model.SourceRef.Valid {
		dto.SourceRef = &model.SourceRef.String
	}

	return dto
}

// --- Feed lookups / mutations ---

func (r *Repo) UpsertFeed(ctx context.Context, userID string, kind models.FeedKind) (*FeedDTO, error) {
	feed, err := r.q.UpsertFeed(ctx, sqlc.UpsertFeedParams{
		ID:     uuid.New().String(),
		UserID: userID,
		Kind:   kind,
	})
	if err != nil {
		return nil, err
	}
	return feedDTOFromModel(feed), nil
}

func (r *Repo) GetFeedByID(ctx context.Context, feedID, userID string) (*FeedDTO, error) {
	feed, err := r.q.GetFeedByID(ctx, sqlc.GetFeedByIDParams{
		ID:     feedID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return feedDTOFromModel(feed), nil
}

func (r *Repo) GetFeedsByUserID(ctx context.Context, userID string) ([]FeedDTO, error) {
	feeds, err := r.q.GetFeedsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]FeedDTO, 0, len(feeds))
	for _, f := range feeds {
		out = append(out, *feedDTOFromModel(f))
	}
	return out, nil
}

func (r *Repo) UpdateFeed(ctx context.Context, feed FeedDTO) (*FeedDTO, error) {
	model, err := r.q.UpdateFeed(ctx, sqlc.UpdateFeedParams{
		ID:                  feed.ID,
		LastSyncStatus:      feed.LastSyncStatus,
		LastSyncStartedAt:   sqlx.NewNullTime(feed.LastSyncStartedAt),
		LastSyncCompletedAt: sqlx.NewNullTime(feed.LastSyncCompletedAt),
	})
	if err != nil {
		return nil, err
	}
	return feedDTOFromModel(model), nil
}

// SetFeedSourceRef stores a feed's external source handle. An empty sourceRef
// clears it (stored as NULL).
func (r *Repo) SetFeedSourceRef(ctx context.Context, feedID, sourceRef string) error {
	return r.q.SetFeedSourceRef(ctx, sqlc.SetFeedSourceRefParams{
		SourceRef: sqlx.NewNullString(sourceRef),
		ID:        feedID,
	})
}

// DeleteFeed removes a feed row.
func (r *Repo) DeleteFeed(ctx context.Context, feedID string) error {
	return r.q.DeleteFeed(ctx, feedID)
}

// GetSyncableRadarFeeds returns radar inbox feeds eligible to sync (those with a
// playlist handle), least-recently-synced first.
func (r *Repo) GetSyncableRadarFeeds(ctx context.Context) ([]FeedDTO, error) {
	feeds, err := r.q.GetSyncableRadarFeeds(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]FeedDTO, 0, len(feeds))
	for _, f := range feeds {
		out = append(out, *feedDTOFromModel(f))
	}
	return out, nil
}

// GetStaleFeedsBatch returns the batch of feeds the database considers stale
// for the given kind. Callers may apply additional filtering.
func (r *Repo) GetStaleFeedsBatch(ctx context.Context, kind models.FeedKind, minStaleDuration time.Duration) ([]FeedDTO, error) {
	feeds, err := r.q.GetStaleFeedsBatch(ctx, sqlc.GetStaleFeedsBatchParams{
		Datetime: sqlx.DurationToSQLiteDatetime(minStaleDuration),
		Kind:     kind,
	})
	if err != nil {
		return nil, err
	}

	out := make([]FeedDTO, 0, len(feeds))
	for _, f := range feeds {
		out = append(out, *feedDTOFromModel(f))
	}
	return out, nil
}
