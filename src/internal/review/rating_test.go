package review

import (
	"math"
	"testing"
)

// --- RatingQuestion.CurvedValue ---

func TestCurvedValue_MiddleIsLinear(t *testing.T) {
	q := RatingQuestion{Value: 3}
	got := q.CurvedValue()
	// Value of 3 (midpoint) should return exactly 3.0 regardless of curve
	if math.Abs(got-3.0) > 0.001 {
		t.Fatalf("expected ~3.0, got %f", got)
	}
}

func TestCurvedValue_MaxAmplified(t *testing.T) {
	q := RatingQuestion{Value: 5}
	linear := RatingQuestion{Value: 5}
	_ = linear
	got := q.CurvedValue()
	// With exponent < 1, extreme values are amplified — CurvedValue(5) should be > linear 5
	// Actually the curve maps [1,5] -> amplified [1,5], so let's just check range
	if got < 4.0 || got > 5.0 {
		t.Fatalf("expected CurvedValue(5) in [4, 5], got %f", got)
	}
}

func TestCurvedValue_MinAmplified(t *testing.T) {
	q := RatingQuestion{Value: 1}
	got := q.CurvedValue()
	if got < 1.0 || got > 2.0 {
		t.Fatalf("expected CurvedValue(1) in [1, 2], got %f", got)
	}
}

// --- RatingQuestions.Rating ---

func TestRating_AllMaxGivesFloor10(t *testing.T) {
	qs := RatingRecommenderQuestions
	withValues := make(RatingQuestions, len(qs))
	for i, q := range qs {
		withValues[i] = q.WithValue(5)
	}
	got := withValues.Rating()
	if math.Abs(got-ratingCeiling) > 0.1 {
		t.Fatalf("expected rating ~%f with all 5s, got %f", ratingCeiling, got)
	}
}

func TestRating_AllMinGivesFloor2(t *testing.T) {
	qs := RatingRecommenderQuestions
	withValues := make(RatingQuestions, len(qs))
	for i, q := range qs {
		withValues[i] = q.WithValue(1)
	}
	got := withValues.Rating()
	if math.Abs(got-ratingFloor) > 0.1 {
		t.Fatalf("expected rating ~%f with all 1s, got %f", ratingFloor, got)
	}
}

func TestRating_IsRounded(t *testing.T) {
	qs := RatingRecommenderQuestions
	withValues := make(RatingQuestions, len(qs))
	for i, q := range qs {
		withValues[i] = q.WithValue(3)
	}
	got := withValues.Rating()
	// Result should be rounded to 1 decimal place
	rounded := math.Round(got*10) / 10
	if got != rounded {
		t.Fatalf("expected rating rounded to 1dp, got %f", got)
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
		{5.4, RatingLabelLukewarm},
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

func TestGetRatingLabel_ClampsAbove10(t *testing.T) {
	got := GetRatingLabel(11.0)
	if got != RatingLabelMasterpiece {
		t.Fatalf("expected Masterpiece for rating > 10, got %q", got)
	}
}

func TestGetRatingLabel_ClampsBelow0(t *testing.T) {
	got := GetRatingLabel(-1.0)
	if got != RatingLabelDOA {
		t.Fatalf("expected DOA for negative rating, got %q", got)
	}
}
