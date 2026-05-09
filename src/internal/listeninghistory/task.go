package listeninghistory

import (
	"errors"
	"fmt"
	"log/slog"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/task"
	"github.com/alecdray/wax/src/internal/spotify"
)

type SyncListeningHistoryTask struct {
	service *Service
}

var _ task.Task = SyncListeningHistoryTask{}

func NewSyncListeningHistoryTask(service *Service) task.Task {
	return SyncListeningHistoryTask{service: service}
}

func (t SyncListeningHistoryTask) Run(ctx contextx.ContextX) error {
	userIDs, err := t.service.GetUserIDsWithSpotifyToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users with spotify token: %w", err)
	}

	for _, userID := range userIDs {
		if err := t.service.SyncUser(ctx, userID); err != nil {
			if errors.Is(err, spotify.ErrFailedToGetToken) {
				slog.Warn("skipping listening history sync: token error", "userId", userID, "error", err)
				continue
			}
			slog.Error("failed to sync listening history", "userId", userID, "error", err)
			continue
		}
		slog.Debug("synced listening history", "userId", userID)
	}

	return nil
}

func (t SyncListeningHistoryTask) Schedule() *task.CronExpression {
	schedule := task.CronExpression("0 * * * *") // Every hour
	return &schedule
}

func (t SyncListeningHistoryTask) Name() string {
	return "sync_listening_history"
}
