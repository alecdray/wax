package adapters

import (
	"github.com/alecdray/wax/src/internal/core/httpx"
)

// RegisterRoutes mounts all /app/tags/... routes on the given mux. The mux
// is expected to be the authenticated app sub-mux (JWT middleware applied).
func RegisterRoutes(mux *httpx.Mux, h *HttpHandler) {
	mux.Handle("GET /app/tags/album", httpx.HandlerFunc(h.GetTagsModal))
	mux.Handle("POST /app/tags/album", httpx.HandlerFunc(h.SubmitAlbumTags))
}
