package review

import (
	"context"
	"database/sql"
	"math"
	"path/filepath"
	"testing"

	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
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

// --- RatingStateLogLabel ---

func TestRatingStateLogLabel(t *testing.T) {
	cases := []struct {
		in   RatingState
		want string
	}{
		{RatingStateProvisional, "Provisional"},
		{RatingStateFinalized, "Finalized"},
		{"stalled", "Stalled"},
	}
	for _, tc := range cases {
		got := RatingStateLogLabel(tc.in)
		if got != tc.want {
			t.Errorf("RatingStateLogLabel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// --- FinalizeWithRating ---

func TestFinalizeWithRating_PromotesProvisionalAndWritesLog(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()

	seedUserAndAlbum(t, sqlDB, "user-1", "album-1")
	if _, err := svc.CreateRatingState(ctx, "user-1", "album-1"); err != nil {
		t.Fatalf("seed provisional state: %v", err)
	}

	logEntry, newState, err := svc.FinalizeWithRating(ctx, "user-1", "album-1", 7.4, "")
	if err != nil {
		t.Fatalf("FinalizeWithRating: %v", err)
	}
	if logEntry == nil || logEntry.Rating == nil || *logEntry.Rating != 7.4 {
		t.Fatalf("expected log entry with rating 7.4, got %+v", logEntry)
	}
	if newState == nil || newState.State != RatingStateFinalized {
		t.Fatalf("expected state to be finalized, got %+v", newState)
	}

	log, err := svc.GetRatingLog(ctx, "user-1", "album-1")
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if len(log) != 1 {
		t.Fatalf("expected exactly one log row, got %d", len(log))
	}

	st, err := svc.GetRatingState(ctx, "user-1", "album-1")
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if st == nil || st.State != RatingStateFinalized {
		t.Fatalf("expected finalized state, got %+v", st)
	}
}

func TestFinalizeWithRating_FinalizesUnratedAlbumInOneStep(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "user-1", "album-1")

	logEntry, newState, err := svc.FinalizeWithRating(ctx, "user-1", "album-1", 5.0, "")
	if err != nil {
		t.Fatalf("FinalizeWithRating: %v", err)
	}
	if logEntry == nil || logEntry.Rating == nil || *logEntry.Rating != 5.0 {
		t.Fatalf("expected log entry with rating 5.0, got %+v", logEntry)
	}
	if newState == nil || newState.State != RatingStateFinalized {
		t.Fatalf("expected finalized state, got %+v", newState)
	}
	log, err := svc.GetRatingLog(ctx, "user-1", "album-1")
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if len(log) != 1 {
		t.Fatalf("expected exactly one log row, got %d", len(log))
	}
}

func TestFinalizeWithRating_OnFinalizedAlbumStaysFinalizedAndAppendsLog(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "user-1", "album-1")
	if _, _, err := svc.FinalizeWithRating(ctx, "user-1", "album-1", 8.0, ""); err != nil {
		t.Fatalf("seed finalize: %v", err)
	}

	if _, st, err := svc.FinalizeWithRating(ctx, "user-1", "album-1", 9.0, ""); err != nil {
		t.Fatalf("re-finalize: %v", err)
	} else if st == nil || st.State != RatingStateFinalized {
		t.Fatalf("expected finalized, got %+v", st)
	}
	log, err := svc.GetRatingLog(ctx, "user-1", "album-1")
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if len(log) != 2 {
		t.Fatalf("expected two log rows, got %d", len(log))
	}
}

func TestSaveRating_OnUnratedCreatesProvisional(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "user-1", "album-1")

	if _, st, err := svc.SaveRating(ctx, "user-1", "album-1", 6.0, ""); err != nil {
		t.Fatalf("SaveRating: %v", err)
	} else if st == nil || st.State != RatingStateProvisional {
		t.Fatalf("expected provisional, got %+v", st)
	}
}

func TestSaveRating_OnFinalizedDemotesToProvisional(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "user-1", "album-1")
	if _, _, err := svc.FinalizeWithRating(ctx, "user-1", "album-1", 8.0, ""); err != nil {
		t.Fatalf("seed finalize: %v", err)
	}

	if _, st, err := svc.SaveRating(ctx, "user-1", "album-1", 7.0, ""); err != nil {
		t.Fatalf("SaveRating: %v", err)
	} else if st == nil || st.State != RatingStateProvisional {
		t.Fatalf("save must demote finalized to provisional, got %+v", st)
	}
}

// helpers

func allQuestions() BaseQuestions {
	qs := make(BaseQuestions, len(AllBaseQuestions))
	copy(qs, AllBaseQuestions)
	return qs
}

// newTestService opens a fresh sqlite DB, applies every migration in the repo,
// and returns a Service plus the raw *sql.DB for fixture seeding.
func newTestService(t *testing.T) (*Service, *sql.DB) {
	t.Helper()

	migrationsDir, err := filepath.Abs("../../../db/migrations")
	if err != nil {
		t.Fatalf("resolve migrations dir: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	return NewService(db.WrapSqlDB(sqlDB)), sqlDB
}

func seedUserAndAlbum(t *testing.T, sqlDB *sql.DB, userID, albumID string) {
	t.Helper()
	if _, err := sqlDB.Exec(`INSERT INTO users (id, spotify_id) VALUES (?, ?)`, userID, "spotify-"+userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO albums (id, spotify_id, title) VALUES (?, ?, ?)`, albumID, "spotify-"+albumID, "Album "+albumID); err != nil {
		t.Fatalf("seed album: %v", err)
	}
}
