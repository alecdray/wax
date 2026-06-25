package discogs

import (
	"log/slog"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/genregraph"
)

type Service struct {
	client *Client
	dag    *genregraph.DAG
}

func NewService(client *Client, dag *genregraph.DAG) *Service {
	return &Service{client: client, dag: dag}
}

func (s *Service) Client() *Client {
	return s.client
}

func (s *Service) SearchReleases(ctx contextx.ContextX, query string) (*SearchResult, error) {
	return s.client.SearchDatabase(ctx, SearchProps{
		Query: query,
		Type:  SearchTypeRelease,
		Page:  PageProps{PerPage: 25},
	})
}

func (s *Service) SearchReleasesForFormat(ctx contextx.ContextX, query, format string) (*SearchResult, error) {
	return s.client.SearchDatabase(ctx, SearchProps{
		Query:  query,
		Type:   SearchTypeRelease,
		Format: format,
		Page:   PageProps{PerPage: 25},
	})
}

func (s *Service) SearchMasters(ctx contextx.ContextX, query string) (*SearchResult, error) {
	return s.client.SearchDatabase(ctx, SearchProps{
		Query: query,
		Type:  SearchTypeMaster,
		Page:  PageProps{PerPage: 25},
	})
}

func (s *Service) SearchMasterByAlbum(ctx contextx.ContextX, title, artist string) (*SearchItem, error) {
	result, err := s.client.SearchDatabase(ctx, SearchProps{
		ReleaseTitle: title,
		Artist:       artist,
		Type:         SearchTypeMaster,
		Page:         PageProps{PerPage: 1},
	})
	if err != nil {
		return nil, err
	}
	if len(result.Results) == 0 {
		return nil, nil
	}
	return &result.Results[0], nil
}

func (s *Service) SearchReleaseByAlbum(ctx contextx.ContextX, title, artist string) (*SearchItem, error) {
	result, err := s.client.SearchDatabase(ctx, SearchProps{
		ReleaseTitle: title,
		Artist:       artist,
		Type:         SearchTypeRelease,
		Page:         PageProps{PerPage: 1},
	})
	if err != nil {
		return nil, err
	}
	if len(result.Results) == 0 {
		return nil, nil
	}
	return &result.Results[0], nil
}

func (s *Service) GetMaster(ctx contextx.ContextX, id int) (*Master, error) {
	return s.client.GetMaster(ctx, id)
}

func (s *Service) GetRelease(ctx contextx.ContextX, id int) (*Release, error) {
	return s.client.GetRelease(ctx, id)
}

// ResolveAlbumGenreNodes searches Discogs for the album and resolves its genre
// and style terms against the genre graph, returning the matched nodes (Q-id +
// label). Errors are logged and suppressed so callers always get a (possibly
// empty) slice.
func (s *Service) ResolveAlbumGenreNodes(ctx contextx.ContextX, title, artist string) []*genregraph.Node {
	item, err := s.SearchMasterByAlbum(ctx, title, artist)
	if err != nil {
		slog.Warn("discogs search failed for genre resolution", "title", title, "err", err)
		return nil
	}
	if item == nil {
		item, _ = s.SearchReleaseByAlbum(ctx, title, artist)
	}
	if item == nil {
		return nil
	}
	return Resolve(s.dag, append(item.Genre, item.Style...))
}

// GetAlbumGenreSuggestions searches Discogs for the album, resolves the genres and styles
// against the genre graph, and returns the normalized genre labels.
// Errors are logged and suppressed so callers always get a (possibly empty) slice.
func (s *Service) GetAlbumGenreSuggestions(ctx contextx.ContextX, title, artist string) []string {
	nodes := s.ResolveAlbumGenreNodes(ctx, title, artist)
	labels := make([]string, 0, len(nodes))
	for _, n := range nodes {
		labels = append(labels, n.Label)
	}
	return labels
}
