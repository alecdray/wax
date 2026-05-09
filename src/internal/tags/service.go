package tags

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecdray/wax/src/internal/core/db"

	"github.com/google/uuid"
)

var invalidTagChars = regexp.MustCompile(`[^\p{L}\p{M}0-9 \-&]+`)

func normalizeTag(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = invalidTagChars.ReplaceAllString(name, "")
	return strings.TrimSpace(name)
}

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

// GetOrCreateDefaultGroups ensures "Sound" and "Mood" tag groups exist for the user.
func (s *Service) GetOrCreateDefaultGroups(ctx context.Context, userId string) ([]*TagGroupDTO, error) {
	groupNames := []string{TagGroupSound, TagGroupMood}
	groups := make([]*TagGroupDTO, 0, len(groupNames))
	for _, name := range groupNames {
		group, err := s.repo.GetOrCreateTagGroup(ctx, uuid.NewString(), userId, name)
		if err != nil {
			return nil, fmt.Errorf("failed to get or create tag group %q: %w", name, err)
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// GetUserTagGroups returns all tag groups owned by the user.
func (s *Service) GetUserTagGroups(ctx context.Context, userId string) ([]*TagGroupDTO, error) {
	groups, err := s.repo.GetTagGroupsByUserID(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag groups: %w", err)
	}
	return groups, nil
}

// GetUserTags returns all tags owned by the user (for autocomplete).
func (s *Service) GetUserTags(ctx context.Context, userId string) ([]TagDTO, error) {
	dtos, err := s.repo.GetTagsByUserID(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tags: %w", err)
	}
	return dtos, nil
}

// GetAlbumTagsByAlbumIds returns a map of albumId → []TagDTO for bulk fetching.
func (s *Service) GetAlbumTagsByAlbumIds(ctx context.Context, userId string, albumIds []string) (map[string][]TagDTO, error) {
	if len(albumIds) == 0 {
		return map[string][]TagDTO{}, nil
	}
	result, err := s.repo.GetAlbumTagsByAlbumIDs(ctx, userId, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}
	return result, nil
}

// GetAlbumTags returns the tags for a single album.
func (s *Service) GetAlbumTags(ctx context.Context, userId, albumId string) ([]TagDTO, error) {
	dtos, err := s.repo.GetAlbumTagsByAlbumID(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}
	return dtos, nil
}

// SetAlbumTags replaces the tags on an album. It resolves each name to a tag
// (get-or-create), then replaces all existing album_tags rows.
func (s *Service) SetAlbumTags(ctx context.Context, userId, albumId string, inputs []TagInput) ([]TagDTO, error) {
	var result []TagDTO

	err := s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())

		if err := txRepo.DeleteAlbumTagsByAlbumID(ctx, userId, albumId); err != nil {
			return fmt.Errorf("failed to clear album tags: %w", err)
		}

		// Cache groups so we resolve group names without re-querying per tag.
		var groupsCache []*TagGroupDTO
		groupsLoaded := false
		groupNameByID := func(id string) string {
			if id == "" {
				return ""
			}
			if !groupsLoaded {
				groups, err := txRepo.GetTagGroupsByUserID(ctx, userId)
				if err == nil {
					groupsCache = groups
				}
				groupsLoaded = true
			}
			for _, g := range groupsCache {
				if g.ID == id {
					return g.Name
				}
			}
			return ""
		}

		for _, input := range inputs {
			input.Name = normalizeTag(input.Name)
			if input.Name == "" {
				continue
			}

			tag, err := txRepo.GetOrCreateTag(ctx, uuid.NewString(), userId, input.Name, input.GroupID)
			if err != nil {
				return fmt.Errorf("failed to get or create tag %q: %w", input.Name, err)
			}

			if err := txRepo.CreateAlbumTag(ctx, uuid.NewString(), userId, albumId, tag.ID); err != nil {
				return fmt.Errorf("failed to create album tag: %w", err)
			}

			// Repo returns Group with ID only; resolve the name here.
			if tag.Group != nil && tag.Group.Name == "" {
				tag.Group.Name = groupNameByID(tag.Group.ID)
			}

			result = append(result, tag)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
