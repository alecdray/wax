package adapters

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alecdray/wax/src/internal/auth"
	"github.com/alecdray/wax/src/internal/auth/adapters/views"
	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/httpx"
)

type HttpHandler struct {
	service *auth.Service
}

func NewHttpHandler(service *auth.Service) *HttpHandler {
	return &HttpHandler{service: service}
}

func (h *HttpHandler) GetLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	a, err := ctx.App()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get app: %w", err),
		})
		return
	}

	claims, err := app.ValidateClaimsFromRequest(r, a.Config().JwtSecret)
	if err != nil {
		slog.DebugContext(ctx, fmt.Errorf("failed to validate claims: %w", err).Error())
	}

	redirect, err := h.service.LoginRedirect(ctx, claims)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
		return
	}

	loginPage := views.LoginPage(views.LoginPageProps{
		AuthURL: h.service.SpotifyAuthURL(a.Config().StateCode),
	})
	loginPage.Render(r.Context(), w)
}

func (h *HttpHandler) GetUnauthorizedPage(w http.ResponseWriter, r *http.Request) {
	views.UnauthorizedPage().Render(r.Context(), w)
}

func (h *HttpHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	a, err := ctx.App()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get app: %w", err),
		})
		return
	}

	a.DeleteClaims(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *HttpHandler) AuthorizeSpotify(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())
	a, err := ctx.App()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get app: %w", err),
		})
		return
	}

	userID, err := h.service.CompleteSpotifyLogin(ctx, r)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	claims := a.Claims()
	if claims == nil {
		claims = app.NewClaims()
	}
	claims.UserID = &userID
	if err := a.SetClaims(w, claims); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to update JWT with user ID: %w", err),
		})
		return
	}
	ctx = ctx.WithApp(a)

	// If this (re)auth was started by an action that needed a scope (e.g.
	// enabling the radar inbox), return the user to where they began.
	dest := "/"
	if returnTo := httpx.TakePostAuthRedirect(w, r); returnTo != "" {
		dest = returnTo
	}
	http.Redirect(w, r.WithContext(ctx), dest, http.StatusSeeOther)
}
