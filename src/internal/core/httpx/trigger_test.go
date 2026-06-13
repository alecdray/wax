package httpx

import (
	"net/http/httptest"
	"testing"
)

func TestSetHXTrigger(t *testing.T) {
	t.Run("sets HX-Trigger header with event name and detail payload", func(t *testing.T) {
		rec := httptest.NewRecorder()

		if err := SetHXTrigger(rec, "album-changed", map[string]string{"albumId": "abc"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := rec.Header().Get("HX-Trigger")
		want := `{"album-changed":{"albumId":"abc"}}`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
