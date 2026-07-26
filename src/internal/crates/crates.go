package crates

import (
	"time"

	"github.com/alecdray/wax/src/internal/library"
)

type CrateDTO struct {
	ID         string
	Name       string
	AlbumCount int
	CreatedAt  time.Time
}

type CrateDetailDTO struct {
	CrateDTO
	Albums []library.AlbumDTO
}
