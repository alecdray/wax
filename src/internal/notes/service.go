package notes

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

const MaxSleeveNoteLength = 10_000

type AlbumNoteDTO struct {
	ID        string
	UserID    string
	AlbumID   string
	Content   string
	UpdatedAt time.Time
}

func newAlbumNoteDTOFromModel(m sqlc.AlbumNote) *AlbumNoteDTO {
	return &AlbumNoteDTO{
		ID:        m.ID,
		UserID:    m.UserID,
		AlbumID:   m.AlbumID,
		Content:   m.Content,
		UpdatedAt: m.UpdatedAt,
	}
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.Linkify),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

// RenderMarkdown converts markdown content to safe HTML.
// All links are given target="_blank" rel="noopener noreferrer".
func RenderMarkdown(content string) template.HTML {
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(content))
	}
	out := strings.ReplaceAll(buf.String(), "<a ", `<a target="_blank" rel="noopener noreferrer" `)
	return template.HTML(out)
}

type Service struct {
	db *db.DB
}

func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

// UpsertAlbumNote creates or updates the sleeve note for an album.
func (s *Service) UpsertAlbumNote(ctx context.Context, userID, albumID, content string) (*AlbumNoteDTO, error) {
	model, err := s.db.Queries().UpsertAlbumNote(ctx, sqlc.UpsertAlbumNoteParams{
		ID:      uuid.NewString(),
		UserID:  userID,
		AlbumID: albumID,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert album note: %w", err)
	}
	return newAlbumNoteDTOFromModel(model), nil
}

// GetAlbumNote returns the sleeve note for an album, or nil if none exists.
func (s *Service) GetAlbumNote(ctx context.Context, userID, albumID string) (*AlbumNoteDTO, error) {
	model, err := s.db.Queries().GetAlbumNote(ctx, sqlc.GetAlbumNoteParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get album note: %w", err)
	}
	return newAlbumNoteDTOFromModel(model), nil
}

// GetAlbumNotesByAlbumIds returns a map of albumID → *AlbumNoteDTO for bulk fetching.
func (s *Service) GetAlbumNotesByAlbumIds(ctx context.Context, userID string, albumIDs []string) (map[string]*AlbumNoteDTO, error) {
	if len(albumIDs) == 0 {
		return map[string]*AlbumNoteDTO{}, nil
	}
	rows, err := s.db.Queries().GetAlbumNotesByAlbumIds(ctx, sqlc.GetAlbumNotesByAlbumIdsParams{
		UserID:   userID,
		AlbumIds: albumIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album notes: %w", err)
	}
	result := make(map[string]*AlbumNoteDTO, len(rows))
	for _, row := range rows {
		dto := newAlbumNoteDTOFromModel(row)
		result[row.AlbumID] = dto
	}
	return result, nil
}
