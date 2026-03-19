package discogs

import "github.com/alecdray/wax/src/internal/core/contextx"

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client: client}
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
