package adapters

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
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
	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

// HTTP-level coverage for the provisional carousel endpoint. Mirrors the
// review adapter's harness pattern: fresh sqlite DB + every migration + real
// library service wired through the real mux. The repo-level tests in
// library/repo_provisional_test.go cover the DTO shape at the repo→service
// seam; these tests cover the rendered HTTP response — both the status code
// for every DB-permitted (state row, log rows) combination, and the per-card
// rating marker carried in the rendered HTML.

type carouselHarness struct {
	mux    *httpx.Mux
	sqlDB  *sql.DB
	userID string
}

func newCarouselHarness(t *testing.T) *carouselHarness {
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
	// listening-history's spotify client is not exercised by the carousel
	// endpoint — nil is safe here.
	histSvc := listeninghistory.NewService(wrapped, nil)
	// spotify service is also unused on this path; library's constructor
	// accepts a nil pointer for it.
	libSvc := library.NewService(wrapped, nil, histSvc, tagsSvc, notesSvc, reviewSvc)

	// The carousel handler only consumes libraryService — every other
	// collaborator is nil-safe along this request path.
	handler := NewHttpHandler(nil, nil, nil, libSvc, nil, nil, nil)
	mux := httpx.NewMux(app.NewApp(app.Config{}))
	RegisterRoutes(mux, handler)

	h := &carouselHarness{
		mux:    mux,
		sqlDB:  sqlDB,
		userID: "u-carousel-test",
	}
	h.seedUser()
	return h
}

func (h *carouselHarness) seedUser() {
	if _, err := h.sqlDB.Exec(
		`INSERT INTO users (id, spotify_id) VALUES (?, ?)`,
		h.userID, "sp-"+h.userID,
	); err != nil {
		panic(err)
	}
}

func (h *carouselHarness) seedAlbum(t *testing.T, albumID, title string) {
	t.Helper()
	if _, err := h.sqlDB.Exec(
		`INSERT INTO albums (id, spotify_id, title) VALUES (?, ?, ?)`,
		albumID, "sp-"+albumID, title,
	); err != nil {
		t.Fatalf("seed album: %v", err)
	}
}

func (h *carouselHarness) seedProvisionalState(t *testing.T, albumID string) {
	t.Helper()
	if _, err := h.sqlDB.Exec(
		`INSERT INTO album_rating_state (id, user_id, album_id, state)
		 VALUES (?, ?, ?, 'provisional')`,
		"ars-"+h.userID+"-"+albumID, h.userID, albumID,
	); err != nil {
		t.Fatalf("seed provisional state: %v", err)
	}
}

func (h *carouselHarness) seedRatingLog(t *testing.T, id, albumID string, rating float64, createdAt string) {
	t.Helper()
	if _, err := h.sqlDB.Exec(
		`INSERT INTO album_rating_log (id, user_id, album_id, rating, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		id, h.userID, albumID, rating, createdAt,
	); err != nil {
		t.Fatalf("seed rating log: %v", err)
	}
}

func (h *carouselHarness) getProvisionalCarousel(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/app/library/dashboard/carousel?view=provisional", nil)
	ctx := contextx.NewContextX(req.Context()).WithUserId(h.userID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	return rec
}

// --- Status-code coverage for every DB-permitted state/log shape ---

func TestProvisionalCarousel_ReturnsOK_ForEmptyLibrary(t *testing.T) {
	h := newCarouselHarness(t)

	rec := h.getProvisionalCarousel(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("empty library: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `data-testid="provisional-carousel-strip-empty"`) {
		t.Fatalf("empty library: expected empty-state marker in body; got: %s", rec.Body.String())
	}
}

func TestProvisionalCarousel_ReturnsOK_WhenProvisionalAlbumHasNoLogRow(t *testing.T) {
	h := newCarouselHarness(t)

	h.seedAlbum(t, "a1", "Backfilled From Stalled")
	h.seedProvisionalState(t, "a1")
	// No album_rating_log row — this is the shape that used to 500.

	rec := h.getProvisionalCarousel(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("log-less provisional album: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestProvisionalCarousel_ReturnsOK_WhenProvisionalAlbumHasLogRows(t *testing.T) {
	h := newCarouselHarness(t)

	h.seedAlbum(t, "a1", "Rated Provisional")
	h.seedProvisionalState(t, "a1")
	h.seedRatingLog(t, "log-1", "a1", 6.0, "2026-04-01 12:00:00")
	h.seedRatingLog(t, "log-2", "a1", 7.5, "2026-05-01 12:00:00")

	rec := h.getProvisionalCarousel(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("rated provisional album: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestProvisionalCarousel_ReturnsOK_ForMixedLibrary(t *testing.T) {
	h := newCarouselHarness(t)

	h.seedAlbum(t, "a-empty", "Log-less Provisional")
	h.seedProvisionalState(t, "a-empty")

	h.seedAlbum(t, "a-rated", "Rated Provisional")
	h.seedProvisionalState(t, "a-rated")
	h.seedRatingLog(t, "log-1", "a-rated", 8.25, "2026-05-10 09:00:00")

	rec := h.getProvisionalCarousel(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("mixed library: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
}

// --- Per-card rating fidelity in the rendered HTML ---
//
// The provisional card carries a data-album-id attribute on every card and a
// data-rating attribute only when the underlying ProvisionalAlbumDTO.Rating
// is non-nil. That gives a stable, response-observable proxy for the DTO
// contract: rating absent iff the album has no album_rating_log row,
// otherwise equal to the most recent log row's rating.

func TestProvisionalCarousel_CardRatingMarker_ReflectsUnderlyingData(t *testing.T) {
	h := newCarouselHarness(t)

	h.seedAlbum(t, "a-empty", "Log-less Provisional")
	h.seedProvisionalState(t, "a-empty")

	h.seedAlbum(t, "a-rated", "Rated Provisional")
	h.seedProvisionalState(t, "a-rated")
	// Older log row exists — handler should surface only the latest.
	h.seedRatingLog(t, "log-old", "a-rated", 5.0, "2026-04-01 12:00:00")
	h.seedRatingLog(t, "log-new", "a-rated", 7.5, "2026-05-01 12:00:00")

	h.seedAlbum(t, "a-integral", "Integral Rating")
	h.seedProvisionalState(t, "a-integral")
	h.seedRatingLog(t, "log-int", "a-integral", 8.0, "2026-05-05 12:00:00")

	rec := h.getProvisionalCarousel(t)
	if rec.Code != http.StatusOK {
		t.Fatalf("mixed library: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	cards := parseProvisionalCards(t, body)

	if len(cards) != 3 {
		t.Fatalf("expected 3 provisional cards in rendered HTML, got %d; body: %s", len(cards), body)
	}

	wantRating := map[string]*float64{
		"a-empty":    nil,
		"a-rated":    ptrFloat(7.5),
		"a-integral": ptrFloat(8.0),
	}
	for albumID, want := range wantRating {
		got, ok := cards[albumID]
		if !ok {
			t.Fatalf("card for album %q missing from rendered response; body: %s", albumID, body)
		}
		switch {
		case want == nil && got.hasRating:
			t.Fatalf("album %q: expected no data-rating attribute, got %q", albumID, got.rating)
		case want != nil && !got.hasRating:
			t.Fatalf("album %q: expected data-rating=%v, attribute absent", albumID, *want)
		case want != nil && got.hasRating:
			gotF := parseRatingAttr(t, albumID, got.rating)
			if gotF != *want {
				t.Fatalf("album %q: expected data-rating=%v, got %v (raw=%q)", albumID, *want, gotF, got.rating)
			}
		}
	}
}

// --- helpers ---

type provisionalCardAttrs struct {
	hasRating bool
	rating    string // raw attribute value, only valid when hasRating is true
}

// parseProvisionalCards extracts the per-card attributes from the rendered
// carousel HTML, keyed by data-album-id. The rendered card is a single <a>
// element with data-testid="provisional-carousel-strip-album-card", always
// carrying data-album-id and optionally data-rating.
func parseProvisionalCards(t *testing.T, body string) map[string]provisionalCardAttrs {
	t.Helper()

	// Anchor on the card's testid, then capture the full opening tag so we
	// can pick out data-album-id and (optionally) data-rating regardless of
	// attribute order.
	cardRe := regexp.MustCompile(`<a [^>]*data-testid="provisional-carousel-strip-album-card"[^>]*>`)
	idRe := regexp.MustCompile(`data-album-id="([^"]*)"`)
	ratingRe := regexp.MustCompile(`data-rating="([^"]*)"`)

	out := make(map[string]provisionalCardAttrs)
	for _, openTag := range cardRe.FindAllString(body, -1) {
		idMatch := idRe.FindStringSubmatch(openTag)
		if len(idMatch) != 2 {
			t.Fatalf("provisional card missing data-album-id: %s", openTag)
		}
		attrs := provisionalCardAttrs{}
		if r := ratingRe.FindStringSubmatch(openTag); len(r) == 2 {
			attrs.hasRating = true
			attrs.rating = r[1]
		}
		out[idMatch[1]] = attrs
	}
	return out
}

func parseRatingAttr(t *testing.T, albumID, raw string) float64 {
	t.Helper()
	var f float64
	if _, err := fmt.Sscanf(raw, "%g", &f); err != nil {
		t.Fatalf("album %q: data-rating=%q is not a float: %v", albumID, raw, err)
	}
	return f
}

func ptrFloat(v float64) *float64 { return &v }
