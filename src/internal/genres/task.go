package genres

import (
	"log/slog"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/task"
)

// enrichBatchLimit bounds how many albums one run resolves against Discogs. The
// Discogs client self-throttles (~60 req/min), so a bounded batch keeps each run
// short and lets successive runs work through the backlog.
const enrichBatchLimit = 50

// EnrichGenresTask backfills album genres from Discogs for albums not yet
// enriched, and keeps newly-synced albums covered.
type EnrichGenresTask struct {
	service *Service
	source  AlbumGenreSource
}

var _ task.Task = EnrichGenresTask{}

// NewEnrichGenresTask wires the task with the genres service and the album
// catalog source (the library service satisfies AlbumGenreSource).
func NewEnrichGenresTask(service *Service, source AlbumGenreSource) task.Task {
	return EnrichGenresTask{service: service, source: source}
}

func (t EnrichGenresTask) Run(ctx contextx.ContextX) error {
	n, err := t.service.EnrichPending(ctx, t.source, enrichBatchLimit)
	if err != nil {
		return err
	}
	if n > 0 {
		slog.Info("enriched album genres from Discogs", "count", n)
	}
	return nil
}

func (t EnrichGenresTask) Schedule() *task.CronExpression {
	schedule := task.CronExpression("*/15 * * * *") // every 15 minutes
	return &schedule
}

func (t EnrichGenresTask) Name() string {
	return "enrich_album_genres"
}
