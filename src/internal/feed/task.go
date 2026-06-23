package feed

import (
	"fmt"
	"log/slog"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/task"
)

type SyncSpotifyFeedTask struct {
	feed        FeedDTO
	feedService *Service
}

var _ task.Task = SyncSpotifyFeedTask{}

func NewSyncSpotifyFeedTask(feedService *Service, feed FeedDTO) task.Task {
	return SyncSpotifyFeedTask{feed: feed, feedService: feedService}
}

func (t SyncSpotifyFeedTask) Run(ctx contextx.ContextX) error {
	_, err := t.feedService.SyncSpotifyFeed(ctx, t.feed)
	return err
}

func (t SyncSpotifyFeedTask) Schedule() *task.CronExpression {
	return nil
}

func (t SyncSpotifyFeedTask) Name() string {
	return "sync_spotify_feed"
}

type SyncStaleSpotifyFeedsTask struct {
	feedService *Service
}

var _ task.Task = SyncStaleSpotifyFeedsTask{}

func NewSyncStaleSpotifyFeedsTask(feedService *Service) task.Task {
	return SyncStaleSpotifyFeedsTask{feedService: feedService}
}

func (t SyncStaleSpotifyFeedsTask) Run(ctx contextx.ContextX) error {
	staleFeeds, err := t.feedService.GetStaleSpotifyFeeds(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get stale feeds: %w", err)
		return err
	}

	for _, feed := range staleFeeds {
		if feed.LastSyncStatus.IsSyncing() {
			continue
		}

		_, err := t.feedService.SyncSpotifyFeed(ctx, feed)
		if err != nil {
			err = fmt.Errorf("failed to sync spotify feed %s: %w", feed.ID, err)
			return err
		}

		slog.Debug("synced spotify feed", "id", feed.ID)
	}

	return nil
}

func (t SyncStaleSpotifyFeedsTask) Schedule() *task.CronExpression {
	schedule := task.CronExpression("* * * * *") // Every minute
	return &schedule
}

func (t SyncStaleSpotifyFeedsTask) Name() string {
	return "sync_stale_spotify_feeds"
}

// SyncStaleSpotifyRadarFeedsTask polls every user's radar inbox playlist each
// cron tick, ingesting added albums onto their radar.
type SyncStaleSpotifyRadarFeedsTask struct {
	feedService *Service
}

var _ task.Task = SyncStaleSpotifyRadarFeedsTask{}

func NewSyncStaleSpotifyRadarFeedsTask(feedService *Service) task.Task {
	return SyncStaleSpotifyRadarFeedsTask{feedService: feedService}
}

func (t SyncStaleSpotifyRadarFeedsTask) Run(ctx contextx.ContextX) error {
	feeds, err := t.feedService.GetSyncableRadarFeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get syncable radar feeds: %w", err)
	}

	for _, feed := range feeds {
		if feed.LastSyncStatus.IsSyncing() {
			continue
		}
		// One user's failure (e.g. a revoked token) must not block the others.
		if _, err := t.feedService.SyncSpotifyRadarFeed(ctx, feed); err != nil {
			slog.Error("failed to sync radar inbox feed", "id", feed.ID, "error", err)
			continue
		}
		slog.Debug("synced radar inbox feed", "id", feed.ID)
	}

	return nil
}

func (t SyncStaleSpotifyRadarFeedsTask) Schedule() *task.CronExpression {
	schedule := task.CronExpression("* * * * *") // Every minute
	return &schedule
}

func (t SyncStaleSpotifyRadarFeedsTask) Name() string {
	return "sync_stale_spotify_radar_feeds"
}
