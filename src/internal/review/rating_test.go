package review

import (
	"math"
	"testing"
)

// --- BaseQuestions.Score ---

func TestBaseScore_AllMax_Finalized(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 5
	}
	got := qs.Score(RatingModeFinalized)
	if math.Abs(got-10.0) > 0.01 {
		t.Fatalf("expected 10.0 with all 5s (finalized), got %f", got)
	}
}

func TestBaseScore_AllMin_Finalized(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 1
	}
	got := qs.Score(RatingModeFinalized)
	if math.Abs(got-0.0) > 0.01 {
		t.Fatalf("expected 0.0 with all 1s (finalized), got %f", got)
	}
}

func TestBaseScore_AllMid_Finalized(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 3
	}
	got := qs.Score(RatingModeFinalized)
	if math.Abs(got-5.0) > 0.01 {
		t.Fatalf("expected 5.0 with all 3s (finalized), got %f", got)
	}
}

func TestBaseScore_Provisional_ExcludesAttachmentQuestions(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		switch qs[i].Key {
		case QuestionReturnRate, QuestionShelfTest:
			qs[i].Value = 1 // would drag score down if included
		default:
			qs[i].Value = 5
		}
	}
	got := qs.Score(RatingModeProvisional)
	// ReturnRate and ShelfTest are excluded in provisional — only TQ, Cohesion, ER, SP contribute
	// all at 5, so score should be 10
	if math.Abs(got-10.0) > 0.01 {
		t.Fatalf("expected 10.0 with attachment questions excluded, got %f", got)
	}
}

func TestFinalScore_Provisional_CappedAt8(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 5
	}
	// Base score alone is uncapped now; cap applied in FinalScore
	base := qs.Score(RatingModeProvisional)
	got := FinalScore(base, 0, RatingModeProvisional)
	if got > ProvisionalScoreCap {
		t.Fatalf("provisional final score %f exceeds cap %f", got, ProvisionalScoreCap)
	}
}

func TestBaseScore_IsRounded(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 2
	}
	got := qs.Score(RatingModeFinalized)
	rounded := math.Round(got*10) / 10
	if got != rounded {
		t.Fatalf("expected score rounded to 1dp, got %f", got)
	}
}

// --- Modifiers.Adjustment ---

func TestModifierAdjustment_AllPositive(t *testing.T) {
	mods := defaultModifiers()
	for i := range mods {
		mods[i].Value = 1
	}
	got := mods.Adjustment()
	if math.Abs(got-ModifierMaxSwing) > 0.001 {
		t.Fatalf("expected +%f, got %f", ModifierMaxSwing, got)
	}
}

func TestModifierAdjustment_AllNegative(t *testing.T) {
	mods := defaultModifiers()
	for i := range mods {
		mods[i].Value = -1
	}
	got := mods.Adjustment()
	if math.Abs(got-(-ModifierMaxSwing)) > 0.001 {
		t.Fatalf("expected -%f, got %f", ModifierMaxSwing, got)
	}
}

func TestModifierAdjustment_Mixed_Dampens(t *testing.T) {
	mods := defaultModifiers()
	mods[0].Value = 1
	mods[1].Value = -1
	got := mods.Adjustment()
	if math.Abs(got) > 0.001 {
		t.Fatalf("expected 0 for mixed modifiers, got %f", got)
	}
}

// --- FinalScore ---

func TestFinalScore_ClampedAbove10(t *testing.T) {
	got := FinalScore(10.0, ModifierMaxSwing, RatingModeFinalized)
	if got > 10.0 {
		t.Fatalf("expected clamped to 10.0, got %f", got)
	}
}

func TestFinalScore_ClampedBelow0(t *testing.T) {
	got := FinalScore(0.0, -ModifierMaxSwing, RatingModeFinalized)
	if got < 0.0 {
		t.Fatalf("expected clamped to 0.0, got %f", got)
	}
}

func TestBaseScore_Provisional_CappedAt8_AfterModifiers(t *testing.T) {
	// All max answers in provisional mode gives base of 10.0 before cap
	// but with positive modifiers, FinalScore should still be capped at 8.0
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 5
	}
	mods := defaultModifiers()
	for i := range mods {
		mods[i].Value = 1 // all positive
	}
	base := qs.Score(RatingModeProvisional)
	modAdj := mods.Adjustment()
	got := FinalScore(base, modAdj, RatingModeProvisional)
	if got > ProvisionalScoreCap {
		t.Fatalf("expected provisional final score capped at %f, got %f", ProvisionalScoreCap, got)
	}
}

// --- DetectContradictions ---

func TestDetectContradictions_HighERAndSP_LowRR_Finalized(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		switch qs[i].Key {
		case QuestionEmotionalResonance, QuestionSonicPleasure:
			qs[i].Value = 4
		case QuestionReturnRate:
			qs[i].Value = 2
		default:
			qs[i].Value = 3
		}
	}
	mods := defaultModifiers()
	if !DetectContradictions(qs, mods, 5.0, RatingModeFinalized) {
		t.Fatal("expected contradiction detected")
	}
}

func TestDetectContradictions_HighERAndSP_LowRR_Provisional_NoFlag(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		switch qs[i].Key {
		case QuestionEmotionalResonance, QuestionSonicPleasure:
			qs[i].Value = 4
		default:
			qs[i].Value = 3
		}
	}
	mods := defaultModifiers()
	if DetectContradictions(qs, mods, 5.0, RatingModeProvisional) {
		t.Fatal("expected no contradiction in provisional (return rate excluded)")
	}
}


func TestDetectContradictions_HighScore_AllNegativeMods(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 3
	}
	mods := defaultModifiers()
	for i := range mods {
		mods[i].Value = -1
	}
	if !DetectContradictions(qs, mods, 7.5, RatingModeFinalized) {
		t.Fatal("expected contradiction: high base score + all negative mods")
	}
}

func TestDetectContradictions_NoContradiction(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 3
	}
	mods := defaultModifiers()
	if DetectContradictions(qs, mods, 5.0, RatingModeFinalized) {
		t.Fatal("expected no contradiction with mid scores and neutral mods")
	}
}

func TestDetectContradictions_LowScore_AllNegativeMods_NoFlag(t *testing.T) {
	// Contradiction 3 should NOT fire when base score < 7.0
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 2 // low scores → base well below 7.0
	}
	mods := defaultModifiers()
	for i := range mods {
		mods[i].Value = -1
	}
	if DetectContradictions(qs, mods, 5.0, RatingModeFinalized) {
		t.Fatal("expected no contradiction when base score < 7.0")
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

func TestBaseScore_AllUnanswered_ReturnsZero(t *testing.T) {
	qs := finalizedQuestions()
	// all Value fields stay 0 (unanswered)
	got := qs.Score(RatingModeFinalized)
	if got != 0 {
		t.Fatalf("expected 0 for unanswered questions, got %f", got)
	}
}

func TestGetRatingLabel_MidRangeFloat_NoGap(t *testing.T) {
	// 2.95 is between boundary entries — should not return Masterpiece
	got := GetRatingLabel(2.95)
	if got == RatingLabelMasterpiece {
		t.Fatal("GetRatingLabel(2.95) should not return Masterpiece")
	}
	if got != RatingLabelDOA {
		t.Fatalf("expected DOA for 2.95, got %q", got)
	}
}

func TestDetectContradictions_EmptyMods_NoFlag(t *testing.T) {
	qs := finalizedQuestions()
	for i := range qs {
		qs[i].Value = 5
	}
	var emptyMods Modifiers
	if DetectContradictions(qs, emptyMods, 9.0, RatingModeFinalized) {
		t.Fatal("expected no contradiction with empty mods slice")
	}
}

func TestModifierAdjustment_EmptySlice_ReturnsZero(t *testing.T) {
	var mods Modifiers
	got := mods.Adjustment()
	if got != 0 {
		t.Fatalf("expected 0 for empty modifiers, got %f", got)
	}
}

// helpers

func finalizedQuestions() BaseQuestions {
	qs := make(BaseQuestions, len(AllBaseQuestions))
	copy(qs, AllBaseQuestions)
	return qs
}

func defaultModifiers() Modifiers {
	ms := make(Modifiers, len(AllModifiers))
	copy(ms, AllModifiers)
	return ms
}
