package crates

import (
	"fmt"
	"strings"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/google/uuid"
)

type Service struct {
	repo    *Repo
	library *library.Service
}

func NewService(repo *Repo, library *library.Service) *Service {
	return &Service{repo: repo, library: library}
}

// ListCrates returns all crates owned by the authenticated user.
func (s *Service) ListCrates(ctx contextx.ContextX) ([]CrateDTO, error) {
	userID, err := ctx.UserId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}
	dtos, err := s.repo.ListCrates(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list crates: %w", err)
	}
	return dtos, nil
}

// GetCrate returns a crate with its albums fully hydrated.
func (s *Service) GetCrate(ctx contextx.ContextX, id string) (CrateDetailDTO, error) {
	userID, err := ctx.UserId()
	if err != nil {
		return CrateDetailDTO{}, fmt.Errorf("failed to get user id: %w", err)
	}
	base, err := s.repo.GetCrate(ctx, id, userID)
	if err != nil {
		return CrateDetailDTO{}, fmt.Errorf("failed to get crate: %w", err)
	}
	albumIDs, err := s.repo.GetCrateAlbumIDs(ctx, id)
	if err != nil {
		return CrateDetailDTO{}, fmt.Errorf("failed to get crate album ids: %w", err)
	}
	albums, err := s.library.GetAlbumsByIDs(ctx, albumIDs)
	if err != nil {
		return CrateDetailDTO{}, fmt.Errorf("failed to hydrate crate albums: %w", err)
	}
	return CrateDetailDTO{
		CrateDTO: base,
		Albums:   albums,
	}, nil
}

// CreateCrate creates a new named crate for the authenticated user.
// Returns an error if name is empty after trimming.
func (s *Service) CreateCrate(ctx contextx.ContextX, name string) (CrateDTO, error) {
	userID, err := ctx.UserId()
	if err != nil {
		return CrateDTO{}, fmt.Errorf("failed to get user id: %w", err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return CrateDTO{}, fmt.Errorf("crate name must not be empty")
	}
	dto, err := s.repo.InsertCrate(ctx, uuid.NewString(), userID, name)
	if err != nil {
		return CrateDTO{}, fmt.Errorf("failed to create crate: %w", err)
	}
	return dto, nil
}

// DeleteCrate removes a crate owned by the authenticated user.
func (s *Service) DeleteCrate(ctx contextx.ContextX, id string) error {
	userID, err := ctx.UserId()
	if err != nil {
		return fmt.Errorf("failed to get user id: %w", err)
	}
	if err := s.repo.DeleteCrate(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete crate: %w", err)
	}
	return nil
}

// AddAlbum adds an album to a crate. Idempotent — duplicate adds are silently ignored.
func (s *Service) AddAlbum(ctx contextx.ContextX, crateID, albumID string) error {
	if err := s.repo.InsertCrateAlbum(ctx, uuid.NewString(), crateID, albumID); err != nil {
		return fmt.Errorf("failed to add album to crate: %w", err)
	}
	return nil
}

// RemoveAlbum removes an album from a crate.
func (s *Service) RemoveAlbum(ctx contextx.ContextX, crateID, albumID string) error {
	if err := s.repo.DeleteCrateAlbum(ctx, crateID, albumID); err != nil {
		return fmt.Errorf("failed to remove album from crate: %w", err)
	}
	return nil
}

// SearchNonMembers returns library albums that are not already in the crate,
// filtered by q (case-insensitive substring match against title or artist names).
// An empty q returns all non-members.
func (s *Service) SearchNonMembers(ctx contextx.ContextX, crateID, q string) ([]library.AlbumDTO, error) {
	userID, err := ctx.UserId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}
	memberIDs, err := s.repo.GetCrateAlbumIDs(ctx, crateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get crate album ids: %w", err)
	}
	memberSet := make(map[string]struct{}, len(memberIDs))
	for _, id := range memberIDs {
		memberSet[id] = struct{}{}
	}
	all, err := s.library.GetAlbumsInLibrary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get library albums: %w", err)
	}
	needle := strings.ToLower(q)
	var out []library.AlbumDTO
	for _, album := range all {
		if _, isMember := memberSet[album.ID]; isMember {
			continue
		}
		if needle == "" {
			out = append(out, album)
			continue
		}
		if strings.Contains(strings.ToLower(album.Title), needle) {
			out = append(out, album)
			continue
		}
		for _, artist := range album.Artists {
			if strings.Contains(strings.ToLower(artist.Name), needle) {
				out = append(out, album)
				break
			}
		}
	}
	return out, nil
}
