package adapters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/httpx"
)

// The legacy snooze endpoint backing the time-based rerate lifecycle is no
// longer registered. A request to it must hit the framework's not-found
// response.
func TestSnoozeRoute_NotRegistered(t *testing.T) {
	mux := httpx.NewMux(app.NewApp(app.Config{}))
	RegisterRoutes(mux, &HttpHandler{})

	req := httptest.NewRequest(http.MethodPost, "/app/review/rating/snooze?albumId=x", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for retired snooze route, got %d", rec.Code)
	}
}
