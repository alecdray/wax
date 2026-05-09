package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/alecdray/wax/src/internal/core/utils"
	"github.com/alecdray/wax/src/internal/listeninghistory"
	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/spotify"
	"github.com/alecdray/wax/src/internal/tags"
	"time"

	"github.com/google/uuid"
)

type AlbumSummaryDTO struct {
	ID        string
	SpotifyID string
	Title     string
	Artists   string
	ImageURL  string
	InLibrary bool
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
	RatingLog    []*review.AlbumRatingDTO
	Tags         []tags.TagDTO
	SleeveNote   *notes.AlbumNoteDTO
	LastPlayedAt *time.Time
	RatingState  *review.RatingStateDTO
}

type RerateAlbumDTO struct {
	ID          string
	SpotifyID   string
	Title       string
	Artists     string
	ImageURL    string
	Rating      *float64
	RatingState review.RatingState
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

type Service struct {
	db                      *db.DB
	spotifyService          *spotify.Service
	listeningHistoryService *listeninghistory.Service
	tagsService             *tags.Service
	notesService            *notes.Service
	reviewService           *review.Service
}

func NewService(db *db.DB, spotifyService *spotify.Service, listeningHistoryService *listeninghistory.Service, tagsService *tags.Service, notesService *notes.Service, reviewService *review.Service) *Service {
	return &Service{
		db:                      db,
		spotifyService:          spotifyService,
		listeningHistoryService: listeningHistoryService,
		tagsService:             tagsService,
		notesService:            notesService,
		reviewService:           reviewService,
	}
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

	ratings, err := s.db.Queries().GetLatestUserAlbumRatings(ctx, sqlc.GetLatestUserAlbumRatingsParams{
		UserID:   userId,
		UserID_2: userId,
	})
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

	notesByAlbumId, err := s.notesService.GetAlbumNotesByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		err = fmt.Errorf("failed to get album notes: %w", err)
		return nil, err
	}

	ratingStates, err := s.reviewService.GetAllRatingStates(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating states: %w", err)
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
		dto.SleeveNote = notesByAlbumId[album.ID]
		dto.RatingState = ratingStates[album.ID]
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

	latestRating, err := s.db.Queries().GetLatestUserAlbumRating(ctx, sqlc.GetLatestUserAlbumRatingParams{
		UserID:  userId,
		AlbumID: album.ID,
	})
	var ratingDTO *review.AlbumRatingDTO
	if errors.Is(err, sql.ErrNoRows) {
		// no rating
	} else if err != nil {
		err = fmt.Errorf("failed to get rating: %w", err)
		return nil, err
	} else {
		ratingDTO = review.NewAlbumRatingDTOFromModel(latestRating)
	}

	ratingLogRows, err := s.db.Queries().GetUserAlbumRatingLog(ctx, sqlc.GetUserAlbumRatingLogParams{
		UserID:  userId,
		AlbumID: album.ID,
	})
	if err != nil {
		err = fmt.Errorf("failed to get rating log: %w", err)
		return nil, err
	}
	ratingLog := make([]*review.AlbumRatingDTO, len(ratingLogRows))
	for i, row := range ratingLogRows {
		ratingLog[i] = review.NewAlbumRatingDTOFromModel(row)
	}

	albumDto := NewAlbumDTOFromModel(
		album,
		artistDtos,
		trackDtos,
		releasesDtos,
		ratingDTO,
	)
	albumDto.RatingLog = ratingLog

	albumTags, err := s.tagsService.GetAlbumTags(ctx, userId, albumId)
	if err != nil {
		err = fmt.Errorf("failed to get album tags: %w", err)
		return nil, err
	}
	albumDto.Tags = albumTags

	sleeveNote, err := s.notesService.GetAlbumNote(ctx, userId, albumId)
	if err != nil {
		err = fmt.Errorf("failed to get album note: %w", err)
		return nil, err
	}
	albumDto.SleeveNote = sleeveNote

	ratingState, err := s.reviewService.GetRatingState(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating state: %w", err)
	}
	albumDto.RatingState = ratingState

	lastPlayedAtByAlbumId, err := s.listeningHistoryService.GetLastPlayedAtByAlbumIds(ctx, userId, []string{albumId})
	if err != nil {
		err = fmt.Errorf("failed to get last played at: %w", err)
		return nil, err
	}
	if t, ok := lastPlayedAtByAlbumId[albumId]; ok {
		albumDto.LastPlayedAt = &t
	}

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
			InLibrary: row.InLibrary != 0,
		})
	}
	return dtos, nil
}

func (s *Service) RemoveAlbumFromLibrary(ctx contextx.ContextX, userId, albumId string) error {
	album, err := s.db.Queries().GetAlbum(ctx, albumId)
	if err != nil {
		return fmt.Errorf("failed to get album: %w", err)
	}

	if err := s.db.Queries().SoftDeleteUserReleasesByAlbumId(ctx, sqlc.SoftDeleteUserReleasesByAlbumIdParams{
		UserID:  userId,
		AlbumID: albumId,
	}); err != nil {
		return fmt.Errorf("failed to soft delete releases: %w", err)
	}

	if err := s.spotifyService.RemoveAlbumFromSavedLibrary(ctx, userId, album.SpotifyID); err != nil {
		slog.WarnContext(ctx, "failed to remove album from spotify saved library", "error", err)
	}

	return nil
}

func (s *Service) GetRerateQueue(ctx context.Context, userID string) ([]RerateAlbumDTO, error) {
	rows, err := s.db.Queries().GetRerateQueueAlbums(ctx, sqlc.GetRerateQueueAlbumsParams{
		UserID:   userID,
		UserID_2: userID,
		UserID_3: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get rerate queue: %w", err)
	}
	dtos := make([]RerateAlbumDTO, 0, len(rows))
	for _, row := range rows {
		dto := RerateAlbumDTO{
			ID:          row.ID,
			SpotifyID:   row.SpotifyID,
			Title:       row.Title,
			RatingState: review.RatingState(row.State),
		}
		if row.ImageUrl.Valid {
			dto.ImageURL = row.ImageUrl.String
		}
		if names, ok := row.ArtistNames.(string); ok {
			dto.Artists = names
		}
		// Rating is 0 for albums with no rating (LEFT JOIN returns 0 for NULLs)
		// Only set it if there's actually a rating state with existing data
		if row.Rating != 0 {
			r := row.Rating
			dto.Rating = &r
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

func (s *Service) GetUnratedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	rows, err := s.db.Queries().GetUnratedAlbums(ctx, sqlc.GetUnratedAlbumsParams{
		UserID:   userID,
		UserID_2: userID,
		UserID_3: userID,
	})
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
			InLibrary: true,
		})
	}
	return dtos, nil
}
