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
)

type HttpHandler struct {
	cratesService *crates.Service
}

func NewHttpHandler(cratesService *crates.Service) *HttpHandler {
	return &HttpHandler{cratesService: cratesService}
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

	if err := views.CratesPage(crateList).Render(ctx, w); err != nil {
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

	if err := views.CrateDetailPage(crate).Render(ctx, w); err != nil {
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

	http.Redirect(w, r, "/app/crates", http.StatusSeeOther)
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

	crate, err := h.cratesService.GetCrate(ctx, id)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get crate after add: %w", err),
		})
		return
	}

	if err := httpx.SetHXTrigger(w, "crateUpdated", nil); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
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

	crate, err := h.cratesService.GetCrate(ctx, id)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get crate after remove: %w", err),
		})
		return
	}

	if err := httpx.SetHXTrigger(w, "crateUpdated", nil); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
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
