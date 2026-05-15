package adapters

import (
	"github.com/alecdray/wax/src/internal/core/httpx"
)

// RegisterRoutes mounts the public auth routes (`/`, `/logout`,
// `/unauthorized`, `/spotify/callback`) on the given mux. The mux is
// expected to be the root mux — these routes sit outside JWT middleware.
func RegisterRoutes(mux *httpx.Mux, h *HttpHandler) {
	mux.Handle("GET /{$}", httpx.HandlerFunc(h.GetLoginPage))
	mux.Handle("GET /logout", httpx.HandlerFunc(h.Logout))
	mux.Handle("GET /unauthorized", httpx.HandlerFunc(h.GetUnauthorizedPage))
	mux.Handle("GET /spotify/callback", httpx.HandlerFunc(h.AuthorizeSpotify))
}
