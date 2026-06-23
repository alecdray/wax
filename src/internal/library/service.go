package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/utils"
	"github.com/alecdray/wax/src/internal/listeninghistory"
	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/spotify"
	"github.com/alecdray/wax/src/internal/tags"
	"github.com/google/uuid"
	spotifylib "github.com/zmb3/spotify/v2"
)

type Service struct {
	db                      *db.DB
	repo                    *Repo
	spotifyService          *spotify.Service
	listeningHistoryService *listeninghistory.Service
	tagsService             *tags.Service
	notesService            *notes.Service
	reviewService           *review.Service
}

func NewService(d *db.DB, spotifyService *spotify.Service, listeningHistoryService *listeninghistory.Service, tagsService *tags.Service, notesService *notes.Service, reviewService *review.Service) *Service {
	return &Service{
		db:                      d,
		repo:                    NewRepo(d.Queries()),
		spotifyService:          spotifyService,
		listeningHistoryService: listeningHistoryService,
		tagsService:             tagsService,
		notesService:            notesService,
		reviewService:           reviewService,
	}
}

func (s *Service) GetReleasesInLibrary(ctx context.Context, userId string) ([]ReleaseDTO, error) {
	releaseDTOs, err := s.repo.GetUserReleases(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user releases: %w", err)
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

	albums, err := s.repo.GetAlbumsByIDs(ctx, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums: %w", err)
	}

	artistsByAlbumId, err := s.repo.GetArtistsByAlbumIDs(ctx, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album artists: %w", err)
	}

	tracksByAlbumId, err := s.repo.GetTracksByAlbumIDs(ctx, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tracks: %w", err)
	}

	ratingsByAlbumId, err := s.reviewService.GetLatestRatings(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings: %w", err)
	}

	lastPlayedAtByAlbumId, err := s.listeningHistoryService.GetLastPlayedAtByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get last played at: %w", err)
	}

	tagsByAlbumId, err := s.tagsService.GetAlbumTagsByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}

	notesByAlbumId, err := s.notesService.GetAlbumNotesByAlbumIds(ctx, userId, albumIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get album notes: %w", err)
	}

	ratingStates, err := s.reviewService.GetAllRatingStates(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating states: %w", err)
	}

	var albumDTOs []AlbumDTO
	for _, album := range albums {
		dto := album
		dto.Artists = artistsByAlbumId[album.ID]
		dto.Tracks = tracksByAlbumId[album.ID]
		dto.Releases = releasesByAlbumId[album.ID]
		rating := ratingsByAlbumId[album.ID]
		dto.Rating = utils.NewPointer(rating)
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
		return nil, fmt.Errorf("failed to get user albums: %w", err)
	}
	return NewLibrary(userId, albums), nil
}

func (s *Service) AddAlbumsToLibrary(ctx context.Context, userId string, albums []AlbumDTO) error {
	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		for _, album := range albums {
			if _, err := txRepo.AddAlbumToCollection(ctx, userId, album); err != nil {
				return fmt.Errorf("failed to add album to collection: %w", err)
			}
			if err := txRepo.RemoveAlbumFromRadar(ctx, userId, album.ID); err != nil {
				return fmt.Errorf("failed to clear radar: %w", err)
			}
		}
		return nil
	})
}

func (s *Service) GetAlbumInLibrary(ctx context.Context, userId string, albumId string) (*AlbumDTO, error) {
	album, err := s.repo.GetAlbumByID(ctx, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums: %w", err)
	}

	releasesDtos, err := s.repo.GetUserReleasesByAlbumID(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get releases: %w", err)
	}

	if len(releasesDtos) < 1 {
		return nil, errors.New("album not in library")
	}

	artistDtos, err := s.repo.GetArtistsByAlbumID(ctx, album.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album artists: %w", err)
	}

	trackDtos, err := s.repo.GetTracksByAlbumID(ctx, album.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tracks: %w", err)
	}

	ratingDTO, err := s.reviewService.GetLatestRating(ctx, userId, album.ID)
	if errors.Is(err, sql.ErrNoRows) {
		ratingDTO = nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	ratingLog, err := s.reviewService.GetRatingLog(ctx, userId, album.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating log: %w", err)
	}

	album.Artists = artistDtos
	album.Tracks = trackDtos
	album.Releases = releasesDtos
	album.Rating = ratingDTO
	album.RatingLog = ratingLog

	albumTags, err := s.tagsService.GetAlbumTags(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tags: %w", err)
	}
	album.Tags = albumTags

	sleeveNote, err := s.notesService.GetAlbumNote(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get album note: %w", err)
	}
	album.SleeveNote = sleeveNote

	ratingState, err := s.reviewService.GetRatingState(ctx, userId, albumId)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating state: %w", err)
	}
	album.RatingState = ratingState

	lastPlayedAtByAlbumId, err := s.listeningHistoryService.GetLastPlayedAtByAlbumIds(ctx, userId, []string{albumId})
	if err != nil {
		return nil, fmt.Errorf("failed to get last played at: %w", err)
	}
	if t, ok := lastPlayedAtByAlbumId[albumId]; ok {
		album.LastPlayedAt = &t
	}

	return album, nil
}

func (s *Service) GetRecentlyPlayedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	dtos, err := s.repo.GetRecentlyPlayedAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently played albums: %w", err)
	}
	return dtos, nil
}

func (s *Service) RemoveAlbumFromLibrary(ctx contextx.ContextX, userId, albumId string) error {
	spotifyID, err := s.repo.GetAlbumSpotifyID(ctx, albumId)
	if err != nil {
		return fmt.Errorf("failed to get album: %w", err)
	}

	if err := s.repo.MarkReleasesRemovedByAlbumID(ctx, userId, albumId); err != nil {
		return fmt.Errorf("failed to mark releases removed: %w", err)
	}

	if err := s.spotifyService.RemoveAlbumFromSavedLibrary(ctx, userId, spotifyID); err != nil {
		slog.WarnContext(ctx, "failed to remove album from spotify saved library", "error", err)
	}

	return nil
}

func (s *Service) GetProvisionalAlbums(ctx context.Context, userID string) ([]ProvisionalAlbumDTO, error) {
	dtos, err := s.repo.GetProvisionalAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provisional albums: %w", err)
	}
	return dtos, nil
}

func (s *Service) GetAlbumFormats(ctx context.Context, userID, albumID string) ([]AlbumFormatDTO, error) {
	formats, err := s.repo.GetAlbumFormats(ctx, userID, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album formats: %w", err)
	}
	return formats, nil
}

func (s *Service) SaveAlbumFormats(ctx context.Context, userID, albumID string, inputs []SaveFormatInput) error {
	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())

		currentFormats, err := txRepo.GetAlbumFormats(ctx, userID, albumID)
		if err != nil {
			return fmt.Errorf("failed to get current formats: %w", err)
		}
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
					newReleaseID, err := txRepo.AddOwnedRelease(ctx, userID, albumID, input.Format, releaseID, time.Now())
					if err != nil {
						return fmt.Errorf("failed to add owned release: %w", err)
					}
					releaseID = newReleaseID
					if err := txRepo.RemoveAlbumFromRadar(ctx, userID, albumID); err != nil {
						return fmt.Errorf("failed to clear radar: %w", err)
					}
				}

				if releaseID != "" && input.DiscogsID != "" {
					if err := txRepo.UpdateReleaseDiscogsInfo(ctx, releaseID, input.DiscogsID, input.Label, input.ReleasedAt); err != nil {
						return fmt.Errorf("failed to update release discogs info: %w", err)
					}
				} else if releaseID != "" && input.DiscogsID == "" && current.DiscogsID != "" {
					if err := txRepo.ClearReleaseDiscogsInfo(ctx, releaseID); err != nil {
						return fmt.Errorf("failed to clear release discogs info: %w", err)
					}
				}
			} else if current.Owned && releaseID != "" {
				if err := txRepo.MarkReleaseRemoved(ctx, userID, releaseID); err != nil {
					return fmt.Errorf("failed to mark release removed: %w", err)
				}
			}
		}
		return nil
	})
}

func (s *Service) GetUnratedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	dtos, err := s.repo.GetUnratedAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unrated albums: %w", err)
	}
	return dtos, nil
}

// AcquireFromWishlist transitions a user's wishlist release to 'owned' and clears
// the album from the user's radar in a single tx.
func (s *Service) AcquireFromWishlist(ctx context.Context, userID, albumID, releaseID string) error {
	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		if err := txRepo.MarkReleaseOwned(ctx, userID, releaseID); err != nil {
			return fmt.Errorf("failed to acquire from wishlist: %w", err)
		}
		if err := txRepo.RemoveAlbumFromRadar(ctx, userID, albumID); err != nil {
			return fmt.Errorf("failed to clear radar: %w", err)
		}
		return nil
	})
}

// AddReleaseToWishlist upserts a wishlist release for the user and clears the
// album from the user's radar in a single tx. If releaseID is empty, a new
// release row is created for (album, format).
func (s *Service) AddReleaseToWishlist(ctx context.Context, userID, albumID string, format models.ReleaseFormat, releaseID string) (string, error) {
	var resultID string
	err := s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		newID, err := txRepo.AddReleaseToWishlist(ctx, userID, albumID, format, releaseID)
		if err != nil {
			return fmt.Errorf("failed to add release to wishlist: %w", err)
		}
		if err := txRepo.RemoveAlbumFromRadar(ctx, userID, albumID); err != nil {
			return fmt.Errorf("failed to clear radar: %w", err)
		}
		resultID = newID
		return nil
	})
	return resultID, err
}

// ErrAlbumAlreadyDecided is returned by AddAlbumToRadar (and AddSpotifyAlbumToRadar)
// when the user currently owns or wishlists the album, i.e. it is in the library.
// A `removed` album does not trigger this — it is radar-eligible (ADR 0005). The
// radar-inbox sync relies on this distinction: ErrAlbumAlreadyDecided means
// "already handled, drop the track", while other errors leave the track to retry.
var ErrAlbumAlreadyDecided = errors.New("album is already owned or wishlisted; cannot add to radar")

// AddAlbumToRadar adds an album to the user's radar (pre-decision queue).
// Refuses with ErrAlbumAlreadyDecided if the album is already owned or wishlisted.
func (s *Service) AddAlbumToRadar(ctx context.Context, userID, albumID string) error {
	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		hasRelease, err := txRepo.HasOwnedOrWishlistedReleaseForAlbum(ctx, userID, albumID)
		if err != nil {
			return fmt.Errorf("failed to check user releases: %w", err)
		}
		if hasRelease {
			return ErrAlbumAlreadyDecided
		}
		if err := txRepo.AddAlbumToRadar(ctx, userID, albumID); err != nil {
			return fmt.Errorf("failed to add album to radar: %w", err)
		}
		return nil
	})
}

// GetRadarAlbums returns the caller's radar entries as fully-populated
// AlbumDTOs (artists set; tracks/releases left empty — radar entries have no
// release rows). Used by the radar page's grid.
func (s *Service) GetRadarAlbums(ctx context.Context, userID string) ([]AlbumDTO, error) {
	_, albums, err := s.repo.GetRadarAlbums(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get radar albums: %w", err)
	}
	if len(albums) == 0 {
		return nil, nil
	}
	albumIDs := make([]string, len(albums))
	for i, a := range albums {
		albumIDs[i] = a.ID
	}
	artistsByAlbumID, err := s.repo.GetArtistsByAlbumIDs(ctx, albumIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get artists for radar albums: %w", err)
	}
	for i := range albums {
		albums[i].Artists = artistsByAlbumID[albums[i].ID]
	}
	return albums, nil
}

// RemoveAlbumFromRadar deletes the radar row. No-op if the user has no radar
// row for the album.
func (s *Service) RemoveAlbumFromRadar(ctx context.Context, userID, albumID string) error {
	return s.repo.RemoveAlbumFromRadar(ctx, userID, albumID)
}

// PromoteRadarToLibrary transitions a radar album to an owned digital release
// and pushes the album to the user's Spotify saved library. Spotify push is
// best-effort; a failure is logged but does not roll back the local DB
// (mirrors RemoveAlbumFromLibrary).
func (s *Service) PromoteRadarToLibrary(ctx contextx.ContextX, userID, albumID string) error {
	spotifyID, err := s.repo.GetAlbumSpotifyID(ctx, albumID)
	if err != nil {
		return fmt.Errorf("failed to get album spotify id: %w", err)
	}

	err = s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		if _, err := txRepo.AddOwnedRelease(ctx, userID, albumID, models.ReleaseFormatDigital, "", time.Now()); err != nil {
			return fmt.Errorf("failed to add owned digital release: %w", err)
		}
		if err := txRepo.RemoveAlbumFromRadar(ctx, userID, albumID); err != nil {
			return fmt.Errorf("failed to clear radar: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := s.spotifyService.AddAlbumToSavedLibrary(ctx, userID, spotifyID); err != nil {
		slog.WarnContext(ctx, "failed to push album to spotify saved library after radar promotion", "error", err, "album_id", albumID, "spotify_id", spotifyID)
	}
	return nil
}

// spotifyAlbumToDTO converts a Spotify FullAlbum into an AlbumDTO ready for
// EnsureAlbumWithMetadata. Releases is intentionally omitted — radar entries
// are pre-decision; the feed sync that populates Releases lives in
// feed/service.go and writes user_releases rows that radar must not create.
func spotifyAlbumToDTO(album *spotifylib.FullAlbum) AlbumDTO {
	var imageURL string
	if len(album.Images) > 0 {
		imageURL = album.Images[0].URL
	}
	dto := AlbumDTO{
		ID:        uuid.NewString(),
		SpotifyID: album.ID.String(),
		Title:     album.Name,
		ImageURL:  imageURL,
	}
	dto.Artists = make([]ArtistDTO, len(album.Artists))
	for i, a := range album.Artists {
		dto.Artists[i] = ArtistDTO{
			ID:        uuid.NewString(),
			SpotifyID: a.ID.String(),
			Name:      a.Name,
		}
	}
	dto.Tracks = make([]TrackDTO, 0, len(album.Tracks.Tracks))
	for _, t := range album.Tracks.Tracks {
		dto.Tracks = append(dto.Tracks, TrackDTO{
			ID:        uuid.NewString(),
			SpotifyID: t.ID.String(),
			Title:     t.Name,
		})
	}
	return dto
}

// mergeDiscoverState combines a slice of Spotify search results with the
// caller's per-album state lookup, producing one DiscoverResultDTO per
// Spotify result.
func mergeDiscoverState(results []spotifylib.SimpleAlbum, states map[string]UserAlbumStateRow) []DiscoverResultDTO {
	out := make([]DiscoverResultDTO, len(results))
	for i, a := range results {
		var imageURL string
		if len(a.Images) > 0 {
			imageURL = a.Images[0].URL
		}
		artists := make([]ArtistDTO, len(a.Artists))
		for j, ar := range a.Artists {
			artists[j] = ArtistDTO{
				SpotifyID: ar.ID.String(),
				Name:      ar.Name,
			}
		}
		dto := DiscoverResultDTO{
			SpotifyID: a.ID.String(),
			Title:     a.Name,
			Artists:   artists,
			ImageURL:  imageURL,
			State:     DiscoverAlbumStateNone,
		}
		if row, ok := states[a.ID.String()]; ok {
			dto.State = row.State
			dto.AlbumID = row.AlbumID
		}
		out[i] = dto
	}
	return out
}

// SearchAlbumsForDiscover queries Spotify and enriches each result with the
// caller's wax state (in_library, on_radar, removed, or none). Returns an
// empty slice (not nil) when the query is empty or yields no hits.
func (s *Service) SearchAlbumsForDiscover(ctx contextx.ContextX, userID, query string, limit int) ([]DiscoverResultDTO, error) {
	results, err := s.spotifyService.SearchAlbums(ctx, userID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("spotify search failed: %w", err)
	}
	if len(results) == 0 {
		return []DiscoverResultDTO{}, nil
	}
	spotifyIDs := make([]string, len(results))
	for i, a := range results {
		spotifyIDs[i] = a.ID.String()
	}
	states, err := s.repo.GetUserAlbumStateBySpotifyIDs(ctx, userID, spotifyIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to look up album states: %w", err)
	}
	return mergeDiscoverState(results, states), nil
}

// LookupDiscoverState exposes the per-Spotify-ID state lookup for adapters
// that need it after a write (e.g., to re-render a row in its new state).
func (s *Service) LookupDiscoverState(ctx contextx.ContextX, userID string, spotifyIDs []string) (map[string]UserAlbumStateRow, error) {
	return s.repo.GetUserAlbumStateBySpotifyIDs(ctx, userID, spotifyIDs)
}

// GetAlbumSpotifyID returns the Spotify ID for a wax album. Thin wrapper over
// the repo method (already used internally by RemoveAlbumFromLibrary); now
// exposed so adapters can re-render search-result rows after a radar delete.
func (s *Service) GetAlbumSpotifyID(ctx contextx.ContextX, albumID string) (string, error) {
	return s.repo.GetAlbumSpotifyID(ctx, albumID)
}

// GetAlbumActionsResult resolves a Spotify ID into a DiscoverResultDTO with
// state, AlbumID (when known to wax), title, image, and artists. Used to
// render the album-actions modal opened from any surface showing an
// out-of-library album.
//
// If the album exists in wax (any state), metadata is read from the local DB.
// Otherwise (state=none), it falls back to fetching from Spotify.
func (s *Service) GetAlbumActionsResult(ctx contextx.ContextX, userID, spotifyID string) (DiscoverResultDTO, error) {
	states, err := s.repo.GetUserAlbumStateBySpotifyIDs(ctx, userID, []string{spotifyID})
	if err != nil {
		return DiscoverResultDTO{}, fmt.Errorf("failed to look up album state: %w", err)
	}
	if state, ok := states[spotifyID]; ok {
		album, err := s.repo.GetAlbumByID(ctx, state.AlbumID)
		if err != nil {
			return DiscoverResultDTO{}, fmt.Errorf("failed to get album: %w", err)
		}
		artists, err := s.repo.GetArtistsByAlbumID(ctx, state.AlbumID)
		if err != nil {
			return DiscoverResultDTO{}, fmt.Errorf("failed to get artists: %w", err)
		}
		return DiscoverResultDTO{
			SpotifyID: spotifyID,
			AlbumID:   state.AlbumID,
			Title:     album.Title,
			ImageURL:  album.ImageURL,
			Artists:   artists,
			State:     state.State,
		}, nil
	}
	// Album not yet in wax — fetch metadata from Spotify.
	full, err := s.spotifyService.GetFullAlbum(ctx, userID, spotifyID)
	if err != nil {
		return DiscoverResultDTO{}, fmt.Errorf("failed to fetch spotify album: %w", err)
	}
	var imageURL string
	if len(full.Images) > 0 {
		imageURL = full.Images[0].URL
	}
	artists := make([]ArtistDTO, len(full.Artists))
	for i, a := range full.Artists {
		artists[i] = ArtistDTO{
			SpotifyID: a.ID.String(),
			Name:      a.Name,
		}
	}
	return DiscoverResultDTO{
		SpotifyID: spotifyID,
		Title:     full.Name,
		ImageURL:  imageURL,
		Artists:   artists,
		State:     DiscoverAlbumStateNone,
	}, nil
}

// AddSpotifyAlbumToRadar imports a Spotify album's metadata (album, artists,
// tracks) into wax and adds the album to the user's radar. Refuses with
// ErrAlbumAlreadyDecided if the album already has any user_releases row.
func (s *Service) AddSpotifyAlbumToRadar(ctx contextx.ContextX, userID, spotifyID string) error {
	spotifyAlbum, err := s.spotifyService.GetFullAlbum(ctx, userID, spotifyID)
	if err != nil {
		return fmt.Errorf("failed to fetch spotify album: %w", err)
	}
	dto := spotifyAlbumToDTO(spotifyAlbum)

	return s.db.WithTx(func(tx *db.DB) error {
		txRepo := NewRepo(tx.Queries())
		imported, err := txRepo.EnsureAlbumWithMetadata(ctx, dto)
		if err != nil {
			return fmt.Errorf("failed to import album metadata: %w", err)
		}
		hasRelease, err := txRepo.HasOwnedOrWishlistedReleaseForAlbum(ctx, userID, imported.ID)
		if err != nil {
			return fmt.Errorf("failed to check user releases: %w", err)
		}
		if hasRelease {
			return ErrAlbumAlreadyDecided
		}
		if err := txRepo.AddAlbumToRadar(ctx, userID, imported.ID); err != nil {
			return fmt.Errorf("failed to add album to radar: %w", err)
		}
		return nil
	})
}
