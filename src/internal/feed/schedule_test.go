package feed

import (
	"testing"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/utils"
)

func TestBackoffDuration(t *testing.T) {
	cases := []struct {
		failures int
		want     time.Duration
	}{
		{1, SyncInterval},     // 10m
		{2, 2 * SyncInterval}, // 20m
		{3, 4 * SyncInterval}, // 40m
		{4, MaxSyncBackoff},   // 80m -> capped at 60m
		{5, MaxSyncBackoff},   // stays capped
		{20, MaxSyncBackoff},  // far past the cap
	}
	for _, c := range cases {
		if got := backoffDuration(c.failures); got != c.want {
			t.Errorf("backoffDuration(%d) = %s, want %s", c.failures, got, c.want)
		}
	}
}

func TestSetSyncFailedBacksOffAndCounts(t *testing.T) {
	completedAt := time.Now().Add(-2 * time.Hour)
	f := &FeedDTO{
		Kind:                models.FeedKindSpotify,
		LastSyncCompletedAt: utils.NewPointer(completedAt),
		ConsecutiveFailures: 1,
	}

	before := time.Now()
	f.SetSyncFailed()

	if f.LastSyncStatus != models.FeedSyncStatusFailure {
		t.Fatalf("status = %v, want failure", f.LastSyncStatus)
	}
	if f.ConsecutiveFailures != 2 {
		t.Fatalf("ConsecutiveFailures = %d, want 2", f.ConsecutiveFailures)
	}
	// A failure must not advance the last *successful* completion timestamp.
	if !f.LastSyncCompletedAt.Equal(completedAt) {
		t.Fatalf("LastSyncCompletedAt changed on failure: %v", f.LastSyncCompletedAt)
	}
	// next_sync_at ~= now + backoff(2) = now + 20m.
	wantNext := before.Add(backoffDuration(2))
	if f.NextSyncAt == nil || f.NextSyncAt.Sub(wantNext).Abs() > time.Second {
		t.Fatalf("NextSyncAt = %v, want ~%v", f.NextSyncAt, wantNext)
	}
}

func TestSetSyncSuccessResetsAndSchedulesNormalCadence(t *testing.T) {
	f := &FeedDTO{
		Kind:                models.FeedKindSpotify,
		ConsecutiveFailures: 3,
	}

	before := time.Now()
	f.SetSyncSuccess()

	if f.LastSyncStatus != models.FeedSyncStatusSuccess {
		t.Fatalf("status = %v, want success", f.LastSyncStatus)
	}
	if f.ConsecutiveFailures != 0 {
		t.Fatalf("ConsecutiveFailures = %d, want 0 after success", f.ConsecutiveFailures)
	}
	if f.LastSyncCompletedAt == nil || f.LastSyncCompletedAt.Before(before) {
		t.Fatalf("LastSyncCompletedAt not advanced: %v", f.LastSyncCompletedAt)
	}
	wantNext := before.Add(SyncInterval)
	if f.NextSyncAt == nil || f.NextSyncAt.Sub(wantNext).Abs() > time.Second {
		t.Fatalf("NextSyncAt = %v, want ~%v", f.NextSyncAt, wantNext)
	}
}
