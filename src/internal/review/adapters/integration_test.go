package adapters

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/listeninghistory"
	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/tags"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

// --- assembled-system test rig ---
//
// Spins up the real review HTTP handler wired to the real review + library
// services backed by a fresh sqlite DB with every migration applied. The rig
// is deliberately end-to-end at the handler boundary: feature-flag-free, real
// services, real router, no mocking. The only stubbed dependency is the
// listening-history service's spotify client (not exercised by the rating
// endpoints under test) — passed nil and the test seeds zero play history.

type harness struct {
	mux    *httpx.Mux
	svc    *review.Service
	libSvc *library.Service
	sqlDB  *sql.DB
	userID string
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	migrationsDir, err := filepath.Abs("../../../../db/migrations")
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

	wrapped := db.WrapSqlDB(sqlDB)
	reviewSvc := review.NewService(wrapped)
	tagsSvc := tags.NewService(wrapped)
	notesSvc := notes.NewService(wrapped)
	// listening-history's GetLastPlayedAtByAlbumIds method (the only one used
	// downstream of the rating endpoints) hits the local DB only — the spotify
	// service is required by the constructor but unused on this path.
	histSvc := listeninghistory.NewService(wrapped, nil)
	libSvc := library.NewService(wrapped, nil, histSvc, tagsSvc, notesSvc, reviewSvc)

	mux := httpx.NewMux(app.NewApp(app.Config{}))
	RegisterRoutes(mux, NewHttpHandler(libSvc, reviewSvc))

	h := &harness{
		mux:    mux,
		svc:    reviewSvc,
		libSvc: libSvc,
		sqlDB:  sqlDB,
		userID: uuid.NewString(),
	}
	h.seedUser()
	return h
}

// seedUser inserts the harness's user row so the FK on every downstream table
// is satisfied.
func (h *harness) seedUser() {
	if _, err := h.sqlDB.Exec(
		`INSERT INTO users (id, spotify_id) VALUES (?, ?)`,
		h.userID, "sp-"+h.userID,
	); err != nil {
		panic(err)
	}
}

// seedAlbumInLibrary creates an album, a release, and an owned user_release for
// the harness user — the minimum needed for GetAlbumInLibrary to return.
func (h *harness) seedAlbumInLibrary(t *testing.T, title string) string {
	t.Helper()
	albumID := uuid.NewString()
	if _, err := h.sqlDB.Exec(
		`INSERT INTO albums (id, spotify_id, title) VALUES (?, ?, ?)`,
		albumID, "sp-"+albumID, title,
	); err != nil {
		t.Fatalf("seed album: %v", err)
	}
	releaseID := uuid.NewString()
	if _, err := h.sqlDB.Exec(
		`INSERT INTO releases (id, album_id, format) VALUES (?, ?, 'digital')`,
		releaseID, albumID,
	); err != nil {
		t.Fatalf("seed release: %v", err)
	}
	if _, err := h.sqlDB.Exec(
		`INSERT INTO user_releases (id, user_id, release_id, status, created_at, status_updated_at) VALUES (?, ?, ?, 'owned', current_timestamp, current_timestamp)`,
		uuid.NewString(), h.userID, releaseID,
	); err != nil {
		t.Fatalf("seed user_release: %v", err)
	}
	return albumID
}

// do executes a request through the registered mux with the harness's user ID
// installed on the ContextX (mirroring the JWT middleware's behaviour).
func (h *harness) do(method, target string, body url.Values) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, strings.NewReader(body.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	ctx := contextx.NewContextX(req.Context()).WithUserId(h.userID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	return rec
}

// stateRow returns the current state value for the user/album, or "" when no
// row exists.
func (h *harness) stateRow(t *testing.T, albumID string) string {
	t.Helper()
	var st string
	err := h.sqlDB.QueryRow(
		`SELECT state FROM album_rating_state WHERE user_id = ? AND album_id = ?`,
		h.userID, albumID,
	).Scan(&st)
	if err == sql.ErrNoRows {
		return ""
	}
	if err != nil {
		t.Fatalf("read state row: %v", err)
	}
	return st
}

func (h *harness) logRowCount(t *testing.T, albumID string) int {
	t.Helper()
	var n int
	if err := h.sqlDB.QueryRow(
		`SELECT COUNT(*) FROM album_rating_log WHERE user_id = ? AND album_id = ?`,
		h.userID, albumID,
	).Scan(&n); err != nil {
		t.Fatalf("count log rows: %v", err)
	}
	return n
}

func (h *harness) stateRowCount(t *testing.T, albumID string) int {
	t.Helper()
	var n int
	if err := h.sqlDB.QueryRow(
		`SELECT COUNT(*) FROM album_rating_state WHERE user_id = ? AND album_id = ?`,
		h.userID, albumID,
	).Scan(&n); err != nil {
		t.Fatalf("count state rows: %v", err)
	}
	return n
}

// --- Single modal entry point lands on the score-entry form ---
//
// GET /app/review/rating-recommender must always render the score-entry form
// fragment, regardless of the album's rating state. The body must carry the
// rating-confirm-form test id and must not carry any of the retired
// alternate-entry-point ids (base-questions-form, rerate-prompt) as a
// first-view marker.

func TestModalEntry_AlwaysReturnsScoreEntryForm_Unrated(t *testing.T) {
	h := newHarness(t)
	albumID := h.seedAlbumInLibrary(t, "Unrated Album")

	rec := h.do(http.MethodGet, "/app/review/rating-recommender?albumId="+albumID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("unrated modal: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	assertModalLandsOnScoreEntry(t, rec.Body.String())
}

func TestModalEntry_AlwaysReturnsScoreEntryForm_Provisional(t *testing.T) {
	h := newHarness(t)
	albumID := h.seedAlbumInLibrary(t, "Provisional Album")
	if _, err := h.svc.AddRating(t.Context(), h.userID, albumID, 6.0, "", review.RatingStateProvisional); err != nil {
		t.Fatalf("seed rating: %v", err)
	}
	if _, err := h.svc.CreateRatingState(t.Context(), h.userID, albumID); err != nil {
		t.Fatalf("seed state: %v", err)
	}

	rec := h.do(http.MethodGet, "/app/review/rating-recommender?albumId="+albumID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("provisional modal: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	assertModalLandsOnScoreEntry(t, rec.Body.String())
}

func TestModalEntry_AlwaysReturnsScoreEntryForm_Finalized(t *testing.T) {
	h := newHarness(t)
	albumID := h.seedAlbumInLibrary(t, "Finalized Album")
	if _, err := h.svc.AddRating(t.Context(), h.userID, albumID, 6.0, "", review.RatingStateProvisional); err != nil {
		t.Fatalf("seed rating: %v", err)
	}
	if _, err := h.svc.CreateRatingState(t.Context(), h.userID, albumID); err != nil {
		t.Fatalf("seed state: %v", err)
	}
	if _, _, err := h.svc.FinalizeWithRating(t.Context(), h.userID, albumID, 8.0, ""); err != nil {
		t.Fatalf("seed finalize: %v", err)
	}

	rec := h.do(http.MethodGet, "/app/review/rating-recommender?albumId="+albumID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("finalized modal: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	assertModalLandsOnScoreEntry(t, rec.Body.String())
}

func assertModalLandsOnScoreEntry(t *testing.T, body string) {
	t.Helper()
	if !strings.Contains(body, `data-testid="rating-confirm-form"`) {
		t.Fatalf("modal entry-point body missing rating-confirm-form testid; body: %s", body)
	}
	if strings.Contains(body, `data-testid="base-questions-form"`) {
		t.Fatalf("modal entry-point body must not lead with the questionnaire; body: %s", body)
	}
	if strings.Contains(body, `data-testid="rerate-prompt"`) {
		t.Fatalf("modal entry-point body must not contain the retired rerate prompt; body: %s", body)
	}
}

// --- Re-rate save sets the state to provisional (HTTP boundary) ---

func TestSaveForm_OnProvisionalAlbum_LeavesStateProvisional(t *testing.T) {
	h := newHarness(t)
	albumID := h.seedAlbumInLibrary(t, "Provisional Save")
	if _, err := h.svc.AddRating(t.Context(), h.userID, albumID, 5.0, "", review.RatingStateProvisional); err != nil {
		t.Fatalf("seed rating: %v", err)
	}
	if _, err := h.svc.CreateRatingState(t.Context(), h.userID, albumID); err != nil {
		t.Fatalf("seed state: %v", err)
	}

	form := url.Values{}
	form.Set("rating", "7.2")
	form.Set("mode", string(review.RatingModeProvisional))
	rec := h.do(http.MethodPost, "/app/review/rating-recommender/rating?albumId="+albumID, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save POST: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if got := h.stateRow(t, albumID); got != string(review.RatingStateProvisional) {
		t.Fatalf("save on provisional must leave state=provisional, got %q", got)
	}
}

func TestSaveForm_OnFinalizedAlbum_DemotesToProvisional(t *testing.T) {
	h := newHarness(t)
	albumID := h.seedAlbumInLibrary(t, "Finalized Save")
	if _, _, err := h.svc.FinalizeWithRating(t.Context(), h.userID, albumID, 8.0, ""); err != nil {
		t.Fatalf("seed finalize: %v", err)
	}

	form := url.Values{}
	form.Set("rating", "8.5")
	form.Set("mode", string(review.RatingModeProvisional))
	rec := h.do(http.MethodPost, "/app/review/rating-recommender/rating?albumId="+albumID, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("save POST: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if got := h.stateRow(t, albumID); got != string(review.RatingStateProvisional) {
		t.Fatalf("save on finalized must demote to provisional, got %q", got)
	}
}

// --- Questionnaire neutrality (PC7) ---
//
// POST /app/review/rating-recommender/questions writes no rating-log row and
// no rating-state row on its own — its only output is a score-entry form
// fragment with the computed score pre-filled. Persistence happens only on a
// subsequent save / finalize from the score-entry form.

func TestQuestionnaireSubmit_WritesNoRatingRows_AndReturnsScoreEntryWithPrefill(t *testing.T) {
	h := newHarness(t)
	albumID := h.seedAlbumInLibrary(t, "Questionnaire Album")

	beforeLog := h.logRowCount(t, albumID)
	beforeState := h.stateRowCount(t, albumID)

	form := url.Values{}
	// All-1s answers — valid for every base question, deterministic computed score.
	for _, q := range review.AllBaseQuestions {
		form.Set(string(q.Key), "1")
	}
	rec := h.do(http.MethodPost,
		"/app/review/rating-recommender/questions?albumId="+albumID+"&mode=provisional",
		form,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("questionnaire POST: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	if got := h.logRowCount(t, albumID); got != beforeLog {
		t.Fatalf("questionnaire submit must not write album_rating_log rows: before=%d after=%d", beforeLog, got)
	}
	if got := h.stateRowCount(t, albumID); got != beforeState {
		t.Fatalf("questionnaire submit must not write album_rating_state rows: before=%d after=%d", beforeState, got)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `data-testid="rating-confirm-form"`) {
		t.Fatalf("questionnaire submit must return the score-entry form; body: %s", body)
	}
	// All-1s answers yield 0.0, but the pre-fill formats it as "0.0" inside the
	// value attribute. Asserting on the input's value attribute pins the
	// pre-fill contract.
	if !strings.Contains(body, `data-testid="rating-confirm-form-input"`) {
		t.Fatalf("response missing score-entry input; body: %s", body)
	}
	if !strings.Contains(body, `value="0.0"`) {
		t.Fatalf("expected score-entry input pre-filled with the computed score (0.0); body: %s", body)
	}
}
