package listeninghistory

import (
	"errors"
	"fmt"
	"log/slog"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/task"
	"shmoopicks/src/internal/spotify"
)

type SyncListeningHistoryTask struct {
	service *Service
}

var _ task.Task = SyncListeningHistoryTask{}

func NewSyncListeningHistoryTask(service *Service) task.Task {
	return SyncListeningHistoryTask{service: service}
}

func (t SyncListeningHistoryTask) Run(ctx contextx.ContextX) error {
	users, err := t.service.db.Queries().GetUsersWithSpotifyToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users with spotify token: %w", err)
	}

	for _, user := range users {
		if err := t.service.SyncUser(ctx, user.ID); err != nil {
			if errors.Is(err, spotify.ErrFailedToGetToken) {
				slog.Warn("skipping listening history sync: token error", "userId", user.ID, "error", err)
				continue
			}
			slog.Error("failed to sync listening history", "userId", user.ID, "error", err)
			continue
		}
		slog.Debug("synced listening history", "userId", user.ID)
	}

	return nil
}

func (t SyncListeningHistoryTask) Schedule() *task.CronExpression {
	schedule := task.CronExpression("* * * * *") // Every hour
	return &schedule
}

func (t SyncListeningHistoryTask) Name() string {
	return "sync_listening_history"
}
