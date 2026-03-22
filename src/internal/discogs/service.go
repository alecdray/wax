package discogs

import (
	"log/slog"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/genres"
)

type Service struct {
	client *Client
	dag    *genres.DAG
}

func NewService(client *Client, dag *genres.DAG) *Service {
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

// GetAlbumGenreSuggestions searches Discogs for the album, resolves the genres and styles
// against the genre DAG, and returns the normalized genre labels.
// Errors are logged and suppressed so callers always get a (possibly empty) slice.
func (s *Service) GetAlbumGenreSuggestions(ctx contextx.ContextX, title, artist string) []string {
	item, err := s.SearchMasterByAlbum(ctx, title, artist)
	if err != nil {
		slog.Warn("discogs search failed for genre suggestions", "title", title, "err", err)
		return nil
	}
	if item == nil {
		item, err = s.SearchReleaseByAlbum(ctx, title, artist)
	}
	return resolveItemGenres(s.dag, item)
}
