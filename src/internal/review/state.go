package review

import "time"

type RatingState string

const (
	RatingStateProvisional RatingState = "provisional"
	RatingStateFinalized   RatingState = "finalized"
	RatingStateStalled     RatingState = "stalled"
)

type RatingStateDTO struct {
	ID           string
	AlbumID      string
	UserID       string
	State        RatingState
	SnoozeCount  int
	LastRatedAt  time.Time
	NextRerateAt *time.Time
}

const (
	MaxSnoozeCount      = 3
	SnoozeDuration      = 7 * 24 * time.Hour
	RerateCycleDuration = 30 * 24 * time.Hour
)

func (s RatingStateDTO) IsRerateDue() bool {
	if s.NextRerateAt == nil {
		return false
	}
	return s.NextRerateAt.Before(time.Now())
}

// NextRerateTime returns the time at which a rerate should next be prompted.
// snoozeCount is the number of snoozes already applied (before this call).
// Returns nil when the album is stalled (snoozeCount >= MaxSnoozeCount).
func NextRerateTime(snoozeCount int) *time.Time {
	if snoozeCount >= MaxSnoozeCount {
		return nil
	}
	t := time.Now().Add(RerateCycleDuration)
	return &t
}

// StateAfterSnooze returns the RatingState that results from applying one snooze.
// current.SnoozeCount is expected to be less than MaxSnoozeCount (i.e. snoozing is still allowed).
func StateAfterSnooze(current RatingStateDTO) RatingState {
	if current.SnoozeCount+1 >= MaxSnoozeCount {
		return RatingStateStalled
	}
	return current.State
}
