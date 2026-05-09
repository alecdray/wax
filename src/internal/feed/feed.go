package feed

import (
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/utils"
)

const (
	MinStaleDuration = 1 * time.Hour
)

type FeedDTO struct {
	ID                  string
	UserID              string
	Kind                models.FeedKind
	LastSyncStatus      models.FeedSyncStatus
	LastSyncCompletedAt *time.Time
	LastSyncStartedAt   *time.Time
}

func (f FeedDTO) IsSyncStale() bool {
	if f.LastSyncStatus.IsUnsyned() {
		return false
	}
	if f.LastSyncCompletedAt == nil {
		return true
	}
	minStaleTime := time.Now().Add(-MinStaleDuration)
	return f.LastSyncCompletedAt.Before(minStaleTime)
}

func (f *FeedDTO) SetSyncFailed() {
	f.LastSyncStatus = models.FeedSyncStatusFailure
}

func (f *FeedDTO) SetSyncSuccess() {
	f.LastSyncStatus = models.FeedSyncStatusSuccess
	f.LastSyncCompletedAt = utils.NewPointer(time.Now())
}

func (f *FeedDTO) SetSyncing() {
	f.LastSyncStatus = models.FeedSyncStatusPending
	f.LastSyncStartedAt = utils.NewPointer(time.Now())
}
