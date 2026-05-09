package tags

import (
	"context"
	"database/sql"

	"github.com/alecdray/wax/src/internal/core/db/sqlc"
)

// Repo is the tags module's data access layer. It is the only file in
// package tags that imports core/db/sqlc. Repo methods return tag DTOs —
// never sqlc.* types.
type Repo struct {
	q *sqlc.Queries
}

// NewRepo binds a Repo to the given Queries. Callers can bind to db.Queries()
// for the global handle or to tx.Queries() inside a db.WithTx callback for
// transactional work.
func NewRepo(q *sqlc.Queries) *Repo {
	return &Repo{q: q}
}

// --- DTO conversion helpers (private — only repo.go touches sqlc types) ---

func newTagGroupDTO(id, name string) *TagGroupDTO {
	if id == "" {
		return nil
	}
	return &TagGroupDTO{ID: id, Name: name}
}

// tagDTOFromRow constructs a TagDTO from query rows that use COALESCE for group fields.
func tagDTOFromRow(tag sqlc.Tag, groupIDValue, groupName string) TagDTO {
	dto := TagDTO{
		ID:   tag.ID,
		Name: tag.Name,
	}
	if tag.GroupID.Valid && groupIDValue != "" {
		dto.Group = newTagGroupDTO(groupIDValue, groupName)
	}
	return dto
}

// --- Tag group lookups / mutations ---

// GetOrCreateTagGroup ensures a tag group with the given name exists for the user
// and returns its DTO.
func (r *Repo) GetOrCreateTagGroup(ctx context.Context, id, userID, name string) (*TagGroupDTO, error) {
	model, err := r.q.GetOrCreateTagGroup(ctx, sqlc.GetOrCreateTagGroupParams{
		ID:     id,
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		return nil, err
	}
	return &TagGroupDTO{ID: model.ID, Name: model.Name}, nil
}

// GetTagGroupsByUserID returns all tag groups owned by the user.
func (r *Repo) GetTagGroupsByUserID(ctx context.Context, userID string) ([]*TagGroupDTO, error) {
	models, err := r.q.GetTagGroupsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]*TagGroupDTO, 0, len(models))
	for _, m := range models {
		out = append(out, &TagGroupDTO{ID: m.ID, Name: m.Name})
	}
	return out, nil
}

// --- Tag lookups ---

// GetTagsByUserID returns all tags owned by the user.
func (r *Repo) GetTagsByUserID(ctx context.Context, userID string) ([]TagDTO, error) {
	rows, err := r.q.GetTagsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]TagDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, tagDTOFromRow(row.Tag, row.GroupIDValue, row.GroupName))
	}
	return out, nil
}

// GetAlbumTagsByAlbumID returns the tags for a single album.
func (r *Repo) GetAlbumTagsByAlbumID(ctx context.Context, userID, albumID string) ([]TagDTO, error) {
	rows, err := r.q.GetAlbumTagsByAlbumId(ctx, sqlc.GetAlbumTagsByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]TagDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, tagDTOFromRow(row.Tag, row.GroupIDValue, row.GroupName))
	}
	return out, nil
}

// GetAlbumTagsByAlbumIDs returns tags grouped by album ID.
func (r *Repo) GetAlbumTagsByAlbumIDs(ctx context.Context, userID string, albumIDs []string) (map[string][]TagDTO, error) {
	rows, err := r.q.GetAlbumTagsByAlbumIds(ctx, sqlc.GetAlbumTagsByAlbumIdsParams{
		UserID:   userID,
		AlbumIds: albumIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make(map[string][]TagDTO, len(albumIDs))
	for _, row := range rows {
		result[row.AlbumID] = append(result[row.AlbumID], tagDTOFromRow(row.Tag, row.GroupIDValue, row.GroupName))
	}
	return result, nil
}

// --- Album-tag mutations ---

// DeleteAlbumTagsByAlbumID removes all tag associations for one user/album.
func (r *Repo) DeleteAlbumTagsByAlbumID(ctx context.Context, userID, albumID string) error {
	return r.q.DeleteAlbumTagsByAlbumId(ctx, sqlc.DeleteAlbumTagsByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
}

// GetOrCreateTag ensures a tag with the given name (and optional group) exists
// for the user and returns its DTO. The returned DTO's Group is nil unless the
// caller resolves the group name (groups are looked up separately because the
// underlying query doesn't join tag_groups).
func (r *Repo) GetOrCreateTag(ctx context.Context, id, userID, name, groupID string) (TagDTO, error) {
	groupNullable := sql.NullString{}
	if groupID != "" {
		groupNullable = sql.NullString{String: groupID, Valid: true}
	}
	tag, err := r.q.GetOrCreateTag(ctx, sqlc.GetOrCreateTagParams{
		ID:      id,
		UserID:  userID,
		Name:    name,
		GroupID: groupNullable,
	})
	if err != nil {
		return TagDTO{}, err
	}
	dto := TagDTO{ID: tag.ID, Name: tag.Name}
	if tag.GroupID.Valid {
		// Group name is left empty here; callers that need it resolve it via
		// GetTagGroupsByUserID. Keep the ID so consumers know the group exists.
		dto.Group = &TagGroupDTO{ID: tag.GroupID.String}
	}
	return dto, nil
}

// CreateAlbumTag links an existing tag to an album for one user.
func (r *Repo) CreateAlbumTag(ctx context.Context, id, userID, albumID, tagID string) error {
	_, err := r.q.CreateAlbumTag(ctx, sqlc.CreateAlbumTagParams{
		ID:      id,
		UserID:  userID,
		AlbumID: albumID,
		TagID:   tagID,
	})
	return err
}
