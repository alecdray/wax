package views

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/review"
)

// Historical rating-log rows can carry lifecycle values that are no longer
// produced by the live state machine. The history fragment must render those
// values with a recognisable label rather than panicking or emitting an empty
// or raw-enum cell.
func TestAlbumRatingHistoryFrag_RendersHistoricalStalledLabel(t *testing.T) {
	rating := 7.5
	stalled := review.RatingState("stalled")
	album := library.AlbumDTO{
		ID:    "album-1",
		Title: "Test Album",
		RatingLog: []*review.AlbumRatingDTO{
			{
				ID:        "log-1",
				UserID:    "user-1",
				AlbumID:   "album-1",
				Rating:    &rating,
				State:     &stalled,
				CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	var buf bytes.Buffer
	if err := AlbumRatingHistoryFrag(album, false).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Stalled") {
		t.Fatalf("expected output to contain a recognisable label for stalled history rows; got: %s", out)
	}
	// The raw enum value alone must not be the only thing rendered for the
	// historical state — the label is its human-readable form.
	if strings.Contains(out, ">stalled<") {
		t.Fatalf("expected the human-readable label, not the raw enum value; got: %s", out)
	}
}
