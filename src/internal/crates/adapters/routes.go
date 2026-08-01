package adapters

import (
	"github.com/alecdray/wax/src/internal/core/httpx"
)

// RegisterRoutes mounts all /app/crates/... routes on the given mux. The mux
// is expected to be the authenticated app sub-mux (JWT middleware applied).
func RegisterRoutes(mux *httpx.Mux, h *HttpHandler) {
	mux.Handle("GET /app/crates", httpx.HandlerFunc(h.GetCratesPage))
	mux.Handle("GET /app/crates/new-modal", httpx.HandlerFunc(h.GetNewCrateModal))
	mux.Handle("GET /app/crates/picker", httpx.HandlerFunc(h.GetCratePicker))
	mux.Handle("GET /app/crates/{id}", httpx.HandlerFunc(h.GetCrateDetailPage))
	mux.Handle("POST /app/crates", httpx.HandlerFunc(h.CreateCrate))
	mux.Handle("DELETE /app/crates/{id}", httpx.HandlerFunc(h.DeleteCrate))
	mux.Handle("GET /app/crates/{id}/members", httpx.HandlerFunc(h.GetCrateMembers))
	mux.Handle("GET /app/crates/{id}/edit-modal", httpx.HandlerFunc(h.GetEditCrateModal))
	mux.Handle("GET /app/crates/{id}/edit-modal/search", httpx.HandlerFunc(h.SearchNonMembers))
	mux.Handle("GET /app/crates/{id}/albums/{albumId}/actions-modal", httpx.HandlerFunc(h.GetCrateMemberActionsModal))
	mux.Handle("POST /app/crates/{id}/albums/{albumId}", httpx.HandlerFunc(h.AddAlbum))
	mux.Handle("DELETE /app/crates/{id}/albums/{albumId}", httpx.HandlerFunc(h.RemoveAlbum))
}
