package tags

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"

	"github.com/google/uuid"
)

type Genre string

const (
	GenreClassical  Genre = "Classical"
	GenreCountry    Genre = "Country"
	GenreElectronic Genre = "Electronic"
	GenreHipHop     Genre = "Hip-hop"
	GenreJazz       Genre = "Jazz"
	GenreLatin      Genre = "Latin"
	GenrePop        Genre = "Pop"
	GenrePunk       Genre = "Punk"
	GenreReggae     Genre = "Reggae"
	GenreRock       Genre = "Rock"
	GenreMetal      Genre = "Metal"
	GenreFunk       Genre = "Funk"
	GenreSoul       Genre = "Soul"
	GenreRnB        Genre = "R&B"
	GenreBlues      Genre = "Blues"
	GenreFolk       Genre = "Folk"
	GenreWorld      Genre = "World"
	GenreTheater    Genre = "Theater"
	GenreOther      Genre = "Other"
	GenreUnknown    Genre = "Unknown"
)

type Subgenre string

const ()

var invalidTagChars = regexp.MustCompile(`[^\p{L}\p{M}0-9 \-&]+`)

func normalizeTag(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = invalidTagChars.ReplaceAllString(name, "")
	return strings.TrimSpace(name)
}

const (
	TagGroupSound = "Sound"
	TagGroupMood  = "Mood"
)

type TagGroupDTO struct {
	ID   string
	Name string
}

type TagDTO struct {
	ID    string
	Name  string
	Group *TagGroupDTO
}

type TagInput struct {
	Name    string
	GroupID string // empty = ungrouped
}

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

func newTagGroupDTO(id, name string) *TagGroupDTO {
	if id == "" {
		return nil
	}
	return &TagGroupDTO{ID: id, Name: name}
}

// newTagDTOFromRow constructs a TagDTO from query rows that use COALESCE for group fields.
func newTagDTOFromRow(tag sqlc.Tag, groupIDValue, groupName string) TagDTO {
	dto := TagDTO{
		ID:   tag.ID,
		Name: tag.Name,
	}
	if tag.GroupID.Valid && groupIDValue != "" {
		dto.Group = newTagGroupDTO(groupIDValue, groupName)
	}
	return dto
}

// GetOrCreateDefaultGroups ensures "Sound" and "Mood" tag groups exist for the user.
func (s *Service) GetOrCreateDefaultGroups(ctx context.Context, userId string) ([]*TagGroupDTO, error) {
	groupNames := []string{TagGroupSound, TagGroupMood}
	groups := make([]*TagGroupDTO, 0, len(groupNames))
	for _, name := range groupNames {
		model, err := s.db.Queries().GetOrCreateTagGroup(ctx, sqlc.GetOrCreateTagGroupParams{
			ID:     uuid.NewString(),
			UserID: userId,
			Name:   name,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get or create tag group %q: %w", name, err)
		}
		groups = append(groups, &TagGroupDTO{ID: model.ID, Name: model.Name})
	}
	return groups, nil
}

// GetUserTagGroups returns all tag groups owned by the user.
func (s *Service) GetUserTagGroups(ctx context.Context, userId string) ([]*TagGroupDTO, error) {
	models, err := s.db.Queries().GetTagGroupsByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag groups: %w", err)
	}
	dtos := make([]*TagGroupDTO, 0, len(models))
	for _, m := range models {
		dtos = append(dtos, &TagGroupDTO{ID: m.ID, Name: m.Name})
	}
	return dtos, nil
}

// GetUserTags returns all tags owned by the user (for autocomplete).
func (s *Service) GetUserTags(ctx context.Context, userId string) ([]TagDTO, error) {
	rows, err := s.db.Queries().GetTagsByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tags: %w", err)
	}
	dtos := make([]TagDTO, 0, len(rows))
	for _, row := range rows {
		dtos = append(dtos, newTagDTOFromRow(row.Tag, row.GroupIDValue, row.GroupName))
	}
	return dtos, nil
}

// GetAlbumTagsByAlbumIds returns a map of albumId → []TagDTO for bulk fetching.
func (s *Service) GetAlbumTagsByAlbumIds(ctx context.Context, userId string, albumIds []string) (map[string][]TagDTO, error) {
	if len(albumIds) == 0 {
		return map[string][]TagDTO{}, nil
	}
	rows, err := s.db.Queries().GetAlbumTagsByAlbumIds(ctx, sqlc.GetAlbumTagsByAlbumIdsParams{
		UserID:   userId,
		AlbumIds: albumIds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}
	result := make(map[string][]TagDTO, len(albumIds))
	for _, row := range rows {
		result[row.AlbumID] = append(result[row.AlbumID], newTagDTOFromRow(row.Tag, row.GroupIDValue, row.GroupName))
	}
	return result, nil
}

// GetAlbumTags returns the tags for a single album.
func (s *Service) GetAlbumTags(ctx context.Context, userId, albumId string) ([]TagDTO, error) {
	rows, err := s.db.Queries().GetAlbumTagsByAlbumId(ctx, sqlc.GetAlbumTagsByAlbumIdParams{
		UserID:  userId,
		AlbumID: albumId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}
	dtos := make([]TagDTO, 0, len(rows))
	for _, row := range rows {
		dtos = append(dtos, newTagDTOFromRow(row.Tag, row.GroupIDValue, row.GroupName))
	}
	return dtos, nil
}

// SetAlbumTags replaces the tags on an album. It resolves each name to a tag
// (get-or-create), then replaces all existing album_tags rows.
func (s *Service) SetAlbumTags(ctx context.Context, userId, albumId string, inputs []TagInput) ([]TagDTO, error) {
	var result []TagDTO

	err := s.db.WithTx(func(tx *db.DB) error {
		if err := tx.Queries().DeleteAlbumTagsByAlbumId(ctx, sqlc.DeleteAlbumTagsByAlbumIdParams{
			UserID:  userId,
			AlbumID: albumId,
		}); err != nil {
			return fmt.Errorf("failed to clear album tags: %w", err)
		}

		for _, input := range inputs {
			input.Name = normalizeTag(input.Name)
			if input.Name == "" {
				continue
			}

			groupID := sql.NullString{}
			if input.GroupID != "" {
				groupID = sql.NullString{String: input.GroupID, Valid: true}
			}

			tag, err := tx.Queries().GetOrCreateTag(ctx, sqlc.GetOrCreateTagParams{
				ID:      uuid.NewString(),
				UserID:  userId,
				Name:    input.Name,
				GroupID: groupID,
			})
			if err != nil {
				return fmt.Errorf("failed to get or create tag %q: %w", input.Name, err)
			}

			_, err = tx.Queries().CreateAlbumTag(ctx, sqlc.CreateAlbumTagParams{
				ID:      uuid.NewString(),
				UserID:  userId,
				AlbumID: albumId,
				TagID:   tag.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to create album tag: %w", err)
			}

			// Resolve group name for DTO from the tag's stored group_id
			groupIDValue := ""
			groupName := ""
			if tag.GroupID.Valid {
				groupIDValue = tag.GroupID.String
				// Look up the group name
				groups, err := tx.Queries().GetTagGroupsByUserId(ctx, userId)
				if err == nil {
					for _, g := range groups {
						if g.ID == tag.GroupID.String {
							groupName = g.Name
							break
						}
					}
				}
			}

			result = append(result, newTagDTOFromRow(tag, groupIDValue, groupName))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
