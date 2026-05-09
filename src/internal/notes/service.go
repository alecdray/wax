package notes

import (
	"context"
	"fmt"

	"github.com/alecdray/wax/src/internal/core/db"
)

type Service struct {
	db   *db.DB
	repo *Repo
}

func NewService(d *db.DB) *Service {
	return &Service{
		db:   d,
		repo: NewRepo(d.Queries()),
	}
}

// UpsertAlbumNote creates or updates the sleeve note for an album.
func (s *Service) UpsertAlbumNote(ctx context.Context, userID, albumID, content string) (*AlbumNoteDTO, error) {
	dto, err := s.repo.UpsertAlbumNote(ctx, userID, albumID, content)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert album note: %w", err)
	}
	return dto, nil
}

// GetAlbumNote returns the sleeve note for an album, or nil if none exists.
func (s *Service) GetAlbumNote(ctx context.Context, userID, albumID string) (*AlbumNoteDTO, error) {
	dto, err := s.repo.GetAlbumNote(ctx, userID, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album note: %w", err)
	}
	return dto, nil
}

// GetAlbumNotesByAlbumIds returns a map of albumID → *AlbumNoteDTO for bulk fetching.
func (s *Service) GetAlbumNotesByAlbumIds(ctx context.Context, userID string, albumIDs []string) (map[string]*AlbumNoteDTO, error) {
	if len(albumIDs) == 0 {
		return map[string]*AlbumNoteDTO{}, nil
	}
	result, err := s.repo.GetAlbumNotesByAlbumIDs(ctx, userID, albumIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get album notes: %w", err)
	}
	return result, nil
}
