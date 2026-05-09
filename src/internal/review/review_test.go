package review

import (
	"math"
	"testing"
	"time"
)

// --- BaseQuestions.Score ---

func TestBaseScore_AllMax(t *testing.T) {
	qs := allQuestions()
	for i := range qs {
		qs[i].Value = 5
	}
	got := qs.Score()
	if math.Abs(got-10.0) > 0.01 {
		t.Fatalf("expected 10.0 with all 5s, got %f", got)
	}
}

func TestBaseScore_AllMin(t *testing.T) {
	qs := allQuestions()
	for i := range qs {
		qs[i].Value = 1
	}
	got := qs.Score()
	if math.Abs(got-0.0) > 0.01 {
		t.Fatalf("expected 0.0 with all 1s, got %f", got)
	}
}

func TestBaseScore_AllMid(t *testing.T) {
	qs := allQuestions()
	for i := range qs {
		qs[i].Value = 3
	}
	got := qs.Score()
	if math.Abs(got-5.0) > 0.01 {
		t.Fatalf("expected 5.0 with all 3s, got %f", got)
	}
}

func TestBaseScore_IsRounded(t *testing.T) {
	qs := allQuestions()
	for i := range qs {
		qs[i].Value = 2
	}
	got := qs.Score()
	rounded := math.Round(got*10) / 10
	if got != rounded {
		t.Fatalf("expected score rounded to 1dp, got %f", got)
	}
}

func TestBaseScore_AllUnanswered_ReturnsZero(t *testing.T) {
	qs := allQuestions()
	got := qs.Score()
	if got != 0 {
		t.Fatalf("expected 0 for unanswered questions, got %f", got)
	}
}

// --- FinalScore ---

func TestFinalScore_ClampedAbove10(t *testing.T) {
	got := FinalScore(11.0)
	if got > 10.0 {
		t.Fatalf("expected clamped to 10.0, got %f", got)
	}
}

func TestFinalScore_ClampedBelow0(t *testing.T) {
	got := FinalScore(-1.0)
	if got < 0.0 {
		t.Fatalf("expected clamped to 0.0, got %f", got)
	}
}

// --- DetectContradictions ---

func TestDetectContradictions_HighSP_LowRR_Finalized(t *testing.T) {
	qs := allQuestions()
	for i := range qs {
		switch qs[i].Key {
		case QuestionSonicPleasure:
			qs[i].Value = 4
		case QuestionReturnRate:
			qs[i].Value = 2
		default:
			qs[i].Value = 3
		}
	}
	if !DetectContradictions(qs, RatingModeFinalized) {
		t.Fatal("expected contradiction: high SP + low RR in finalized mode")
	}
}

func TestDetectContradictions_HighSP_LowRR_Provisional_NoFlag(t *testing.T) {
	qs := allQuestions()
	for i := range qs {
		switch qs[i].Key {
		case QuestionSonicPleasure:
			qs[i].Value = 4
		case QuestionReturnRate:
			qs[i].Value = 2
		default:
			qs[i].Value = 3
		}
	}
	if DetectContradictions(qs, RatingModeProvisional) {
		t.Fatal("expected no contradiction in provisional mode")
	}
}

func TestDetectContradictions_NoContradiction(t *testing.T) {
	qs := allQuestions()
	for i := range qs {
		qs[i].Value = 3
	}
	if DetectContradictions(qs, RatingModeFinalized) {
		t.Fatal("expected no contradiction with neutral scores")
	}
}

// --- GetRatingLabel ---

func TestGetRatingLabel_Ranges(t *testing.T) {
	cases := []struct {
		rating float64
		want   RatingLabel
	}{
		{0.0, RatingLabelDOA},
		{2.9, RatingLabelDOA},
		{3.0, RatingLabelNope},
		{3.9, RatingLabelNope},
		{4.0, RatingLabelNotForMe},
		{4.9, RatingLabelNotForMe},
		{5.0, RatingLabelLukewarm},
		{5.9, RatingLabelLukewarm},
		{6.0, RatingLabelSolid},
		{6.9, RatingLabelSolid},
		{7.0, RatingLabelRecommended},
		{7.9, RatingLabelRecommended},
		{8.0, RatingLabelEssential},
		{8.9, RatingLabelEssential},
		{9.0, RatingLabelInstantClassic},
		{9.9, RatingLabelInstantClassic},
		{10.0, RatingLabelMasterpiece},
	}
	for _, tc := range cases {
		got := GetRatingLabel(tc.rating)
		if got != tc.want {
			t.Errorf("GetRatingLabel(%v) = %q, want %q", tc.rating, got, tc.want)
		}
	}
}

func TestGetRatingLabel_MidRangeFloat_NoGap(t *testing.T) {
	got := GetRatingLabel(2.95)
	if got != RatingLabelDOA {
		t.Fatalf("expected DOA for 2.95, got %q", got)
	}
}

// --- RatingStateDTO.IsRerateDue ---

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

// --- NextRerateTime ---

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

// --- StateAfterSnooze ---

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
			name        string
			state       RatingState
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

// helpers

func allQuestions() BaseQuestions {
	qs := make(BaseQuestions, len(AllBaseQuestions))
	copy(qs, AllBaseQuestions)
	return qs
}
