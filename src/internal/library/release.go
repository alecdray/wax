package library

import (
	"strconv"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/discogs"
)

type ReleaseDTO struct {
	ID         string
	AlbumID    string
	Format     models.ReleaseFormat
	AddedAt    *time.Time
	DiscogsID  string
	Label      string
	ReleasedAt *time.Time
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

type SaveFormatInput struct {
	Format     models.ReleaseFormat
	Owned      bool
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
// fallback values. Discogs's Release.Released can be "YYYY-MM-DD", "YYYY", or empty; this
// function picks the most precise option and pads bare years to YYYY-01-01. release may be nil.
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
