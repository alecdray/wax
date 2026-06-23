package feed

import (
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/utils"
)

const (
	// MinStaleDuration is the display-freshness horizon: how old a feed's last
	// successful sync can get before the UI flags it as stale. Distinct from the
	// sync schedule (NextSyncAt), which governs when the feed is actually polled.
	MinStaleDuration = 1 * time.Hour

	// SyncInterval is the recurring cadence both Spotify feed kinds poll at when
	// healthy. MaxSyncBackoff caps the exponential backoff a failing feed grows
	// to, so a persistently-broken feed retries at most this often. See ADR 0006.
	SyncInterval   = 10 * time.Minute
	MaxSyncBackoff = 1 * time.Hour
)

type FeedDTO struct {
	ID                  string
	UserID              string
	Kind                models.FeedKind
	LastSyncStatus      models.FeedSyncStatus
	LastSyncCompletedAt *time.Time
	LastSyncStartedAt   *time.Time
	// SourceRef is the external source handle a feed kind needs, where it needs
	// one — for the radar inbox feed, the Spotify playlist id. nil otherwise.
	SourceRef *string
	// NextSyncAt is when the feed is next eligible to sync. nil means due now.
	NextSyncAt *time.Time
	// ConsecutiveFailures counts syncs that have failed in a row; it drives the
	// backoff and resets to zero on the next success.
	ConsecutiveFailures int
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
	f.ConsecutiveFailures++
	f.NextSyncAt = utils.NewPointer(time.Now().Add(backoffDuration(f.ConsecutiveFailures)))
}

func (f *FeedDTO) SetSyncSuccess() {
	f.LastSyncStatus = models.FeedSyncStatusSuccess
	f.LastSyncCompletedAt = utils.NewPointer(time.Now())
	f.ConsecutiveFailures = 0
	f.NextSyncAt = utils.NewPointer(time.Now().Add(SyncInterval))
}

func (f *FeedDTO) SetSyncing() {
	f.LastSyncStatus = models.FeedSyncStatusPending
	f.LastSyncStartedAt = utils.NewPointer(time.Now())
}

// backoffDuration returns how long to wait before retrying a feed that has
// failed `failures` times in a row: the normal interval doubled per failure,
// capped at MaxSyncBackoff. failures is expected to be >= 1.
func backoffDuration(failures int) time.Duration {
	d := SyncInterval
	for i := 1; i < failures; i++ {
		d *= 2
		if d >= MaxSyncBackoff {
			return MaxSyncBackoff
		}
	}
	return d
}
