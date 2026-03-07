package adapters

import (
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db/models"
	"shmoopicks/src/internal/core/task"
	"shmoopicks/src/internal/feed"
	"shmoopicks/src/internal/library"
	"shmoopicks/src/internal/musicbrainz"
	"shmoopicks/src/internal/spotify"
)

type HttpHandler struct {
	spotifyAuth    *spotify.AuthService
	mb             *musicbrainz.Service
	feedService    *feed.Service
	libraryService *library.Service
	taskManager    *task.TaskManager
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

	dashboardPage := DashboardPage(DashboardPageProps{
		Library: lib,
		Feeds:   feeds,
	})
	dashboardPage.Render(r.Context(), w)
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

	// Default to ascending if not specified
	ascending := dir != "desc"

	// Sort albums based on sortBy parameter
	switch sortBy {
	case "album":
		albums.SortByTitle(ascending)
	case "artist":
		albums.SortByArtist(ascending)
	case "date":
		albums.SortByDate(ascending)
	}

	component := AlbumsTable(albums, sortBy, dir)
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
