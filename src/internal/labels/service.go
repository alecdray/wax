package labels

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/alecdray/wax/src/internal/genres"

	"github.com/google/uuid"
)

var invalidLabelChars = regexp.MustCompile(`[^\p{L}\p{M}0-9 \-&]+`)

func normalizeLabel(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = invalidLabelChars.ReplaceAllString(name, "")
	return strings.TrimSpace(name)
}

// GenreDTO is a genre assigned to an album.
type GenreDTO struct {
	ID    string // Wikidata QID, e.g. "Q11399"
	Label string // display label snapshot
}

// GenreSuggestion is returned from genre search, including parent breadcrumb.
type GenreSuggestion struct {
	ID          string
	Label       string
	ParentLabel string
}

// AlbumLabels holds all three label types for one album.
type AlbumLabels struct {
	Genres   []GenreDTO
	Moods    []string
	UserTags []string
}

type Service struct {
	db  *db.DB
	dag *genres.DAG
}

func NewService(db *db.DB, dag *genres.DAG) *Service {
	return &Service{db: db, dag: dag}
}

// SearchGenres searches the DAG for genres matching query. DAG-only, no DB hit.
func (s *Service) SearchGenres(query string) []GenreSuggestion {
	if s.dag == nil || query == "" {
		return nil
	}
	nodes := s.dag.Search(query)
	results := make([]GenreSuggestion, 0, len(nodes))
	for _, n := range nodes {
		sug := GenreSuggestion{ID: n.ID, Label: n.Label}
		if len(n.Parents) > 0 {
			sug.ParentLabel = n.Parents[0].Label
		}
		results = append(results, sug)
	}
	return results
}

// SetAlbumGenres replaces genre assignments for an album. genreIDs are Wikidata QIDs.
func (s *Service) SetAlbumGenres(ctx context.Context, userID, albumID string, genreIDs []string) ([]GenreDTO, error) {
	var result []GenreDTO

	err := s.db.WithTx(func(tx *db.DB) error {
		if err := tx.Queries().DeleteAlbumGenresByAlbumId(ctx, sqlc.DeleteAlbumGenresByAlbumIdParams{
			UserID:  userID,
			AlbumID: albumID,
		}); err != nil {
			return fmt.Errorf("failed to clear album genres: %w", err)
		}

		for _, gid := range genreIDs {
			gid = strings.TrimSpace(gid)
			if gid == "" {
				continue
			}
			label := gid
			if s.dag != nil {
				if node := s.dag.Get(gid); node != nil {
					label = node.Label
				}
			}

			row, err := tx.Queries().CreateAlbumGenre(ctx, sqlc.CreateAlbumGenreParams{
				ID:         uuid.NewString(),
				UserID:     userID,
				AlbumID:    albumID,
				GenreID:    gid,
				GenreLabel: label,
			})
			if err != nil {
				return fmt.Errorf("failed to create album genre: %w", err)
			}
			result = append(result, GenreDTO{ID: row.GenreID, Label: row.GenreLabel})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAlbumGenresByAlbumIds returns a map of albumID → []GenreDTO.
func (s *Service) GetAlbumGenresByAlbumIds(ctx context.Context, userID string, albumIDs []string) (map[string][]GenreDTO, error) {
	if len(albumIDs) == 0 {
		return map[string][]GenreDTO{}, nil
	}
	rows, err := s.db.Queries().GetAlbumGenresByAlbumIds(ctx, sqlc.GetAlbumGenresByAlbumIdsParams{
		UserID:   userID,
		AlbumIds: albumIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album genres: %w", err)
	}
	result := make(map[string][]GenreDTO, len(albumIDs))
	for _, row := range rows {
		result[row.AlbumID] = append(result[row.AlbumID], GenreDTO{ID: row.GenreID, Label: row.GenreLabel})
	}
	return result, nil
}

// GetAlbumGenres returns genres for a single album.
func (s *Service) GetAlbumGenres(ctx context.Context, userID, albumID string) ([]GenreDTO, error) {
	rows, err := s.db.Queries().GetAlbumGenresByAlbumId(ctx, sqlc.GetAlbumGenresByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album genres: %w", err)
	}
	result := make([]GenreDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, GenreDTO{ID: row.GenreID, Label: row.GenreLabel})
	}
	return result, nil
}

// SetAlbumMoods replaces mood assignments for an album.
func (s *Service) SetAlbumMoods(ctx context.Context, userID, albumID string, moods []string) ([]string, error) {
	var result []string

	err := s.db.WithTx(func(tx *db.DB) error {
		if err := tx.Queries().DeleteAlbumMoodsByAlbumId(ctx, sqlc.DeleteAlbumMoodsByAlbumIdParams{
			UserID:  userID,
			AlbumID: albumID,
		}); err != nil {
			return fmt.Errorf("failed to clear album moods: %w", err)
		}

		for _, m := range moods {
			m = normalizeLabel(m)
			if m == "" {
				continue
			}
			row, err := tx.Queries().CreateAlbumMood(ctx, sqlc.CreateAlbumMoodParams{
				ID:      uuid.NewString(),
				UserID:  userID,
				AlbumID: albumID,
				Mood:    m,
			})
			if err != nil {
				return fmt.Errorf("failed to create album mood: %w", err)
			}
			result = append(result, row.Mood)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAlbumMoodsByAlbumIds returns a map of albumID → []string.
func (s *Service) GetAlbumMoodsByAlbumIds(ctx context.Context, userID string, albumIDs []string) (map[string][]string, error) {
	if len(albumIDs) == 0 {
		return map[string][]string{}, nil
	}
	rows, err := s.db.Queries().GetAlbumMoodsByAlbumIds(ctx, sqlc.GetAlbumMoodsByAlbumIdsParams{
		UserID:   userID,
		AlbumIds: albumIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album moods: %w", err)
	}
	result := make(map[string][]string, len(albumIDs))
	for _, row := range rows {
		result[row.AlbumID] = append(result[row.AlbumID], row.Mood)
	}
	return result, nil
}

// GetAlbumMoods returns moods for a single album.
func (s *Service) GetAlbumMoods(ctx context.Context, userID, albumID string) ([]string, error) {
	rows, err := s.db.Queries().GetAlbumMoodsByAlbumId(ctx, sqlc.GetAlbumMoodsByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album moods: %w", err)
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Mood)
	}
	return result, nil
}

// GetDistinctUserMoods returns all distinct moods the user has used.
func (s *Service) GetDistinctUserMoods(ctx context.Context, userID string) ([]string, error) {
	moods, err := s.db.Queries().GetDistinctUserMoods(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct moods: %w", err)
	}
	return moods, nil
}

// SetAlbumUserTags replaces user tag assignments for an album.
func (s *Service) SetAlbumUserTags(ctx context.Context, userID, albumID string, tags []string) ([]string, error) {
	var result []string

	err := s.db.WithTx(func(tx *db.DB) error {
		if err := tx.Queries().DeleteAlbumUserTagsByAlbumId(ctx, sqlc.DeleteAlbumUserTagsByAlbumIdParams{
			UserID:  userID,
			AlbumID: albumID,
		}); err != nil {
			return fmt.Errorf("failed to clear album user tags: %w", err)
		}

		for _, t := range tags {
			t = normalizeLabel(t)
			if t == "" {
				continue
			}
			row, err := tx.Queries().CreateAlbumUserTag(ctx, sqlc.CreateAlbumUserTagParams{
				ID:      uuid.NewString(),
				UserID:  userID,
				AlbumID: albumID,
				Tag:     t,
			})
			if err != nil {
				return fmt.Errorf("failed to create album user tag: %w", err)
			}
			result = append(result, row.Tag)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAlbumUserTagsByAlbumIds returns a map of albumID → []string.
func (s *Service) GetAlbumUserTagsByAlbumIds(ctx context.Context, userID string, albumIDs []string) (map[string][]string, error) {
	if len(albumIDs) == 0 {
		return map[string][]string{}, nil
	}
	rows, err := s.db.Queries().GetAlbumUserTagsByAlbumIds(ctx, sqlc.GetAlbumUserTagsByAlbumIdsParams{
		UserID:   userID,
		AlbumIds: albumIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album user tags: %w", err)
	}
	result := make(map[string][]string, len(albumIDs))
	for _, row := range rows {
		result[row.AlbumID] = append(result[row.AlbumID], row.Tag)
	}
	return result, nil
}

// GetAlbumUserTags returns user tags for a single album.
func (s *Service) GetAlbumUserTags(ctx context.Context, userID, albumID string) ([]string, error) {
	rows, err := s.db.Queries().GetAlbumUserTagsByAlbumId(ctx, sqlc.GetAlbumUserTagsByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get album user tags: %w", err)
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Tag)
	}
	return result, nil
}

// GetDistinctUserTags returns all distinct user tags the user has used.
func (s *Service) GetDistinctUserTags(ctx context.Context, userID string) ([]string, error) {
	tags, err := s.db.Queries().GetDistinctUserTags(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct user tags: %w", err)
	}
	return tags, nil
}

// GetAlbumLabelsByAlbumIds fetches all three label types for a set of albums.
func (s *Service) GetAlbumLabelsByAlbumIds(ctx context.Context, userID string, albumIDs []string) (map[string]AlbumLabels, error) {
	genresByAlbum, err := s.GetAlbumGenresByAlbumIds(ctx, userID, albumIDs)
	if err != nil {
		return nil, err
	}
	moodsByAlbum, err := s.GetAlbumMoodsByAlbumIds(ctx, userID, albumIDs)
	if err != nil {
		return nil, err
	}
	tagsByAlbum, err := s.GetAlbumUserTagsByAlbumIds(ctx, userID, albumIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]AlbumLabels, len(albumIDs))
	for _, id := range albumIDs {
		result[id] = AlbumLabels{
			Genres:   genresByAlbum[id],
			Moods:    moodsByAlbum[id],
			UserTags: tagsByAlbum[id],
		}
	}
	return result, nil
}

// GetAlbumLabels fetches all three label types for a single album.
func (s *Service) GetAlbumLabels(ctx context.Context, userID, albumID string) (AlbumLabels, error) {
	genrs, err := s.GetAlbumGenres(ctx, userID, albumID)
	if err != nil {
		return AlbumLabels{}, err
	}
	moods, err := s.GetAlbumMoods(ctx, userID, albumID)
	if err != nil {
		return AlbumLabels{}, err
	}
	tags, err := s.GetAlbumUserTags(ctx, userID, albumID)
	if err != nil {
		return AlbumLabels{}, err
	}
	return AlbumLabels{Genres: genrs, Moods: moods, UserTags: tags}, nil
}
