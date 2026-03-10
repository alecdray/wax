package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/db/models"
	"shmoopicks/src/internal/core/db/sqlc"
	"shmoopicks/src/internal/core/utils"
	"shmoopicks/src/internal/listeninghistory"
	"shmoopicks/src/internal/review"
	"shmoopicks/src/internal/tags"
	"sort"
	"time"

	"github.com/google/uuid"
)

type AlbumSummaryDTO struct {
	ID        string
	SpotifyID string
	Title     string
	Artists   string
	ImageURL  string
}

type ReleaseDTO struct {
	ID      string
	AlbumID string
	Format  models.ReleaseFormat
	AddedAt *time.Time
}

func NewReleaseDTOFromModel(model sqlc.Release, userRelease *sqlc.UserRelease) ReleaseDTO {
	dto := ReleaseDTO{
		ID:      model.ID,
		AlbumID: model.AlbumID,
		Format:  model.Format,
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

type TrackDTO struct {
	ID        string
	SpotifyID string
	Title     string
}

func NewTrackDTOFromModel(model sqlc.Track) TrackDTO {
	dto := TrackDTO{
		ID:        model.ID,
		SpotifyID: model.SpotifyID,
		Title:     model.Title,
	}

	return dto
}

type ArtistDTO struct {
	ID        string
	SpotifyID string
	Name      string
}

func NewArtistDTOFromModel(model sqlc.Artist) ArtistDTO {
	dto := ArtistDTO{
		ID:        model.ID,
		SpotifyID: model.SpotifyID,
		Name:      model.Name,
	}

	return dto
}

type AlbumDTO struct {
	ID           string
	SpotifyID    string
	Title        string
	ImageURL     string
	Artists      []ArtistDTO
	Tracks       []TrackDTO
	Releases     ReleaseDTOs
	Rating       *review.AlbumRatingDTO
	Tags         []tags.TagDTO
	LastPlayedAt *time.Time
}

func NewAlbumDTOFromModel(model sqlc.Album, artists []ArtistDTO, tracks []TrackDTO, releases []ReleaseDTO, rating *review.AlbumRatingDTO) AlbumDTO {
	return AlbumDTO{
		ID:        model.ID,
		SpotifyID: model.SpotifyID,
		Title:     model.Title,
		ImageURL:  model.ImageUrl.String,
		Artists:   artists,
		Tracks:    tracks,
		Releases:  releases,
		Rating:    rating,
	}
}

type AlbumDTOs []AlbumDTO

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

type Library struct {
	OwnerUserID string
	Albums      AlbumDTOs
	Artists     []ArtistDTO
	Tracks      []TrackDTO
}

func NewLibrary(ownerUserID string, albums []AlbumDTO) *Library {
	lib := &Library{
		OwnerUserID: ownerUserID,
		Albums:      albums,
	}

	lib.Artists = lib.artists()
	lib.Tracks = lib.tracks()

	return lib
}

func (l *Library) artists() []ArtistDTO {
	artistsSet := make(map[string]ArtistDTO)
	for _, album := range l.Albums {
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

func (l *Library) tracks() []TrackDTO {
	tracksSet := make(map[string]TrackDTO)
	for _, album := range l.Albums {
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

type Service struct {
	db                      *db.DB
	listeningHistoryService *listeninghistory.Service
	tagsService             *tags.Service
}

func NewService(db *db.DB, listeningHistoryService *listeninghistory.Service, tagsService *tags.Service) *Service {
	return &Service{
		db:                      db,
		listeningHistoryService: listeningHistoryService,
		tagsService:             tagsService,
	}
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

func (s *Service) GetAlbumsInLibrary(ctx context.Context, userId string) ([]AlbumDTO, error) {
	releases, err := s.GetReleasesInLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get releases: %w", err)
	}

	releasesByAlbumId := make(map[string][]ReleaseDTO, len(releases))
	albumIds := make([]string, 0, len(releases))
	for _, release := range releases {
		albumIds = append(albumIds, release.AlbumID)
		releasesByAlbumId[release.AlbumID] = append(releasesByAlbumId[release.AlbumID], release)

	}

	albums, err := s.db.Queries().GetAlbumsByIDs(ctx, albumIds)
	if err != nil {
		err = fmt.Errorf("failed to get albums: %w", err)
		return nil, err
	}

	artists, err := s.db.Queries().GetAlbumArtistsByAlbumIds(ctx, albumIds)
	if err != nil {
		err = fmt.Errorf("failed to get album artists: %w", err)
		return nil, err
	}

	artistsByAlbumId := make(map[string][]ArtistDTO, len(albumIds))
	for _, artist := range artists {
		artistsByAlbumId[artist.AlbumID] = append(artistsByAlbumId[artist.AlbumID], NewArtistDTOFromModel(artist.Artist))
	}

	tracks, err := s.db.Queries().GetAlbumTracksByAlbumIds(ctx, albumIds)
	if err != nil {
		err = fmt.Errorf("failed to get album tracks: %w", err)
		return nil, err
	}

	tracksByAlbumId := make(map[string][]TrackDTO, len(albumIds))
	for _, track := range tracks {
		tracksByAlbumId[track.AlbumID] = append(tracksByAlbumId[track.AlbumID], NewTrackDTOFromModel(track.Track))
	}

	ratings, err := s.db.Queries().GetUserAlbumRatings(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get ratings: %w", err)
		return nil, err
	}

	ratingsByAlbumId := make(map[string]review.AlbumRatingDTO, len(ratings))
	for _, rating := range ratings {
		ratingsByAlbumId[rating.AlbumID] = *review.NewAlbumRatingDTOFromModel(rating)
	}

	lastPlayedAtByAlbumId, err := s.listeningHistoryService.GetLastPlayedAtByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		err = fmt.Errorf("failed to get last played at: %w", err)
		return nil, err
	}

	tagsByAlbumId, err := s.tagsService.GetAlbumTagsByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		err = fmt.Errorf("failed to get album tags: %w", err)
		return nil, err
	}

	var albumDTOs []AlbumDTO
	for _, album := range albums {
		dto := NewAlbumDTOFromModel(
			album,
			artistsByAlbumId[album.ID],
			tracksByAlbumId[album.ID],
			releasesByAlbumId[album.ID],
			utils.NewPointer(ratingsByAlbumId[album.ID]),
		)
		if t, ok := lastPlayedAtByAlbumId[album.ID]; ok {
			dto.LastPlayedAt = &t
		}
		dto.Tags = tagsByAlbumId[album.ID]
		albumDTOs = append(albumDTOs, dto)
	}

	return albumDTOs, nil
}

func (s *Service) GetLibrary(ctx context.Context, userId string) (*Library, error) {
	albums, err := s.GetAlbumsInLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get user albums: %w", err)
		return nil, err
	}

	return NewLibrary(userId, albums), nil
}

func (s *Service) AddAlbumsToLibrary(ctx context.Context, userId string, albums []AlbumDTO) error {
	err := s.db.WithTx(func(tx *db.DB) error {
		for _, album := range albums {
			// insert album
			albumModel, err := tx.Queries().GetOrCreateAlbum(ctx, sqlc.GetOrCreateAlbumParams{
				ID:        album.ID,
				SpotifyID: album.SpotifyID,
				Title:     album.Title,
				ImageUrl:  sql.NullString{String: album.ImageURL, Valid: album.ImageURL != ""},
			})
			if err != nil {
				err = fmt.Errorf("failed to get/create album: %w", err)
				return err
			}
			album = NewAlbumDTOFromModel(albumModel, album.Artists, album.Tracks, album.Releases, album.Rating)

			for i, track := range album.Tracks {
				// insert tracks
				trackModel, err := tx.Queries().GetOrCreateTrack(ctx, sqlc.GetOrCreateTrackParams{
					ID:        track.ID,
					SpotifyID: track.SpotifyID,
					Title:     track.Title,
				})
				if err != nil {
					err = fmt.Errorf("failed to get/create track: %w", err)
					return err
				}

				// insert album_tracks
				_, err = tx.Queries().GetOrCreateAlbumTrack(ctx, sqlc.GetOrCreateAlbumTrackParams{
					AlbumID: albumModel.ID,
					TrackID: trackModel.ID,
				})
				if err != nil {
					err = fmt.Errorf("failed to get/create album track: %w", err)
					return err
				}

				album.Tracks[i] = NewTrackDTOFromModel(trackModel)
			}

			for i, artist := range album.Artists {
				// insert artsits
				artistModel, err := tx.Queries().GetOrCreateArtist(ctx, sqlc.GetOrCreateArtistParams{
					ID:        artist.ID,
					SpotifyID: artist.SpotifyID,
					Name:      artist.Name,
				})
				if err != nil {
					err = fmt.Errorf("failed to get/create artist: %w", err)
					return err
				}

				// insert album_artists
				_, err = tx.Queries().GetOrCreateAlbumArtist(ctx, sqlc.GetOrCreateAlbumArtistParams{
					AlbumID:  albumModel.ID,
					ArtistID: artistModel.ID,
				})
				if err != nil {
					err = fmt.Errorf("failed to get/create album artist: %w", err)
					return err
				}

				album.Artists[i] = NewArtistDTOFromModel(artistModel)
			}

			for i, release := range album.Releases {
				// insert releases
				releaseModel, err := tx.Queries().GetOrCreateRelease(ctx, sqlc.GetOrCreateReleaseParams{
					ID:      release.ID,
					AlbumID: album.ID,
					Format:  release.Format,
				})
				if err != nil {
					err = fmt.Errorf("failed to get/create release: %w", err)
					return err
				}

				// insert user_releases
				userRelease, err := tx.Queries().UpsertUserRelease(ctx, sqlc.UpsertUserReleaseParams{
					ID:        uuid.New().String(),
					UserID:    userId,
					ReleaseID: releaseModel.ID,
					AddedAt:   *release.AddedAt,
				})
				if err != nil {
					err = fmt.Errorf("failed to get/create user release: %w", err)
					return err
				}

				album.Releases[i] = NewReleaseDTOFromModel(releaseModel, &userRelease)
			}
		}

		return nil
	})

	return err
}

func (s *Service) GetAlbumInLibrary(ctx context.Context, userId string, albumId string) (*AlbumDTO, error) {
	album, err := s.db.Queries().GetAlbum(ctx, albumId)
	if err != nil {
		err = fmt.Errorf("failed to get albums: %w", err)
		return nil, err
	}

	releases, err := s.db.Queries().GetUserReleasesByAlbumId(ctx, sqlc.GetUserReleasesByAlbumIdParams{
		UserID:  userId,
		AlbumID: albumId,
	})
	if err != nil {
		err = fmt.Errorf("failed to get releases: %w", err)
		return nil, err
	}

	if len(releases) < 1 {
		return nil, errors.New("album not in library")
	}

	releasesDtos := make([]ReleaseDTO, len(releases))
	for i, release := range releases {
		releasesDtos[i] = NewReleaseDTOFromModel(release.Release, &release.UserRelease)
	}

	artists, err := s.db.Queries().GetAlbumArtistByAlbumId(ctx, album.ID)
	if err != nil {
		err = fmt.Errorf("failed to get album artists: %w", err)
		return nil, err
	}

	artistDtos := make([]ArtistDTO, len(artists))
	for i, artist := range artists {
		artistDtos[i] = NewArtistDTOFromModel(artist.Artist)
	}

	tracks, err := s.db.Queries().GetAlbumTracksByAlbumId(ctx, album.ID)
	if err != nil {
		err = fmt.Errorf("failed to get album tracks: %w", err)
		return nil, err
	}

	trackDtos := make([]TrackDTO, len(tracks))
	for i, track := range tracks {
		trackDtos[i] = NewTrackDTOFromModel(track.Track)
	}

	rating, err := s.db.Queries().GetUserAlbumRating(ctx, sqlc.GetUserAlbumRatingParams{
		UserID:  userId,
		AlbumID: album.ID,
	})
	if errors.Is(sql.ErrNoRows, err) {
		// pass
	} else if err != nil {
		err = fmt.Errorf("failed to get ratings: %w", err)
		return nil, err
	}

	albumDto := NewAlbumDTOFromModel(
		album,
		artistDtos,
		trackDtos,
		releasesDtos,
		review.NewAlbumRatingDTOFromModel(rating),
	)

	albumTags, err := s.tagsService.GetAlbumTags(ctx, userId, albumId)
	if err != nil {
		err = fmt.Errorf("failed to get album tags: %w", err)
		return nil, err
	}
	albumDto.Tags = albumTags

	return &albumDto, nil
}

func (s *Service) GetRecentlyPlayedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	rows, err := s.db.Queries().GetRecentlyPlayedAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently played albums: %w", err)
	}

	dtos := make([]AlbumSummaryDTO, 0, len(rows))
	for _, row := range rows {
		dtos = append(dtos, AlbumSummaryDTO{
			ID:        row.ID,
			SpotifyID: row.SpotifyID,
			Title:     row.Title,
			Artists:   fmt.Sprintf("%s", row.ArtistNames),
			ImageURL:  row.ImageUrl.String,
		})
	}
	return dtos, nil
}

func (s *Service) GetUnratedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	rows, err := s.db.Queries().GetUnratedAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unrated albums: %w", err)
	}

	dtos := make([]AlbumSummaryDTO, 0, len(rows))
	for _, row := range rows {
		dtos = append(dtos, AlbumSummaryDTO{
			ID:        row.ID,
			SpotifyID: row.SpotifyID,
			Title:     row.Title,
			Artists:   fmt.Sprintf("%s", row.ArtistNames),
			ImageURL:  row.ImageUrl.String,
		})
	}
	return dtos, nil
}
