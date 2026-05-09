package library

import (
	"sort"

	"github.com/alecdray/wax/src/internal/core/db/models"
)

const AlbumsPageSize = 20

type AlbumDTOs []AlbumDTO

func (albums AlbumDTOs) Page(offset int) AlbumDTOs {
	if offset >= len(albums) {
		return nil
	}
	end := offset + AlbumsPageSize
	if end > len(albums) {
		end = len(albums)
	}
	return albums[offset:end]
}

func (albums AlbumDTOs) SortByTitle(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		if ascending {
			return albums[i].Title < albums[j].Title
		}
		return albums[i].Title > albums[j].Title
	})
}

func (albums AlbumDTOs) SortByArtist(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		if len(albums[i].Artists) == 0 && len(albums[j].Artists) == 0 {
			return false
		}
		if len(albums[i].Artists) == 0 {
			return ascending
		}
		if len(albums[j].Artists) == 0 {
			return !ascending
		}
		if ascending {
			return albums[i].Artists[0].Name < albums[j].Artists[0].Name
		}
		return albums[i].Artists[0].Name > albums[j].Artists[0].Name
	})
}

func (albums AlbumDTOs) SortByRating(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		var ratingI, ratingJ *float64
		if albums[i].Rating != nil {
			ratingI = albums[i].Rating.Rating
		}
		if albums[j].Rating != nil {
			ratingJ = albums[j].Rating.Rating
		}
		if ratingI == nil && ratingJ == nil {
			return false
		}
		if ratingI == nil {
			return ascending
		}
		if ratingJ == nil {
			return !ascending
		}
		if ascending {
			return *ratingI < *ratingJ
		}
		return *ratingI > *ratingJ
	})
}

func (albums AlbumDTOs) SortByLastPlayed(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		if albums[i].LastPlayedAt == nil && albums[j].LastPlayedAt == nil {
			return false
		}
		if albums[i].LastPlayedAt == nil {
			return ascending
		}
		if albums[j].LastPlayedAt == nil {
			return !ascending
		}
		if ascending {
			return albums[i].LastPlayedAt.Before(*albums[j].LastPlayedAt)
		}
		return albums[i].LastPlayedAt.After(*albums[j].LastPlayedAt)
	})
}

func (albums AlbumDTOs) SortByDate(ascending bool) {
	sort.Slice(albums, func(i, j int) bool {
		dateI := albums[i].Releases.OldestAddedAtDate()
		dateJ := albums[j].Releases.OldestAddedAtDate()
		if dateI == nil && dateJ == nil {
			return false
		}
		if dateI == nil {
			return ascending
		}
		if dateJ == nil {
			return !ascending
		}
		if ascending {
			return dateI.Before(*dateJ)
		}
		return dateI.After(*dateJ)
	})
}

type FilterParams struct {
	MinRating *float64
	MaxRating *float64
	Rated     string // "only" | "unrated" | ""
	Formats   []models.ReleaseFormat
	ArtistIDs []string
}

func (albums AlbumDTOs) Filter(p FilterParams) AlbumDTOs {
	if p.MinRating == nil && p.MaxRating == nil && p.Rated == "" && len(p.Formats) == 0 && len(p.ArtistIDs) == 0 {
		return albums
	}
	result := make(AlbumDTOs, 0, len(albums))
	for _, album := range albums {
		if p.MinRating != nil {
			if album.Rating == nil || album.Rating.Rating == nil || *album.Rating.Rating < *p.MinRating {
				continue
			}
		}
		if p.MaxRating != nil {
			if album.Rating == nil || album.Rating.Rating == nil || *album.Rating.Rating > *p.MaxRating {
				continue
			}
		}
		switch p.Rated {
		case "only":
			if album.Rating == nil || album.Rating.Rating == nil {
				continue
			}
		case "unrated":
			if album.Rating != nil && album.Rating.Rating != nil {
				continue
			}
		}
		if len(p.Formats) > 0 {
			hasFormat := false
			for _, format := range p.Formats {
				if release := album.Releases.FindFormat(format); release != nil && release.AddedAt != nil {
					hasFormat = true
					break
				}
			}
			if !hasFormat {
				continue
			}
		}
		if len(p.ArtistIDs) > 0 {
			hasArtist := false
		outer:
			for _, artistID := range p.ArtistIDs {
				for _, artist := range album.Artists {
					if artist.ID == artistID {
						hasArtist = true
						break outer
					}
				}
			}
			if !hasArtist {
				continue
			}
		}
		result = append(result, album)
	}
	return result
}

// Library is the aggregate the library dashboard view binds to: a user's
// albums plus derived collections (unique artists and tracks across those albums).
type Library struct {
	OwnerUserID string
	Albums      AlbumDTOs
	Artists     []ArtistDTO
	Tracks      []TrackDTO
}

func NewLibrary(ownerUserID string, albums []AlbumDTO) *Library {
	d := &Library{
		OwnerUserID: ownerUserID,
		Albums:      albums,
	}

	d.Artists = d.artists()
	d.Tracks = d.tracks()

	return d
}

func (d *Library) artists() []ArtistDTO {
	artistsSet := make(map[string]ArtistDTO)
	for _, album := range d.Albums {
		for _, artist := range album.Artists {
			artistsSet[artist.ID] = artist
		}
	}

	artists := make([]ArtistDTO, 0, len(artistsSet))
	for _, artist := range artistsSet {
		artists = append(artists, artist)
	}

	return artists
}

func (d *Library) tracks() []TrackDTO {
	tracksSet := make(map[string]TrackDTO)
	for _, album := range d.Albums {
		for _, track := range album.Tracks {
			tracksSet[track.ID] = track
		}
	}

	tracks := make([]TrackDTO, 0, len(tracksSet))
	for _, track := range tracksSet {
		tracks = append(tracks, track)
	}

	return tracks
}
