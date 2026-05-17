package review

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

// These tests exercise the assembled review service against a real sqlite DB
// to demonstrate the system-level rating-lifecycle invariants that the
// rework's modal / handler / state-machine pieces all must agree on.

// --- Manual-transition state machine ---
//
// The live state machine has exactly four legitimate transitions:
//   (no row)      -> provisional   via CreateRatingState (implicit on first save)
//   provisional   -> finalized     via FinalizeRating / FinalizeWithRating
//   provisional   -> provisional   via AddRating (re-rate save)
//   finalized     -> finalized     via AddRating (re-rate save)
//
// Nothing in the live system writes 'stalled' to album_rating_state.state, and
// nothing transitions provisional->finalized on a time trigger.

func TestRatingLifecycle_FirstSaveCreatesProvisional(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "u1", "a1")

	// Pre-condition: no state row exists.
	state, err := svc.GetRatingState(ctx, "u1", "a1")
	if err != nil {
		t.Fatalf("pre-check state: %v", err)
	}
	if state != nil {
		t.Fatalf("expected no state row before first save, got %+v", state)
	}

	// First save: mirror the handler's behaviour (AddRating then CreateRatingState
	// when no state row exists).
	if _, err := svc.AddRating(ctx, "u1", "a1", 7.0, "", RatingStateProvisional); err != nil {
		t.Fatalf("AddRating: %v", err)
	}
	if _, err := svc.CreateRatingState(ctx, "u1", "a1"); err != nil {
		t.Fatalf("CreateRatingState: %v", err)
	}

	got, err := svc.GetRatingState(ctx, "u1", "a1")
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if got == nil || got.State != RatingStateProvisional {
		t.Fatalf("expected state=provisional after first save, got %+v", got)
	}
}

func TestRatingLifecycle_RerateLeavesProvisionalUntouched(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "u1", "a1")

	if _, err := svc.AddRating(ctx, "u1", "a1", 5.0, "", RatingStateProvisional); err != nil {
		t.Fatalf("initial AddRating: %v", err)
	}
	if _, err := svc.CreateRatingState(ctx, "u1", "a1"); err != nil {
		t.Fatalf("CreateRatingState: %v", err)
	}
	beforeID := stateRowID(t, sqlDB, "u1", "a1")
	beforeCreated := stateRowCreatedAt(t, sqlDB, "u1", "a1")

	// Re-rate via the plain save path: write a new log entry carrying the
	// current state and do nothing to the state row.
	if _, err := svc.AddRating(ctx, "u1", "a1", 6.5, "", RatingStateProvisional); err != nil {
		t.Fatalf("re-rate AddRating: %v", err)
	}

	got, err := svc.GetRatingState(ctx, "u1", "a1")
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if got == nil || got.State != RatingStateProvisional {
		t.Fatalf("re-rate must leave state=provisional, got %+v", got)
	}
	if afterID := stateRowID(t, sqlDB, "u1", "a1"); afterID != beforeID {
		t.Fatalf("state row identity changed: %q -> %q", beforeID, afterID)
	}
	if afterCreated := stateRowCreatedAt(t, sqlDB, "u1", "a1"); afterCreated != beforeCreated {
		t.Fatalf("state row created_at changed: %q -> %q", beforeCreated, afterCreated)
	}
}

func TestRatingLifecycle_RerateLeavesFinalizedUntouched(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "u1", "a1")

	if _, err := svc.AddRating(ctx, "u1", "a1", 6.0, "", RatingStateProvisional); err != nil {
		t.Fatalf("initial AddRating: %v", err)
	}
	if _, err := svc.CreateRatingState(ctx, "u1", "a1"); err != nil {
		t.Fatalf("CreateRatingState: %v", err)
	}
	if _, _, err := svc.FinalizeWithRating(ctx, "u1", "a1", 8.0, ""); err != nil {
		t.Fatalf("FinalizeWithRating: %v", err)
	}
	beforeID := stateRowID(t, sqlDB, "u1", "a1")

	// Re-rate the now-finalized album via the plain save path. The log carries
	// the current (finalized) state on the new entry; the state row stays put.
	if _, err := svc.AddRating(ctx, "u1", "a1", 9.0, "", RatingStateFinalized); err != nil {
		t.Fatalf("re-rate AddRating: %v", err)
	}

	got, err := svc.GetRatingState(ctx, "u1", "a1")
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if got == nil || got.State != RatingStateFinalized {
		t.Fatalf("re-rate must leave state=finalized, got %+v", got)
	}
	if afterID := stateRowID(t, sqlDB, "u1", "a1"); afterID != beforeID {
		t.Fatalf("state row identity changed: %q -> %q", beforeID, afterID)
	}
}

func TestRatingLifecycle_PromotionFromProvisionalIsOnlyPathToFinalized(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "u1", "a1")

	// Unrated -> Finalize: rejected, no rows written.
	if _, _, err := svc.FinalizeWithRating(ctx, "u1", "a1", 7.0, ""); !errors.Is(err, ErrFinalizeRequiresProvisional) {
		t.Fatalf("unrated finalize: expected ErrFinalizeRequiresProvisional, got %v", err)
	}
	if cnt := stateRowCount(t, sqlDB, "u1", "a1"); cnt != 0 {
		t.Fatalf("unrated finalize must not create a state row, got count=%d", cnt)
	}

	// Establish provisional, then promote.
	if _, err := svc.AddRating(ctx, "u1", "a1", 7.0, "", RatingStateProvisional); err != nil {
		t.Fatalf("AddRating: %v", err)
	}
	if _, err := svc.CreateRatingState(ctx, "u1", "a1"); err != nil {
		t.Fatalf("CreateRatingState: %v", err)
	}
	if _, _, err := svc.FinalizeWithRating(ctx, "u1", "a1", 8.0, ""); err != nil {
		t.Fatalf("FinalizeWithRating: %v", err)
	}
	got, err := svc.GetRatingState(ctx, "u1", "a1")
	if err != nil || got == nil || got.State != RatingStateFinalized {
		t.Fatalf("expected finalized after promotion, got state=%+v err=%v", got, err)
	}

	// Already-finalized -> Finalize: rejected.
	if _, _, err := svc.FinalizeWithRating(ctx, "u1", "a1", 9.0, ""); !errors.Is(err, ErrFinalizeRequiresProvisional) {
		t.Fatalf("already-finalized: expected ErrFinalizeRequiresProvisional, got %v", err)
	}
}

func TestRatingLifecycle_LiveStateCheckRejectsStalled(t *testing.T) {
	_, sqlDB := newTestService(t)
	seedUserAndAlbum(t, sqlDB, "u1", "a1")

	// The live album_rating_state.state CHECK is narrowed to {provisional,
	// finalized} after the retirement migration. Direct insert of 'stalled' is
	// rejected by sqlite — proof that even a buggy code path couldn't smuggle
	// the historical value back onto the live state row.
	_, err := sqlDB.Exec(
		`INSERT INTO album_rating_state (id, user_id, album_id, state) VALUES ('s1', 'u1', 'a1', 'stalled')`,
	)
	if err == nil {
		t.Fatal("expected CHECK constraint to reject state='stalled' on album_rating_state")
	}
}

// --- Re-rate save path leaves the state value untouched (PC3) ---
//
// Exercises the precise call sequence used by the http handler's
// SubmitRatingRecommenderRating path: GetRatingState, AddRating (carrying the
// current state on the log row), then CreateRatingState only when no state row
// existed. The state value must end the call equal to what it started at.

func TestSaveFormPath_LeavesProvisionalStateValueUnchanged(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "u1", "a1")

	// Set up: provisional album with one prior log entry.
	if _, err := svc.AddRating(ctx, "u1", "a1", 5.0, "", RatingStateProvisional); err != nil {
		t.Fatalf("setup AddRating: %v", err)
	}
	if _, err := svc.CreateRatingState(ctx, "u1", "a1"); err != nil {
		t.Fatalf("setup CreateRatingState: %v", err)
	}

	// Re-execute the handler's save path verbatim.
	runHandlerSaveFlow(t, svc, ctx, "u1", "a1", 6.7)

	got, err := svc.GetRatingState(ctx, "u1", "a1")
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if got == nil || got.State != RatingStateProvisional {
		t.Fatalf("save on provisional must leave state=provisional, got %+v", got)
	}
}

func TestSaveFormPath_LeavesFinalizedStateValueUnchanged(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "u1", "a1")

	// Set up: finalized album.
	if _, err := svc.AddRating(ctx, "u1", "a1", 7.0, "", RatingStateProvisional); err != nil {
		t.Fatalf("setup AddRating: %v", err)
	}
	if _, err := svc.CreateRatingState(ctx, "u1", "a1"); err != nil {
		t.Fatalf("setup CreateRatingState: %v", err)
	}
	if _, _, err := svc.FinalizeWithRating(ctx, "u1", "a1", 8.0, ""); err != nil {
		t.Fatalf("setup FinalizeWithRating: %v", err)
	}

	runHandlerSaveFlow(t, svc, ctx, "u1", "a1", 7.5)

	got, err := svc.GetRatingState(ctx, "u1", "a1")
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if got == nil || got.State != RatingStateFinalized {
		t.Fatalf("save on finalized must leave state=finalized, got %+v", got)
	}
}

// runHandlerSaveFlow replays the body of SubmitRatingRecommenderRating against
// the service. Kept in the test file so a future handler-shape change shows up
// here as a diff, not a silent drift.
func runHandlerSaveFlow(t *testing.T, svc *Service, ctx context.Context, userID, albumID string, rating float64) {
	t.Helper()
	current, err := svc.GetRatingState(ctx, userID, albumID)
	if err != nil {
		t.Fatalf("GetRatingState: %v", err)
	}
	logState := RatingStateProvisional
	if current != nil {
		logState = current.State
	}
	if _, err := svc.AddRating(ctx, userID, albumID, rating, "", logState); err != nil {
		t.Fatalf("AddRating: %v", err)
	}
	if current == nil {
		if _, err := svc.CreateRatingState(ctx, userID, albumID); err != nil {
			t.Fatalf("CreateRatingState: %v", err)
		}
	}
}

// --- Pre-fill tie-break (PC5 data-layer half) ---
//
// The pre-fill is sourced from GetLatestRating, which the query orders by
// created_at DESC, id DESC. When two log entries share a created_at the row
// with the greater id must win. This guards the SQL ORDER BY against
// regression — the UI half lives in the Playwright suite.

func TestLatestRating_TieBreakOnIDDesc(t *testing.T) {
	svc, sqlDB := newTestService(t)
	ctx := context.Background()
	seedUserAndAlbum(t, sqlDB, "u1", "a1")

	// Two log entries with identical created_at, different ids. The greater id
	// wins the latest-rating selection.
	const sharedTime = "2026-04-01 12:00:00"
	mustExecf(t, sqlDB,
		`INSERT INTO album_rating_log (id, user_id, album_id, rating, created_at) VALUES ('aaaa', 'u1', 'a1', 5.0, ?)`,
		sharedTime,
	)
	mustExecf(t, sqlDB,
		`INSERT INTO album_rating_log (id, user_id, album_id, rating, created_at) VALUES ('zzzz', 'u1', 'a1', 9.0, ?)`,
		sharedTime,
	)

	got, err := svc.GetLatestRating(ctx, "u1", "a1")
	if err != nil {
		t.Fatalf("GetLatestRating: %v", err)
	}
	if got == nil || got.Rating == nil || *got.Rating != 9.0 {
		t.Fatalf("expected the greater-id row (rating=9.0) to win the tie, got %+v", got)
	}
}

// --- helpers ---

func stateRowID(t *testing.T, sqlDB *sql.DB, userID, albumID string) string {
	t.Helper()
	var id string
	err := sqlDB.QueryRow(
		`SELECT id FROM album_rating_state WHERE user_id = ? AND album_id = ?`,
		userID, albumID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("read state row id: %v", err)
	}
	return id
}

func stateRowCreatedAt(t *testing.T, sqlDB *sql.DB, userID, albumID string) string {
	t.Helper()
	var ts string
	err := sqlDB.QueryRow(
		`SELECT created_at FROM album_rating_state WHERE user_id = ? AND album_id = ?`,
		userID, albumID,
	).Scan(&ts)
	if err != nil {
		t.Fatalf("read state row created_at: %v", err)
	}
	return ts
}

func stateRowCount(t *testing.T, sqlDB *sql.DB, userID, albumID string) int {
	t.Helper()
	var n int
	err := sqlDB.QueryRow(
		`SELECT COUNT(*) FROM album_rating_state WHERE user_id = ? AND album_id = ?`,
		userID, albumID,
	).Scan(&n)
	if err != nil {
		t.Fatalf("count state rows: %v", err)
	}
	return n
}

func mustExecf(t *testing.T, sqlDB *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := sqlDB.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
