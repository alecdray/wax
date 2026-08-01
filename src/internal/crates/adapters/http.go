package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/core/templates"
	"github.com/alecdray/wax/src/internal/crates"
	"github.com/alecdray/wax/src/internal/crates/adapters/views"
	"github.com/alecdray/wax/src/internal/feed"
)

type HttpHandler struct {
	cratesService *crates.Service
	feedService   *feed.Service
}

func NewHttpHandler(cratesService *crates.Service, feedService *feed.Service) *HttpHandler {
	return &HttpHandler{cratesService: cratesService, feedService: feedService}
}

// getFeeds fetches the user's feeds and converts them to the plain-value type
// used by core/templates header components. Returns empty slice on error so
// pages still render without a feed indicator.
func (h *HttpHandler) getFeeds(ctx contextx.ContextX) []templates.AppHeaderFeed {
	userID, err := ctx.UserId()
	if err != nil {
		return nil
	}
	feeds, err := h.feedService.GetUsersFeeds(ctx, userID)
	if err != nil {
		return nil
	}
	out := make([]templates.AppHeaderFeed, len(feeds))
	for i, f := range feeds {
		out[i] = templates.AppHeaderFeed{
			ID:       f.ID,
			Name:     string(f.Kind),
			Syncing:  f.LastSyncStatus.IsSyncing(),
			Failed:   f.LastSyncStatus.IsSyncFailed(),
			Unsynced: f.LastSyncStatus.IsUnsyned(),
			Stale:    f.IsSyncStale(),
			Synced:   f.LastSyncStatus.IsSynced(),
		}
	}
	return out
}

// GetCratesPage serves GET /app/crates.
func (h *HttpHandler) GetCratesPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	crateList, err := h.cratesService.ListCrates(ctx)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to list crates: %w", err),
		})
		return
	}

	if err := views.CratesPage(crateList, h.getFeeds(ctx)).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render crates page: %w", err),
		})
	}
}

// GetNewCrateModal serves GET /app/crates/new-modal.
func (h *HttpHandler) GetNewCrateModal(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	if err := views.CreateCrateModalFrag().Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render create crate modal: %w", err),
		})
	}
}

// GetCrateDetailPage serves GET /app/crates/{id}.
func (h *HttpHandler) GetCrateDetailPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	id := r.PathValue("id")

	crate, err := h.cratesService.GetCrate(ctx, id)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("failed to get crate: %w", err),
		})
		return
	}

	if err := views.CrateDetailPage(crate, h.getFeeds(ctx)).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render crate detail page: %w", err),
		})
	}
}

// CreateCrate serves POST /app/crates.
func (h *HttpHandler) CreateCrate(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to parse form: %w", err),
		})
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusUnprocessableEntity,
			Err:    errors.New("crate name cannot be empty"),
		})
		return
	}

	_, err := h.cratesService.CreateCrate(ctx, name)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to create crate: %w", err),
		})
		return
	}

	crateList, err := h.cratesService.ListCrates(ctx)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to list crates after create: %w", err),
		})
		return
	}

	if err := views.CratesListOOBFrag(crateList).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render crates list OOB: %w", err),
		})
		return
	}

	if err := templates.ForceCloseModal("create-crate-modal").Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render force close modal: %w", err),
		})
	}
}

// DeleteCrate serves DELETE /app/crates/{id}.
func (h *HttpHandler) DeleteCrate(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	id := r.PathValue("id")

	if err := h.cratesService.DeleteCrate(ctx, id); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to delete crate: %w", err),
		})
		return
	}

	w.Header().Set("HX-Redirect", "/app/crates")
	w.WriteHeader(http.StatusOK)
}

// GetCrateMembers serves GET /app/crates/{id}/members.
func (h *HttpHandler) GetCrateMembers(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	id := r.PathValue("id")

	crate, err := h.cratesService.GetCrate(ctx, id)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("failed to get crate: %w", err),
		})
		return
	}

	if err := views.CrateMembersFrag(crate.Albums, id).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render crate members: %w", err),
		})
	}
}

// GetEditCrateModal serves GET /app/crates/{id}/edit-modal.
func (h *HttpHandler) GetEditCrateModal(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	id := r.PathValue("id")

	crate, err := h.cratesService.GetCrate(ctx, id)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get crate: %w", err),
		})
		return
	}

	if err := views.EditCrateModalFrag(crate, id).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render edit crate modal: %w", err),
		})
	}
}

// SearchNonMembers serves GET /app/crates/{id}/edit-modal/search.
func (h *HttpHandler) SearchNonMembers(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	id := r.PathValue("id")
	q := r.URL.Query().Get("q")

	albums, err := h.cratesService.SearchNonMembers(ctx, id, q)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to search non-members: %w", err),
		})
		return
	}

	if err := views.NonMemberSearchResultsFrag(albums, id).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render non-member search results: %w", err),
		})
	}
}

// AddAlbum serves POST /app/crates/{id}/albums/{albumId}.
func (h *HttpHandler) AddAlbum(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	id := r.PathValue("id")
	albumID := r.PathValue("albumId")

	if err := h.cratesService.AddAlbum(ctx, id, albumID); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to add album to crate: %w", err),
		})
		return
	}

	h.renderMembershipUpdate(ctx, w, r, id)
}

// RemoveAlbum serves DELETE /app/crates/{id}/albums/{albumId}.
func (h *HttpHandler) RemoveAlbum(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	id := r.PathValue("id")
	albumID := r.PathValue("albumId")

	if err := h.cratesService.RemoveAlbum(ctx, id, albumID); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to remove album from crate: %w", err),
		})
		return
	}

	h.renderMembershipUpdate(ctx, w, r, id)
}

// renderMembershipUpdate is shared by AddAlbum and RemoveAlbum. It dispatches
// crateUpdated and OOB-swaps only the members list and search results, leaving
// the search input untouched so its value is preserved across add/remove actions.
func (h *HttpHandler) renderMembershipUpdate(ctx contextx.ContextX, w http.ResponseWriter, r *http.Request, crateID string) {
	crate, err := h.cratesService.GetCrate(ctx, crateID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get crate after membership change: %w", err),
		})
		return
	}

	q := r.FormValue("q")
	searchResults, err := h.cratesService.SearchNonMembers(ctx, crateID, q)
	if err != nil {
		searchResults = nil
	}

	if err := httpx.SetHXTrigger(w, "crateUpdated", nil); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	if err := views.EditCrateMembersOOBFrag(crate.Albums, crateID).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render members OOB: %w", err),
		})
		return
	}

	if err := views.NonMemberSearchResultsOOBFrag(searchResults, crateID).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render search results OOB: %w", err),
		})
	}
}
