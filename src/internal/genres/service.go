package genres

import (
	"context"
	"fmt"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/genregraph"

	"github.com/google/uuid"
)

func newID() string { return uuid.NewString() }

// AlbumPrimaries returns each album's primary genres, keyed by album ID. An
// album with no resolved genres (or none mapping to a primary) is absent from
// the map — callers treat absence as uncategorized. Primaries are unioned
// across the album's leaf genres and returned in the graph's curated order.
func (s *Service) AlbumPrimaries(ctx context.Context, albumIDs []string) (map[string][]genregraph.Primary, error) {
	out := make(map[string][]genregraph.Primary)
	if len(albumIDs) == 0 || s.graph == nil {
		return out, nil
	}

	byAlbum, err := s.repo.GetAlbumGenresByAlbumIDs(ctx, albumIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load album genres: %w", err)
	}

	for albumID, leaves := range byAlbum {
		hit := make(map[string]bool)
		for _, leaf := range leaves {
			for _, p := range s.graph.PrimariesOf(leaf.GenreID) {
				hit[p.ID] = true
			}
		}
		if len(hit) == 0 {
			continue
		}
		var prims []genregraph.Primary
		for _, p := range s.graph.Primaries() { // curated order
			if hit[p.ID] {
				prims = append(prims, p)
			}
		}
		out[albumID] = prims
	}
	return out, nil
}

// EnrichAlbum resolves an album's genres from Discogs, replaces its stored leaf
// genres, and marks it enriched — atomically. Marking happens even when nothing
// resolved, so the album is not re-queried.
func (s *Service) EnrichAlbum(ctx contextx.ContextX, album AlbumForEnrichment) error {
	if s.discogs == nil {
		return nil
	}
	nodes := s.discogs.ResolveAlbumGenreNodes(ctx, album.Title, album.Artist)

	genres := make([]AlbumGenreDTO, 0, len(nodes))
	for _, n := range nodes {
		genres = append(genres, AlbumGenreDTO{GenreID: n.ID, Label: n.Label})
	}

	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		if err := txRepo.ReplaceAlbumGenres(ctx, album.ID, genres); err != nil {
			return fmt.Errorf("failed to store album genres: %w", err)
		}
		if err := txRepo.MarkEnriched(ctx, album.ID); err != nil {
			return fmt.Errorf("failed to mark album enriched: %w", err)
		}
		return nil
	})
}

// EnrichPending enriches up to limit albums (from the source) that have not yet
// been enriched, querying Discogs for each. The Discogs client self-throttles,
// and the per-run limit keeps each run bounded. Returns the number enriched.
func (s *Service) EnrichPending(ctx contextx.ContextX, source AlbumGenreSource, limit int) (int, error) {
	albums, err := source.AlbumsForGenreEnrichment(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list albums for enrichment: %w", err)
	}

	ids := make([]string, len(albums))
	for i, a := range albums {
		ids[i] = a.ID
	}
	enriched, err := s.repo.EnrichedAlbumIDs(ctx, ids)
	if err != nil {
		return 0, fmt.Errorf("failed to load enrichment state: %w", err)
	}

	count := 0
	for _, album := range albums {
		if enriched[album.ID] {
			continue
		}
		if limit > 0 && count >= limit {
			break
		}
		if err := s.EnrichAlbum(ctx, album); err != nil {
			return count, fmt.Errorf("failed to enrich album %s: %w", album.ID, err)
		}
		count++
	}
	return count, nil
}
