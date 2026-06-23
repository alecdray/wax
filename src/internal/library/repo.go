package library

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/db/sqlc"
	"github.com/alecdray/wax/src/internal/review"

	"github.com/google/uuid"
)

// Repo is the library module's data access layer. It is the only file in
// package library that imports core/db/sqlc. Repo methods return library
// DTOs (AlbumDTO, ArtistDTO, etc.) — never sqlc.* types.
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

func releaseDTOFromModel(model sqlc.Release, userRelease *sqlc.UserRelease) ReleaseDTO {
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
		dto.Status = UserReleaseStatus(userRelease.Status)
		dto.CreatedAt = &userRelease.CreatedAt
		dto.StatusUpdatedAt = &userRelease.StatusUpdatedAt
		if dto.Status == UserReleaseStatusOwned {
			dto.AddedAt = &userRelease.StatusUpdatedAt
		}
	}
	return dto
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
	if ur != nil && ur.Status == string(UserReleaseStatusOwned) {
		dto.Owned = true
		dto.AddedAt = &ur.StatusUpdatedAt
	}
	return dto
}

func trackDTOFromModel(model sqlc.Track) TrackDTO {
	return TrackDTO{
		ID:        model.ID,
		SpotifyID: model.SpotifyID,
		Title:     model.Title,
	}
}

func artistDTOFromModel(model sqlc.Artist) ArtistDTO {
	return ArtistDTO{
		ID:        model.ID,
		SpotifyID: model.SpotifyID,
		Name:      model.Name,
	}
}

func albumDTOFromModel(model sqlc.Album, artists []ArtistDTO, tracks []TrackDTO, releases []ReleaseDTO, rating *review.AlbumRatingDTO) AlbumDTO {
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

// --- Album lookups ---

// GetAlbumByID returns the album row converted to a partial AlbumDTO with no
// artists/tracks/releases/rating populated. Callers compose the rest.
func (r *Repo) GetAlbumByID(ctx context.Context, albumID string) (*AlbumDTO, error) {
	model, err := r.q.GetAlbum(ctx, albumID)
	if err != nil {
		return nil, err
	}
	dto := albumDTOFromModel(model, nil, nil, nil, nil)
	return &dto, nil
}

// GetAlbumSpotifyID returns the spotify ID for an album. Used by callers that
// need to talk to the Spotify API without rebuilding a full DTO.
func (r *Repo) GetAlbumSpotifyID(ctx context.Context, albumID string) (string, error) {
	model, err := r.q.GetAlbum(ctx, albumID)
	if err != nil {
		return "", err
	}
	return model.SpotifyID, nil
}

// GetAlbumsByIDs returns base AlbumDTOs (no artists/tracks/releases/rating)
// for the given album IDs.
func (r *Repo) GetAlbumsByIDs(ctx context.Context, albumIDs []string) ([]AlbumDTO, error) {
	rows, err := r.q.GetAlbumsByIDs(ctx, albumIDs)
	if err != nil {
		return nil, err
	}
	dtos := make([]AlbumDTO, len(rows))
	for i, row := range rows {
		dtos[i] = albumDTOFromModel(row, nil, nil, nil, nil)
	}
	return dtos, nil
}

// --- Artist lookups ---

func (r *Repo) GetArtistsByAlbumID(ctx context.Context, albumID string) ([]ArtistDTO, error) {
	rows, err := r.q.GetAlbumArtistByAlbumId(ctx, albumID)
	if err != nil {
		return nil, err
	}
	out := make([]ArtistDTO, len(rows))
	for i, row := range rows {
		out[i] = artistDTOFromModel(row.Artist)
	}
	return out, nil
}

// GetArtistsByAlbumIDs returns artists grouped by album ID.
func (r *Repo) GetArtistsByAlbumIDs(ctx context.Context, albumIDs []string) (map[string][]ArtistDTO, error) {
	rows, err := r.q.GetAlbumArtistsByAlbumIds(ctx, albumIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]ArtistDTO, len(albumIDs))
	for _, row := range rows {
		result[row.AlbumID] = append(result[row.AlbumID], artistDTOFromModel(row.Artist))
	}
	return result, nil
}

// --- Track lookups ---

func (r *Repo) GetTracksByAlbumID(ctx context.Context, albumID string) ([]TrackDTO, error) {
	rows, err := r.q.GetAlbumTracksByAlbumId(ctx, albumID)
	if err != nil {
		return nil, err
	}
	out := make([]TrackDTO, len(rows))
	for i, row := range rows {
		out[i] = trackDTOFromModel(row.Track)
	}
	return out, nil
}

// GetTracksByAlbumIDs returns tracks grouped by album ID.
func (r *Repo) GetTracksByAlbumIDs(ctx context.Context, albumIDs []string) (map[string][]TrackDTO, error) {
	rows, err := r.q.GetAlbumTracksByAlbumIds(ctx, albumIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]TrackDTO, len(albumIDs))
	for _, row := range rows {
		result[row.AlbumID] = append(result[row.AlbumID], trackDTOFromModel(row.Track))
	}
	return result, nil
}

// --- Release / user-release lookups ---

// GetUserReleases returns all releases the user owns (not soft-deleted).
func (r *Repo) GetUserReleases(ctx context.Context, userID string) ([]ReleaseDTO, error) {
	rows, err := r.q.GetUserReleases(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ReleaseDTO, len(rows))
	for i, row := range rows {
		out[i] = releaseDTOFromModel(row.Release, &row.UserRelease)
	}
	return out, nil
}

// GetUserReleasesByAlbumID returns the user's owned releases for one album.
func (r *Repo) GetUserReleasesByAlbumID(ctx context.Context, userID, albumID string) ([]ReleaseDTO, error) {
	rows, err := r.q.GetUserReleasesByAlbumId(ctx, sqlc.GetUserReleasesByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]ReleaseDTO, len(rows))
	for i, row := range rows {
		out[i] = releaseDTOFromModel(row.Release, &row.UserRelease)
	}
	return out, nil
}

// HasOwnedOrWishlistedReleaseForAlbum reports whether the user currently owns
// or wishlists the album (i.e. it is in the library). A `removed` release does
// not count — such an album is still radar-eligible (ADR 0005).
func (r *Repo) HasOwnedOrWishlistedReleaseForAlbum(ctx context.Context, userID, albumID string) (bool, error) {
	hasRelease, err := r.q.HasOwnedOrWishlistedReleaseForAlbum(ctx, sqlc.HasOwnedOrWishlistedReleaseForAlbumParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return false, err
	}
	return hasRelease != 0, nil
}

// GetAlbumFormats returns the four-format AlbumFormatDTO list for an album,
// merging all known releases with the user's ownership.
func (r *Repo) GetAlbumFormats(ctx context.Context, userID, albumID string) ([]AlbumFormatDTO, error) {
	allReleases, err := r.q.GetReleases(ctx, albumID)
	if err != nil {
		return nil, err
	}
	userReleases, err := r.q.GetUserReleasesByAlbumId(ctx, sqlc.GetUserReleasesByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return nil, err
	}
	return buildAlbumFormats(allReleases, userReleases), nil
}

// --- Summary / queue queries ---

// GetRecentlyPlayedAlbums returns the recently-played album summaries.
func (r *Repo) GetRecentlyPlayedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	rows, err := r.q.GetRecentlyPlayedAlbums(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]AlbumSummaryDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, AlbumSummaryDTO{
			ID:        row.ID,
			SpotifyID: row.SpotifyID,
			Title:     row.Title,
			Artists:   fmt.Sprintf("%s", row.ArtistNames),
			ImageURL:  row.ImageUrl.String,
			InLibrary: row.InLibrary != 0,
			OnRadar:   row.OnRadar != 0,
		})
	}
	return out, nil
}

// GetUnratedAlbums returns library albums with no rating yet.
func (r *Repo) GetUnratedAlbums(ctx context.Context, userID string) ([]AlbumSummaryDTO, error) {
	rows, err := r.q.GetUnratedAlbums(ctx, sqlc.GetUnratedAlbumsParams{
		UserID:   userID,
		UserID_2: userID,
		UserID_3: userID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]AlbumSummaryDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, AlbumSummaryDTO{
			ID:        row.ID,
			SpotifyID: row.SpotifyID,
			Title:     row.Title,
			Artists:   fmt.Sprintf("%s", row.ArtistNames),
			ImageURL:  row.ImageUrl.String,
			InLibrary: true,
		})
	}
	return out, nil
}

// GetProvisionalAlbums returns every album in the user's library whose
// rating-state row is currently `provisional`.
func (r *Repo) GetProvisionalAlbums(ctx context.Context, userID string) ([]ProvisionalAlbumDTO, error) {
	rows, err := r.q.GetProvisionalAlbums(ctx, sqlc.GetProvisionalAlbumsParams{
		UserID:   userID,
		UserID_2: userID,
		UserID_3: userID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]ProvisionalAlbumDTO, 0, len(rows))
	for _, row := range rows {
		dto := ProvisionalAlbumDTO{
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
		// HasRating reflects whether a row in album_rating_log exists for this
		// (user, album); a provisional album with no log row keeps Rating == nil.
		if row.HasRating != 0 {
			rating := row.LatestRating
			dto.Rating = &rating
		}
		out = append(out, dto)
	}
	return out, nil
}

// --- Mutations: collection upserts ---

// EnsureAlbumWithMetadata creates or updates the album row, its artists, its
// tracks, and its releases — but does NOT touch user_releases or
// user_album_radar. Used by both the collection-add flow (which then writes
// user_releases) and the radar-add flow (which then writes user_album_radar).
//
// Returns the canonical AlbumDTO with sqlc-derived IDs filled in.
func (r *Repo) EnsureAlbumWithMetadata(ctx context.Context, album AlbumDTO) (AlbumDTO, error) {
	albumModel, err := r.q.GetOrCreateAlbum(ctx, sqlc.GetOrCreateAlbumParams{
		ID:        album.ID,
		SpotifyID: album.SpotifyID,
		Title:     album.Title,
		ImageUrl:  sql.NullString{String: album.ImageURL, Valid: album.ImageURL != ""},
	})
	if err != nil {
		return album, err
	}
	album = albumDTOFromModel(albumModel, album.Artists, album.Tracks, album.Releases, album.Rating)

	for i, track := range album.Tracks {
		trackModel, err := r.q.GetOrCreateTrack(ctx, sqlc.GetOrCreateTrackParams{
			ID:        track.ID,
			SpotifyID: track.SpotifyID,
			Title:     track.Title,
		})
		if err != nil {
			return album, err
		}
		if _, err := r.q.GetOrCreateAlbumTrack(ctx, sqlc.GetOrCreateAlbumTrackParams{
			AlbumID: albumModel.ID,
			TrackID: trackModel.ID,
		}); err != nil {
			return album, err
		}
		album.Tracks[i] = trackDTOFromModel(trackModel)
	}

	for i, artist := range album.Artists {
		artistModel, err := r.q.GetOrCreateArtist(ctx, sqlc.GetOrCreateArtistParams{
			ID:        artist.ID,
			SpotifyID: artist.SpotifyID,
			Name:      artist.Name,
		})
		if err != nil {
			return album, err
		}
		if _, err := r.q.GetOrCreateAlbumArtist(ctx, sqlc.GetOrCreateAlbumArtistParams{
			AlbumID:  albumModel.ID,
			ArtistID: artistModel.ID,
		}); err != nil {
			return album, err
		}
		album.Artists[i] = artistDTOFromModel(artistModel)
	}

	for i, release := range album.Releases {
		releaseModel, err := r.q.GetOrCreateRelease(ctx, sqlc.GetOrCreateReleaseParams{
			ID:      release.ID,
			AlbumID: albumModel.ID,
			Format:  release.Format,
		})
		if err != nil {
			return album, err
		}
		album.Releases[i] = releaseDTOFromModel(releaseModel, nil)
	}

	return album, nil
}

// AddAlbumToCollection imports the album's metadata and writes an owned
// user_releases row for every release on the AlbumDTO. The caller is
// responsible for clearing the album's radar entry (the cross-cutting rule
// lives at the service layer).
func (r *Repo) AddAlbumToCollection(ctx context.Context, userID string, album AlbumDTO) (AlbumDTO, error) {
	album, err := r.EnsureAlbumWithMetadata(ctx, album)
	if err != nil {
		return album, err
	}
	for i, release := range album.Releases {
		now := time.Now()
		userRelease, err := r.q.UpsertOwnedRelease(ctx, sqlc.UpsertOwnedReleaseParams{
			ID:              uuid.New().String(),
			UserID:          userID,
			ReleaseID:       release.ID,
			CreatedAt:       now,
			StatusUpdatedAt: now,
		})
		if err != nil {
			return album, err
		}
		// UpsertOwnedRelease doesn't return the full release row, so we
		// rebuild a sqlc.Release from the input DTO. ReleasedAt is dropped
		// here — the current caller discards the return value, but a future
		// caller that needs ReleasedAt should re-fetch the release.
		album.Releases[i] = releaseDTOFromModel(sqlc.Release{
			ID:        release.ID,
			AlbumID:   album.ID,
			Format:    release.Format,
			DiscogsID: sql.NullString{String: release.DiscogsID, Valid: release.DiscogsID != ""},
			Label:     sql.NullString{String: release.Label, Valid: release.Label != ""},
		}, &userRelease)
	}
	return album, nil
}

// AddOwnedRelease ensures a release exists for the given album/format and
// records the user as owning it. Returns the release ID. If releaseID is
// empty, a new release row is created (with a fresh UUID).
func (r *Repo) AddOwnedRelease(ctx context.Context, userID, albumID string, format models.ReleaseFormat, releaseID string, addedAt time.Time) (string, error) {
	if releaseID == "" {
		rel, err := r.q.GetOrCreateRelease(ctx, sqlc.GetOrCreateReleaseParams{
			ID:      uuid.New().String(),
			AlbumID: albumID,
			Format:  format,
		})
		if err != nil {
			return "", err
		}
		releaseID = rel.ID
	}
	if _, err := r.q.UpsertOwnedRelease(ctx, sqlc.UpsertOwnedReleaseParams{
		ID:              uuid.New().String(),
		UserID:          userID,
		ReleaseID:       releaseID,
		CreatedAt:       addedAt,
		StatusUpdatedAt: addedAt,
	}); err != nil {
		return "", err
	}
	return releaseID, nil
}

// MarkReleaseRemoved transitions an owned user_release to status='removed'.
func (r *Repo) MarkReleaseRemoved(ctx context.Context, userID, releaseID string) error {
	return r.q.MarkReleaseRemoved(ctx, sqlc.MarkReleaseRemovedParams{
		UserID:    userID,
		ReleaseID: releaseID,
	})
}

// MarkReleasesRemovedByAlbumID transitions all of a user's owned releases for an album to 'removed'.
func (r *Repo) MarkReleasesRemovedByAlbumID(ctx context.Context, userID, albumID string) error {
	return r.q.MarkReleasesRemovedByAlbumId(ctx, sqlc.MarkReleasesRemovedByAlbumIdParams{
		UserID:  userID,
		AlbumID: albumID,
	})
}

// UpdateReleaseDiscogsInfo writes Discogs-derived metadata onto a release.
// Pass empty string / nil time values to leave the column unset (sqlc params
// use sql.NullString / sql.NullTime; an unset Valid field clears the value).
func (r *Repo) UpdateReleaseDiscogsInfo(ctx context.Context, releaseID, discogsID, label string, releasedAt *time.Time) error {
	var rel sql.NullTime
	if releasedAt != nil {
		rel = sql.NullTime{Time: *releasedAt, Valid: true}
	}
	return r.q.UpdateReleaseDiscogsInfo(ctx, sqlc.UpdateReleaseDiscogsInfoParams{
		ID:         releaseID,
		DiscogsID:  sql.NullString{String: discogsID, Valid: discogsID != ""},
		Label:      sql.NullString{String: label, Valid: label != ""},
		ReleasedAt: rel,
	})
}

// ClearReleaseDiscogsInfo wipes the Discogs metadata on a release.
func (r *Repo) ClearReleaseDiscogsInfo(ctx context.Context, releaseID string) error {
	return r.q.UpdateReleaseDiscogsInfo(ctx, sqlc.UpdateReleaseDiscogsInfoParams{
		ID: releaseID,
	})
}

// --- Radar (album-level "want to listen") ---

// AddAlbumToRadar inserts a radar row, idempotent on (user_id, album_id).
func (r *Repo) AddAlbumToRadar(ctx context.Context, userID, albumID string) error {
	_, err := r.q.AddAlbumToRadar(ctx, sqlc.AddAlbumToRadarParams{
		ID:      uuid.New().String(),
		UserID:  userID,
		AlbumID: albumID,
	})
	return err
}

// RemoveAlbumFromRadar deletes the radar row if present (no-op otherwise).
func (r *Repo) RemoveAlbumFromRadar(ctx context.Context, userID, albumID string) error {
	return r.q.RemoveAlbumFromRadar(ctx, sqlc.RemoveAlbumFromRadarParams{
		UserID:  userID,
		AlbumID: albumID,
	})
}

// GetRadarAlbums returns the radar entries that have no user_releases activity.
// Excludes albums that the user has wishlisted, owns, or has removed — radar is strictly pre-decision.
func (r *Repo) GetRadarAlbums(ctx context.Context, userID string) ([]RadarDTO, []AlbumDTO, error) {
	rows, err := r.q.GetRadarAlbums(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	radarDTOs := make([]RadarDTO, len(rows))
	albumDTOs := make([]AlbumDTO, len(rows))
	for i, row := range rows {
		radarDTOs[i] = RadarDTO{
			AlbumID:   row.UserAlbumRadar.AlbumID,
			CreatedAt: row.UserAlbumRadar.CreatedAt,
		}
		albumDTOs[i] = albumDTOFromModel(row.Album, nil, nil, nil, nil)
	}
	return radarDTOs, albumDTOs, nil
}

// IsAlbumOnRadar reports whether the (user, album) pair has a radar row.
func (r *Repo) IsAlbumOnRadar(ctx context.Context, userID, albumID string) (bool, error) {
	onRadar, err := r.q.IsAlbumOnRadar(ctx, sqlc.IsAlbumOnRadarParams{
		UserID:  userID,
		AlbumID: albumID,
	})
	if err != nil {
		return false, err
	}
	return onRadar != 0, nil
}

// --- Wishlist (release-level pre-acquisition) ---

// AddReleaseToWishlist upserts a user_release in the 'wishlist' state.
// If a release row doesn't yet exist for (album, format), one is created.
func (r *Repo) AddReleaseToWishlist(ctx context.Context, userID, albumID string, format models.ReleaseFormat, releaseID string) (string, error) {
	if releaseID == "" {
		rel, err := r.q.GetOrCreateRelease(ctx, sqlc.GetOrCreateReleaseParams{
			ID:      uuid.New().String(),
			AlbumID: albumID,
			Format:  format,
		})
		if err != nil {
			return "", err
		}
		releaseID = rel.ID
	}
	now := time.Now()
	if _, err := r.q.UpsertWishlistRelease(ctx, sqlc.UpsertWishlistReleaseParams{
		ID:              uuid.New().String(),
		UserID:          userID,
		ReleaseID:       releaseID,
		CreatedAt:       now,
		StatusUpdatedAt: now,
	}); err != nil {
		return "", err
	}
	return releaseID, nil
}

// RemoveReleaseFromWishlist hard-deletes the wishlist row.
// Owned and removed rows are untouched (the WHERE clause filters on status='wishlist').
func (r *Repo) RemoveReleaseFromWishlist(ctx context.Context, userID, releaseID string) error {
	return r.q.DeleteWishlistRelease(ctx, sqlc.DeleteWishlistReleaseParams{
		UserID:    userID,
		ReleaseID: releaseID,
	})
}

// GetWishlistReleases returns the user's wishlist releases.
func (r *Repo) GetWishlistReleases(ctx context.Context, userID string) ([]ReleaseDTO, error) {
	rows, err := r.q.GetWishlistReleases(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ReleaseDTO, len(rows))
	for i, row := range rows {
		out[i] = releaseDTOFromModel(row.Release, &row.UserRelease)
	}
	return out, nil
}

// MarkReleaseOwned transitions any user_release row (wishlist or removed) to 'owned'.
// Used by the wishlist acquire flow and any "re-acquire a removed release" path
// that doesn't need to create a new release row.
//
// CreatedAt is only used when no row exists yet (fresh insert). On conflict, the
// upsert preserves the existing row's created_at — passing time.Now() here is harmless.
func (r *Repo) MarkReleaseOwned(ctx context.Context, userID, releaseID string) error {
	now := time.Now()
	if _, err := r.q.UpsertOwnedRelease(ctx, sqlc.UpsertOwnedReleaseParams{
		ID:              uuid.New().String(),
		UserID:          userID,
		ReleaseID:       releaseID,
		CreatedAt:       now,
		StatusUpdatedAt: now,
	}); err != nil {
		return err
	}
	return nil
}

// GetUserAlbumStateBySpotifyIDs returns the caller's wax state for each
// Spotify ID. Missing keys mean the user has no row for that album. When an
// album would qualify for multiple states (defensive — invariants forbid it),
// in_library wins, then on_radar, then removed.
func (r *Repo) GetUserAlbumStateBySpotifyIDs(ctx context.Context, userID string, spotifyIDs []string) (map[string]UserAlbumStateRow, error) {
	if len(spotifyIDs) == 0 {
		return map[string]UserAlbumStateRow{}, nil
	}
	rows, err := r.q.GetUserAlbumStateBySpotifyIds(ctx, sqlc.GetUserAlbumStateBySpotifyIdsParams{
		UserID:     userID,
		UserID_2:   userID,
		UserID_3:   userID,
		SpotifyIds: spotifyIDs,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]UserAlbumStateRow, len(rows))
	rank := func(s DiscoverAlbumState) int {
		switch s {
		case DiscoverAlbumStateInLibrary:
			return 3
		case DiscoverAlbumStateOnRadar:
			return 2
		case DiscoverAlbumStateRemoved:
			return 1
		default:
			return 0
		}
	}
	for _, row := range rows {
		next := UserAlbumStateRow{
			AlbumID: row.AlbumID,
			State:   DiscoverAlbumState(row.State),
		}
		if cur, ok := out[row.SpotifyID]; !ok || rank(next.State) > rank(cur.State) {
			out[row.SpotifyID] = next
		}
	}
	return out, nil
}
