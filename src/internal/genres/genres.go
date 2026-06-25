// Package genres owns an album's structured genre facet: the resolved leaf
// genre nodes (Discogs-derived Wikidata Q-ids) persisted per album, and the
// derivation of an album's primary genres from them via the genre graph.
// See ADR 0009.
package genres

import (
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/genregraph"
)

// AlbumGenreDTO is a resolved leaf genre on an album — a node in the genre graph.
type AlbumGenreDTO struct {
	GenreID string
	Label   string
}

// AlbumForEnrichment is the album metadata the enrichment flow needs to query
// Discogs. It is supplied by the module that owns album metadata (see
// AlbumGenreSource), keeping this module free of that dependency.
type AlbumForEnrichment struct {
	ID     string
	Title  string
	Artist string
}

// AlbumGenreSource supplies the album catalog for genre enrichment. The owner
// of album metadata satisfies it; this module defines it so it never imports
// that owner (avoiding an import cycle).
type AlbumGenreSource interface {
	AlbumsForGenreEnrichment(ctx contextx.ContextX) ([]AlbumForEnrichment, error)
}

// Service owns the album genre facet: reads (per-album primaries) and writes
// (enrichment). Genres are app-curated and album-intrinsic, so storage is
// global per album rather than per user.
type Service struct {
	db      *db.DB
	repo    *Repo
	discogs *discogs.Service
	graph   *genregraph.DAG
}

// NewService builds the genres service. graph may be nil if the genre graph
// failed to load; primary derivation then yields nothing rather than panicking.
func NewService(d *db.DB, discogsSvc *discogs.Service, graph *genregraph.DAG) *Service {
	return &Service{
		db:      d,
		repo:    NewRepo(d.Queries()),
		discogs: discogsSvc,
		graph:   graph,
	}
}
