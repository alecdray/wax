package library

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/alecdray/wax/src/internal/discogs"

	"github.com/google/uuid"
)

// PhysicalFormats lists the physical release formats wax tracks for a library album.
var PhysicalFormats = []models.ReleaseFormat{
	models.ReleaseFormatVinyl,
	models.ReleaseFormatCD,
	models.ReleaseFormatCassette,
}

// IsPhysicalFormat reports whether the given format is one wax tracks as a physical release.
func IsPhysicalFormat(f models.ReleaseFormat) bool {
	for _, pf := range PhysicalFormats {
		if pf == f {
			return true
		}
	}
	return false
}

// NormalizeDiscogsReleasedDate produces a YYYY-MM-DD date string from a Discogs Release plus
// a fallback year. Discogs's Release.Released can be "YYYY-MM-DD", "YYYY", or empty; this
// function picks the most precise option and pads bare years to YYYY-01-01.
func NormalizeDiscogsReleasedDate(release *discogs.Release, fallbackYear string) string {
	releasedDate := fallbackYear
	if release != nil {
		if len(release.Released) >= 10 {
			releasedDate = release.Released
		} else if len(release.Released) >= 4 {
			releasedDate = release.Released[:4] + "-01-01"
		} else if release.Year > 0 {
			releasedDate = strconv.Itoa(release.Year) + "-01-01"
		}
	}
	if len(releasedDate) == 4 {
		releasedDate += "-01-01"
	}
	return releasedDate
}

type ReleaseDTO struct {
	ID         string
	AlbumID    string
	Format     models.ReleaseFormat
	AddedAt    *time.Time
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}

func NewReleaseDTOFromModel(model sqlc.Release, userRelease *sqlc.UserRelease) ReleaseDTO {
	dto := ReleaseDTO{
		ID:        model.ID,
		AlbumID:   model.AlbumID,
		Format:    model.Format,
		DiscogsID: model.DiscogsID.String,
		Label:     model.Label.String,
	}
	if model.ReleasedAt.Valid {
		dto.ReleasedAt = &model.ReleasedAt.Time
	}
	if userRelease != nil {
		dto.AddedAt = &userRelease.AddedAt
	}
	return dto
}

type ReleaseDTOs []ReleaseDTO

func (releases ReleaseDTOs) OldestAddedAtDate() *time.Time {
	var oldest *time.Time
	for _, r := range releases {
		if r.AddedAt != nil {
			if oldest == nil || r.AddedAt.Before(*oldest) {
				oldest = r.AddedAt
			}
		}
	}
	return oldest
}

func (releases ReleaseDTOs) FindFormat(format models.ReleaseFormat) *ReleaseDTO {
	for _, r := range releases {
		if r.Format == format {
			return &r
		}
	}
	return nil
}

// AlbumFormatDTO represents one format row in the formats modal.
// It exists for all 4 formats regardless of whether the user owns that format.
type AlbumFormatDTO struct {
	Format     models.ReleaseFormat
	ReleaseID  string // empty if this format has never been added for this album
	Owned      bool
	AddedAt    *time.Time
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}

var allFormats = []models.ReleaseFormat{
	models.ReleaseFormatDigital,
	models.ReleaseFormatVinyl,
	models.ReleaseFormatCD,
	models.ReleaseFormatCassette,
}

func albumFormatDTOFromRelease(r sqlc.Release, ur *sqlc.UserRelease) AlbumFormatDTO {
	dto := AlbumFormatDTO{
		Format:    r.Format,
		ReleaseID: r.ID,
		DiscogsID: r.DiscogsID.String,
		Label:     r.Label.String,
	}
	if r.ReleasedAt.Valid {
		dto.ReleasedAt = &r.ReleasedAt.Time
	}
	if ur != nil {
		dto.Owned = true
		dto.AddedAt = &ur.AddedAt
	}
	return dto
}

func buildAlbumFormats(releases []sqlc.Release, userReleases []sqlc.GetUserReleasesByAlbumIdRow) []AlbumFormatDTO {
	releaseByFormat := make(map[models.ReleaseFormat]sqlc.Release, len(releases))
	for _, r := range releases {
		releaseByFormat[r.Format] = r
	}

	ownedByReleaseID := make(map[string]sqlc.UserRelease, len(userReleases))
	for _, ur := range userReleases {
		ownedByReleaseID[ur.Release.ID] = ur.UserRelease
	}

	result := make([]AlbumFormatDTO, len(allFormats))
	for i, format := range allFormats {
		if r, ok := releaseByFormat[format]; ok {
			var ur *sqlc.UserRelease
			if entry, owned := ownedByReleaseID[r.ID]; owned {
				ur = &entry
			}
			result[i] = albumFormatDTOFromRelease(r, ur)
		} else {
			result[i] = AlbumFormatDTO{Format: format}
		}
	}

	return result
}

type SaveFormatInput struct {
	Format     models.ReleaseFormat
	Owned      bool
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
}

func (s *Service) GetReleasesInLibrary(ctx context.Context, userId string) ([]ReleaseDTO, error) {
	releases, err := s.db.Queries().GetUserReleases(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get user releases: %w", err)
		return nil, err
	}

	var releaseDTOs []ReleaseDTO
	for _, release := range releases {
		releaseDTOs = append(releaseDTOs, NewReleaseDTOFromModel(release.Release, &release.UserRelease))
	}

	return releaseDTOs, nil
}

func (s *Service) GetAlbumFormats(ctx context.Context, userID, albumID string) ([]AlbumFormatDTO, error) {
	allReleases, err := s.db.Queries().GetReleases(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get releases: %w", err)
	}

	userReleases, err := s.db.Queries().GetUserReleasesByAlbumId(ctx, sqlc.GetUserReleasesByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user releases: %w", err)
	}

	return buildAlbumFormats(allReleases, userReleases), nil
}

func (s *Service) SaveAlbumFormats(ctx context.Context, userID, albumID string, inputs []SaveFormatInput) error {
	return s.db.WithTx(func(tx *db.DB) error {
		allReleases, err := tx.Queries().GetReleases(ctx, albumID)
		if err != nil {
			return fmt.Errorf("failed to get releases: %w", err)
		}
		userReleases, err := tx.Queries().GetUserReleasesByAlbumId(ctx, sqlc.GetUserReleasesByAlbumIdParams{
			UserID:  userID,
			AlbumID: albumID,
		})
		if err != nil {
			return fmt.Errorf("failed to get user releases: %w", err)
		}

		currentFormats := buildAlbumFormats(allReleases, userReleases)
		currentByFormat := make(map[models.ReleaseFormat]AlbumFormatDTO, len(currentFormats))
		for _, f := range currentFormats {
			currentByFormat[f.Format] = f
		}

		for _, input := range inputs {
			if input.Format == models.ReleaseFormatDigital {
				continue // digital is managed by Spotify, never modified here
			}

			current := currentByFormat[input.Format]
			releaseID := current.ReleaseID

			if input.Owned {
				if !current.Owned {
					if releaseID == "" {
						r, err := tx.Queries().GetOrCreateRelease(ctx, sqlc.GetOrCreateReleaseParams{
							ID:      uuid.New().String(),
							AlbumID: albumID,
							Format:  input.Format,
						})
						if err != nil {
							return fmt.Errorf("failed to get/create release: %w", err)
						}
						releaseID = r.ID
					}
					_, err := tx.Queries().UpsertUserRelease(ctx, sqlc.UpsertUserReleaseParams{
						ID:        uuid.New().String(),
						UserID:    userID,
						ReleaseID: releaseID,
						AddedAt:   time.Now(),
					})
					if err != nil {
						return fmt.Errorf("failed to upsert user release: %w", err)
					}
				}

				if releaseID != "" && input.DiscogsID != "" {
					var releasedAt sql.NullTime
					if input.ReleasedAt != nil {
						releasedAt = sql.NullTime{Time: *input.ReleasedAt, Valid: true}
					}
					err := tx.Queries().UpdateReleaseDiscogsInfo(ctx, sqlc.UpdateReleaseDiscogsInfoParams{
						ID:         releaseID,
						DiscogsID:  sql.NullString{String: input.DiscogsID, Valid: true},
						Label:      sql.NullString{String: input.Label, Valid: input.Label != ""},
						ReleasedAt: releasedAt,
					})
					if err != nil {
						return fmt.Errorf("failed to update release discogs info: %w", err)
					}
				} else if releaseID != "" && input.DiscogsID == "" && current.DiscogsID != "" {
					err := tx.Queries().UpdateReleaseDiscogsInfo(ctx, sqlc.UpdateReleaseDiscogsInfoParams{
						ID: releaseID,
					})
					if err != nil {
						return fmt.Errorf("failed to clear release discogs info: %w", err)
					}
				}
			} else if current.Owned && releaseID != "" {
				err := tx.Queries().SoftDeleteUserRelease(ctx, sqlc.SoftDeleteUserReleaseParams{
					UserID:    userID,
					ReleaseID: releaseID,
				})
				if err != nil {
					return fmt.Errorf("failed to soft delete user release: %w", err)
				}
			}
		}
		return nil
	})
}
