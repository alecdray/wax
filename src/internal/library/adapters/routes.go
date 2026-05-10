package adapters

import (
	"github.com/alecdray/wax/src/internal/core/httpx"
)

// RegisterRoutes mounts all /app/library/... routes on the given mux. The mux
// is expected to be the authenticated app sub-mux (JWT middleware applied).
func RegisterRoutes(mux *httpx.Mux, h *HttpHandler) {
	mux.Handle("/app/library/dashboard", httpx.HandlerFunc(h.GetDashboardPage))
	mux.Handle("/app/library/dashboard/feeds-dropdown-content", httpx.HandlerFunc(h.GetFeedsDropdown))
	mux.Handle("GET /app/library/dashboard/stats", httpx.HandlerFunc(h.GetLibraryStats))
	mux.Handle("POST /app/library/dashboard/feeds/sync", httpx.HandlerFunc(h.TriggerFeedSync))
	mux.Handle("/app/library/dashboard/albums-table", httpx.HandlerFunc(h.GetAlbumsTable))
	mux.Handle("GET /app/library/dashboard/albums-page", httpx.HandlerFunc(h.GetAlbumsPage))
	mux.Handle("GET /app/library/dashboard/carousel", httpx.HandlerFunc(h.GetCarousel))

	mux.Handle("GET /app/library/discover", httpx.HandlerFunc(h.GetDiscoverPage))
	mux.Handle("GET /app/library/discover/search", httpx.HandlerFunc(h.GetDiscoverSearch))
	mux.Handle("GET /app/library/discover/radar", httpx.HandlerFunc(h.GetDiscoverRadar))
	mux.Handle("POST /app/library/discover/radar", httpx.HandlerFunc(h.PostDiscoverRadar))

	mux.Handle("GET /app/library/albums/{albumId}", httpx.HandlerFunc(h.GetAlbumDetailPage))
	mux.Handle("DELETE /app/library/albums/{albumId}", httpx.HandlerFunc(h.DeleteAlbum))
	mux.Handle("DELETE /app/library/albums/{albumId}/radar", httpx.HandlerFunc(h.DeleteAlbumRadar))
	mux.Handle("POST /app/library/albums/{albumId}/library", httpx.HandlerFunc(h.PostAlbumLibrary))
	mux.Handle("GET /app/library/albums/{albumId}/formats", httpx.HandlerFunc(h.GetFormatsModal))
	mux.Handle("PUT /app/library/albums/{albumId}/formats", httpx.HandlerFunc(h.PutFormats))
	mux.Handle("GET /app/library/albums/{albumId}/formats/{format}/discogs/search", httpx.HandlerFunc(h.GetDiscogsSearch))
	mux.Handle("GET /app/library/albums/{albumId}/formats/{format}/discogs/releases/{discogsId}", httpx.HandlerFunc(h.GetDiscogsRelease))
	mux.Handle("GET /app/library/albums/{albumId}/sleeve-notes/editor", httpx.HandlerFunc(h.GetSleeveNotesEditor))
	mux.Handle("GET /app/library/albums/{albumId}/sleeve-notes/view", httpx.HandlerFunc(h.GetSleeveNotesView))
	mux.Handle("PUT /app/library/albums/{albumId}/sleeve-notes", httpx.HandlerFunc(h.SaveSleeveNote))
}
