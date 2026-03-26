package adapters

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/core/task"
	"github.com/alecdray/wax/src/internal/feed"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/musicbrainz"
	"github.com/alecdray/wax/src/internal/spotify"
)

type HttpHandler struct {
	spotifyAuth *spotify.AuthService
	mb          *musicbrainz.Service
	feedService *feed.Service
	libraryService *library.Service
	taskManager *task.TaskManager
}

func NewHttpHandler(spotifyAuth *spotify.AuthService, mb *musicbrainz.Service, feedService *feed.Service, libraryService *library.Service, taskManager *task.TaskManager) *HttpHandler {
	return &HttpHandler{
		spotifyAuth:    spotifyAuth,
		mb:             mb,
		feedService:    feedService,
		libraryService: libraryService,
		taskManager:    taskManager,
	}
}

func parseFilterParams(r *http.Request) library.FilterParams {
	q := r.URL.Query()
	var fp library.FilterParams
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
	if format := q.Get("format"); format != "" {
		fp.Formats = []models.ReleaseFormat{models.ReleaseFormat(format)}
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

	lib.Albums.SortByDate(false)

	recentAlbums, err := h.libraryService.GetRecentlyPlayedAlbums(ctx, userId)
	if err != nil {
		err = fmt.Errorf("failed to get recently played albums: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dashboardPage := DashboardPage(DashboardPageProps{
		Library:         lib,
		Feeds:           feeds,
		RecentAlbums:    recentAlbums,
		FirstPageAlbums: lib.Albums.Page(0),
		Artists:         lib.Artists,
		FilterParams:    library.FilterParams{},
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

	view := CarouselView(r.URL.Query().Get("view"))
	if view == "" {
		view = CarouselViewRecentlyPlayed
	}

	var albums []library.AlbumSummaryDTO
	switch view {
	case CarouselViewUnrated:
		albums, err = h.libraryService.GetUnratedAlbums(ctx, userId)
	default:
		view = CarouselViewRecentlyPlayed
		albums, err = h.libraryService.GetRecentlyPlayedAlbums(ctx, userId)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	CarouselSection(albums, view).Render(r.Context(), w)
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

	ascending := dir != "desc"

	switch sortBy {
	case "album":
		albums.SortByTitle(ascending)
	case "artist":
		albums.SortByArtist(ascending)
	case "rating":
		albums.SortByRating(ascending)
	case "date":
		albums.SortByDate(ascending)
	case "lastPlayed":
		albums.SortByLastPlayed(ascending)
	}

	fp := parseFilterParams(r)
	albums = albums.Filter(fp)

	component := AlbumsList(albums.Page(0), sortBy, dir, fp, lib.Artists)
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

	contentComponent := FeedsDropdownContent(feeds)
	contentComponent.Render(r.Context(), w)

	buttonComponent := FeedsDropdownButton(feeds, true)
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

	ascending := dir != "desc"
	albums := lib.Albums
	switch sortBy {
	case "album":
		albums.SortByTitle(ascending)
	case "artist":
		albums.SortByArtist(ascending)
	case "rating":
		albums.SortByRating(ascending)
	case "date":
		albums.SortByDate(ascending)
	case "lastPlayed":
		albums.SortByLastPlayed(ascending)
	}

	fp := parseFilterParams(r)
	albums = albums.Filter(fp)

	page := albums.Page(offset)
	if len(page) == 0 {
		return
	}

	albumsListBody(page, offset, sortBy, dir, fp).Render(r.Context(), w)
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

	AlbumDetailPage(*album).Render(r.Context(), w)
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

	// Render content first
	contentComponent := FeedsDropdownContent(feeds)
	contentComponent.Render(r.Context(), w)

	// Render button as OOB swap
	buttonComponent := FeedsDropdownButton(feeds, true)
	buttonComponent.Render(r.Context(), w)
}
