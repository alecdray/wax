package adapters

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/core/task"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/feed"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/library/adapters/views"
	"github.com/alecdray/wax/src/internal/musicbrainz"
	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/spotify"
)

type HttpHandler struct {
	spotifyAuth    *spotify.AuthService
	mb             *musicbrainz.Service
	feedService    *feed.Service
	libraryService *library.Service
	taskManager    *task.TaskManager
	discogsService *discogs.Service
	notesService   *notes.Service
}

func NewHttpHandler(spotifyAuth *spotify.AuthService, mb *musicbrainz.Service, feedService *feed.Service, libraryService *library.Service, taskManager *task.TaskManager, discogsService *discogs.Service, notesService *notes.Service) *HttpHandler {
	return &HttpHandler{
		spotifyAuth:    spotifyAuth,
		mb:             mb,
		feedService:    feedService,
		libraryService: libraryService,
		taskManager:    taskManager,
		discogsService: discogsService,
		notesService:   notesService,
	}
}

func parseFilterParams(r *http.Request) library.FilterParams {
	q := r.URL.Query()
	var fp library.FilterParams
	fp.Q = strings.TrimSpace(q.Get("q"))
	if minStr := q.Get("minRating"); minStr != "" {
		if v, err := strconv.ParseFloat(minStr, 64); err == nil {
			fp.MinRating = &v
		}
	}
	if maxStr := q.Get("maxRating"); maxStr != "" {
		if v, err := strconv.ParseFloat(maxStr, 64); err == nil {
			fp.MaxRating = &v
		}
	}
	fp.Rated = q.Get("rated")
	for _, format := range q["format"] {
		if format == "" {
			continue
		}
		fp.Formats = append(fp.Formats, models.ReleaseFormat(format))
	}
	fp.ArtistIDs = q["artist"]
	return fp
}

func (h *HttpHandler) GetDashboardPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feeds, err := h.feedService.GetUsersFeeds(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range feeds {
		if f.Kind == models.FeedKindSpotify && f.LastSyncStatus.IsUnsyned() {
			h.taskManager.RegisterAdHocTask(feed.NewSyncSpotifyFeedTask(h.feedService, f))
		}
	}

	lib, err := h.libraryService.GetLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get library: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Honor URL query params on initial render so deep links + reloads
	// reproduce the same view as HTMX-driven interactions.
	sortBy := r.URL.Query().Get("sortBy")
	dir := r.URL.Query().Get("dir")
	ascending := dir == "asc"
	switch sortBy {
	case "album":
		lib.Albums.SortByTitle(ascending)
	case "artist":
		lib.Albums.SortByArtist(ascending)
	case "rating":
		lib.Albums.SortByRating(ascending)
	case "lastPlayed":
		lib.Albums.SortByLastPlayed(ascending)
	default:
		lib.Albums.SortByDate(ascending)
	}

	fp := parseFilterParams(r)
	filtered := lib.Albums.Filter(fp)

	recentAlbums, err := h.libraryService.GetRecentlyPlayedAlbums(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get recently played albums: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dashboardPage := views.DashboardPage(views.DashboardPageProps{
		Library:         lib,
		Feeds:           feeds,
		RecentAlbums:    recentAlbums,
		FirstPageAlbums: filtered.Page(0),
		Artists:         lib.Artists,
		FilterParams:    fp,
		SortBy:          sortBy,
		SortDir:         dir,
	})
	dashboardPage.Render(r.Context(), w)
}

func (h *HttpHandler) GetCarousel(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := views.CarouselView(r.URL.Query().Get("view"))
	if view == "" {
		view = views.CarouselViewRecentlyPlayed
	}

	props := views.CarouselSectionProps{Active: view}

	switch view {
	case views.CarouselViewUnrated:
		albums, err := h.libraryService.GetUnratedAlbums(ctx, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.RegularAlbums = albums
	case views.CarouselViewProvisional:
		albums, err := h.libraryService.GetProvisionalAlbums(ctx, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.ProvisionalAlbums = albums
	default:
		props.Active = views.CarouselViewRecentlyPlayed
		albums, err := h.libraryService.GetRecentlyPlayedAlbums(ctx, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.RegularAlbums = albums
	}

	views.CarouselSectionFrag(props).Render(r.Context(), w)
}

func (h *HttpHandler) GetAlbumsTable(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lib, err := h.libraryService.GetLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get library: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if lib == nil {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	albums := lib.Albums
	sortBy := r.URL.Query().Get("sortBy")
	dir := r.URL.Query().Get("dir")

	ascending := dir == "asc"

	switch sortBy {
	case "album":
		albums.SortByTitle(ascending)
	case "artist":
		albums.SortByArtist(ascending)
	case "rating":
		albums.SortByRating(ascending)
	case "lastPlayed":
		albums.SortByLastPlayed(ascending)
	default:
		albums.SortByDate(ascending)
	}

	fp := parseFilterParams(r)
	albums = albums.Filter(fp)

	component := views.AlbumsListFrag(albums.Page(0), sortBy, dir, fp, lib.Artists)
	component.Render(r.Context(), w)
}

func (h *HttpHandler) TriggerFeedSync(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feedId := r.URL.Query().Get("feedId")
	if feedId == "" {
		http.Error(w, "feedId is required", http.StatusBadRequest)
		return
	}

	f, err := h.feedService.GetFeedByID(ctx, feedId, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if f.Kind == models.FeedKindSpotify && !f.LastSyncStatus.IsSyncing() {
		h.taskManager.RegisterAdHocTask(feed.NewSyncSpotifyFeedTask(h.feedService, *f))
	}

	feeds, err := h.feedService.GetUsersFeeds(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, feed := range feeds {
		if feed.ID == f.ID {
			feed.SetSyncing()
			feeds[i] = feed
			break
		}
	}

	contentComponent := views.FeedsDropdownContentFrag(feeds)
	contentComponent.Render(r.Context(), w)

	buttonComponent := views.FeedsDropdownButtonFrag(feeds, true)
	buttonComponent.Render(r.Context(), w)
}

func (h *HttpHandler) GetAlbumsPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lib, err := h.libraryService.GetLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get library: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sortBy := r.URL.Query().Get("sortBy")
	dir := r.URL.Query().Get("dir")
	offset := 0
	fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)

	ascending := dir == "asc"
	albums := lib.Albums
	switch sortBy {
	case "album":
		albums.SortByTitle(ascending)
	case "artist":
		albums.SortByArtist(ascending)
	case "rating":
		albums.SortByRating(ascending)
	case "lastPlayed":
		albums.SortByLastPlayed(ascending)
	default:
		albums.SortByDate(ascending)
	}

	fp := parseFilterParams(r)
	albums = albums.Filter(fp)

	page := albums.Page(offset)
	if len(page) == 0 {
		return
	}

	views.AlbumsListBodyFrag(page, offset, sortBy, dir, fp).Render(r.Context(), w)
}

func (h *HttpHandler) GetAlbumDetailPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	albumId := r.PathValue("albumId")
	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "album not in library" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views.AlbumDetailPage(*album).Render(r.Context(), w)
}

func (h *HttpHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	albumId := r.PathValue("albumId")
	if err := h.libraryService.RemoveAlbumFromLibrary(ctx, userId, albumId); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to remove album: %w", err),
		})
		return
	}
	w.Header().Set("HX-Redirect", "/app/library/dashboard")
	w.WriteHeader(http.StatusOK)
}

func anyFeedSyncing(feeds []feed.FeedDTO) bool {
	for _, f := range feeds {
		if f.LastSyncStatus.IsSyncing() {
			return true
		}
	}
	return false
}

func (h *HttpHandler) GetFeedsDropdown(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feeds, err := h.feedService.GetUsersFeeds(ctx, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wasSyncing := r.URL.Query().Get("wasSyncing") == "true"
	nowSyncing := anyFeedSyncing(feeds)
	if wasSyncing && !nowSyncing {
		w.Header().Set("HX-Trigger", "libraryUpdated")
	}

	// Render content first
	contentComponent := views.FeedsDropdownContentFrag(feeds)
	contentComponent.Render(r.Context(), w)

	// Render button as OOB swap
	buttonComponent := views.FeedsDropdownButtonFrag(feeds, true)
	buttonComponent.Render(r.Context(), w)
}

func (h *HttpHandler) GetLibraryStats(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lib, err := h.libraryService.GetLibrary(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get library: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views.LibraryStatsFrag(lib).Render(r.Context(), w)
}

func (h *HttpHandler) GetSleeveNotesEditor(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	albumId := r.PathValue("albumId")
	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return
	}

	err = views.SleeveNotesEditorFrag(*album, "").Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render response: %w", err),
		})
	}
}

func (h *HttpHandler) GetSleeveNotesView(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	albumId := r.PathValue("albumId")
	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return
	}

	err = views.SleeveNotesSectionFrag(*album, false).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render response: %w", err),
		})
	}
}

func (h *HttpHandler) SaveSleeveNote(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	albumId := r.PathValue("albumId")
	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	content := r.FormValue("content")
	if len(content) > notes.MaxSleeveNoteLength {
		album, _ := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
		if album != nil {
			album.SleeveNote = nil
		} else {
			album = &library.AlbumDTO{ID: albumId}
		}
		err = views.SleeveNotesEditorFrag(*album, fmt.Sprintf("Note exceeds maximum length of %d characters.", notes.MaxSleeveNoteLength)).Render(ctx, w)
		if err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    err,
			})
		}
		return
	}

	sleeveNote, err := h.notesService.UpsertAlbumNote(ctx, userId, albumId, content)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to save sleeve note: %w", err),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return
	}
	album.SleeveNote = sleeveNote

	err = views.SleeveNotesSectionFrag(*album, false).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = views.AlbumRowTagsSectionFrag(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
	}
}

// GetAlbumSurfaces re-renders the album surfaces affected by a rating or tag
// change as OOB swaps. It is driven by the `album-changed` HTMX event fired by
// peer modules after they mutate album state. Surfaces whose targets are not on
// the current page are no-ops on the client.
//
// The `scope` query param selects which surfaces refresh, matching the
// per-mutation sets the peer modules previously rendered inline:
//   - "tags"    → album tags + row-tags section
//   - "rating" (or absent/unknown) → score readout, score badge, rating history,
//     row-tags section
func (h *HttpHandler) GetAlbumSurfaces(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	albumId := r.URL.Query().Get("albumId")
	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "album not in library" {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusNotFound,
				Err:    fmt.Errorf("album not found: %w", err),
			})
			return
		}
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return
	}

	scope := r.URL.Query().Get("scope")
	if scope == "tags" {
		if err := views.AlbumTagsFrag(*album, true).Render(ctx, w); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    err,
			})
			return
		}
		if err := views.AlbumRowTagsSectionFrag(*album, true).Render(ctx, w); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    err,
			})
			return
		}
	} else {
		if err := views.AlbumScoreReadoutFrag(*album, true).Render(ctx, w); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    err,
			})
			return
		}
		if err := views.AlbumScoreBadgeFrag(*album, true).Render(ctx, w); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    err,
			})
			return
		}
		if err := views.AlbumRatingHistoryFrag(*album, true).Render(ctx, w); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    err,
			})
			return
		}
		if err := views.AlbumRowTagsSectionFrag(*album, true).Render(ctx, w); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    err,
			})
			return
		}
	}
}

// --- Discover page ---

func (h *HttpHandler) GetDiscoverPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	radar, err := h.libraryService.GetRadarAlbums(ctx, userId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get radar albums: %w", err),
		})
		return
	}
	radarInboxPlaylistID, err := h.feedService.RadarInboxPlaylistID(ctx, userId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get radar inbox state: %w", err),
		})
		return
	}
	feeds, err := h.feedService.GetUsersFeeds(ctx, userId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get feeds: %w", err),
		})
		return
	}
	views.DiscoverPage(views.DiscoverPageProps{
		RadarAlbums:          radar,
		Query:                "",
		SearchResults:        nil,
		RadarInboxPlaylistID: radarInboxPlaylistID,
		Feeds:                feeds,
	}).Render(r.Context(), w)
}

// GetRadarInbox renders the radar inbox control for its current state. The
// enabled control polls this so it flips back to Enable if the background sync
// removes the feed (the user deleted the playlist on Spotify).
func (h *HttpHandler) GetRadarInbox(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	playlistID, err := h.feedService.RadarInboxPlaylistID(ctx, userId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get radar inbox state: %w", err),
		})
		return
	}
	views.RadarInboxControlFrag(playlistID, "").Render(r.Context(), w)
}

// PostEnableRadarInbox opts the user into the Spotify radar inbox, creating their
// "wax radar" playlist, and swaps the control to show its link. If Spotify denies
// the request for missing playlist scope (a user connected before the scope was
// requested), the browser is redirected into re-authentication.
func (h *HttpHandler) PostEnableRadarInbox(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	playlistID, err := h.feedService.EnableRadarInbox(ctx, userId)
	if errors.Is(err, spotify.ErrInsufficientScope) {
		app, appErr := ctx.App()
		if appErr != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusInternalServerError,
				Err:    fmt.Errorf("failed to get app for re-auth: %w", appErr),
			})
			return
		}
		// Send the user through Spotify OAuth again to grant the playlist scope,
		// and remember to return them to discover afterwards so they can finish
		// enabling rather than landing on the dashboard.
		httpx.SetPostAuthRedirect(w, "/app/library/discover")
		w.Header().Set("HX-Redirect", h.spotifyAuth.AuthURLForcingConsent(app.Config().StateCode))
		w.WriteHeader(http.StatusOK)
		return
	}
	if err != nil {
		// Not a missing-scope problem (those re-auth above). Surface it inline
		// rather than looping the user or failing silently. The underlying
		// Spotify error is already logged in the service layer.
		views.RadarInboxControlFrag("", "Couldn’t enable — Spotify refused the request. Reconnect Spotify and try again.").Render(r.Context(), w)
		return
	}

	// Tell the header's feeds dropdown to refresh so the new radar feed appears
	// immediately rather than on its next poll.
	w.Header().Set("HX-Trigger", "feedsChanged")
	views.RadarInboxControlFrag(playlistID, "").Render(r.Context(), w)
}

func (h *HttpHandler) GetDiscoverRadar(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	radar, err := h.libraryService.GetRadarAlbums(ctx, userId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get radar albums: %w", err),
		})
		return
	}
	views.RadarCarouselFrag(radar, false).Render(r.Context(), w)
}

func (h *HttpHandler) GetAlbumActionsModal(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	spotifyID := r.URL.Query().Get("spotifyId")
	if spotifyID == "" {
		http.Error(w, "missing spotifyId", http.StatusBadRequest)
		return
	}
	result, err := h.libraryService.GetAlbumActionsResult(ctx, userId, spotifyID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to resolve album: %w", err),
		})
		return
	}
	views.AlbumActionsModalFrag(result).Render(r.Context(), w)
}

func (h *HttpHandler) GetDiscoverSearch(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		views.DiscoverSearchResultsFrag(nil, "").Render(r.Context(), w)
		return
	}
	results, err := h.libraryService.SearchAlbumsForDiscover(ctx, userId, query, 20)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("spotify search failed: %w", err),
		})
		return
	}
	views.DiscoverSearchResultsFrag(results, query).Render(r.Context(), w)
}

func (h *HttpHandler) PostDiscoverRadar(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	spotifyID := r.URL.Query().Get("spotifyId")
	if spotifyID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("missing spotifyId"),
		})
		return
	}
	if err := h.libraryService.AddSpotifyAlbumToRadar(ctx, userId, spotifyID); err != nil {
		if errors.Is(err, library.ErrAlbumAlreadyDecided) {
			// Album already has a user_releases row — UI was stale. The search-list
			// auto-refresh on radarUpdated will pull the actual current state.
			w.Header().Set("HX-Trigger", "radarUpdated")
			w.WriteHeader(http.StatusOK)
			return
		}
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to add album to radar: %w", err),
		})
		return
	}
	w.Header().Set("HX-Trigger", "radarUpdated")
	w.WriteHeader(http.StatusOK)
}

func (h *HttpHandler) DeleteAlbumRadar(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	albumID := r.PathValue("albumId")
	if err := h.libraryService.RemoveAlbumFromRadar(ctx, userId, albumID); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to remove from radar: %w", err),
		})
		return
	}
	w.Header().Set("HX-Trigger", "radarUpdated")
	w.WriteHeader(http.StatusOK)
}

func (h *HttpHandler) PostAlbumLibrary(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}
	albumID := r.PathValue("albumId")
	if err := h.libraryService.PromoteRadarToLibrary(ctx, userId, albumID); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to promote to library: %w", err),
		})
		return
	}
	w.Header().Set("HX-Trigger", "radarUpdated, libraryUpdated")
	w.WriteHeader(http.StatusOK)
}
