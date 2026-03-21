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
