package review

import (
	"testing"
	"time"
)

func TestIsRerateDue(t *testing.T) {
	now := time.Now()

	t.Run("returns true when NextRerateAt is in the past", func(t *testing.T) {
		past := now.Add(-1 * time.Hour)
		state := RatingStateDTO{
			NextRerateAt: &past,
		}
		if !state.IsRerateDue() {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns false when NextRerateAt is nil", func(t *testing.T) {
		state := RatingStateDTO{
			NextRerateAt: nil,
		}
		if state.IsRerateDue() {
			t.Error("expected false, got true")
		}
	})

	t.Run("returns false when NextRerateAt is in the future", func(t *testing.T) {
		future := now.Add(1 * time.Hour)
		state := RatingStateDTO{
			NextRerateAt: &future,
		}
		if state.IsRerateDue() {
			t.Error("expected false, got true")
		}
	})
}

func TestNextRerateTime(t *testing.T) {
	t.Run("returns nil when snoozeCount >= MaxSnoozeCount", func(t *testing.T) {
		for snoozeCount := MaxSnoozeCount; snoozeCount <= MaxSnoozeCount+2; snoozeCount++ {
			result := NextRerateTime(snoozeCount)
			if result != nil {
				t.Errorf("snoozeCount=%d: expected nil, got %v", snoozeCount, result)
			}
		}
	})

	t.Run("returns non-nil time for counts less than MaxSnoozeCount", func(t *testing.T) {
		for snoozeCount := 0; snoozeCount < MaxSnoozeCount; snoozeCount++ {
			result := NextRerateTime(snoozeCount)
			if result == nil {
				t.Errorf("snoozeCount=%d: expected non-nil, got nil", snoozeCount)
			}
		}
	})

	t.Run("returned time is approximately RerateCycleDuration in the future", func(t *testing.T) {
		before := time.Now()
		result := NextRerateTime(0)
		after := time.Now()

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		expectedMin := before.Add(RerateCycleDuration)
		expectedMax := after.Add(RerateCycleDuration)

		if result.Before(expectedMin) || result.After(expectedMax) {
			t.Errorf("returned time %v not in expected range [%v, %v]", result, expectedMin, expectedMax)
		}
	})
}

func TestStateAfterSnooze(t *testing.T) {
	t.Run("returns Stalled when snooze would hit max", func(t *testing.T) {
		state := RatingStateDTO{
			State:       RatingStateProvisional,
			SnoozeCount: MaxSnoozeCount - 1,
		}
		result := StateAfterSnooze(state)
		if result != RatingStateStalled {
			t.Errorf("expected %q, got %q", RatingStateStalled, result)
		}
	})

	t.Run("returns same state when below snooze threshold", func(t *testing.T) {
		tests := []struct {
			name       string
			state      RatingState
			snoozeCount int
		}{
			{"Provisional with count 0", RatingStateProvisional, 0},
			{"Provisional with count 1", RatingStateProvisional, 1},
			{"Finalized with count 0", RatingStateFinalized, 0},
			{"Finalized with count 1", RatingStateFinalized, 1},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				state := RatingStateDTO{
					State:       tt.state,
					SnoozeCount: tt.snoozeCount,
				}
				result := StateAfterSnooze(state)
				if result != tt.state {
					t.Errorf("expected %q, got %q", tt.state, result)
				}
			})
		}
	})
}
