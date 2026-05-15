package httpx

import (
	"fmt"
	"log/slog"
	"net/http"
	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/spotify"
	"github.com/alecdray/wax/src/internal/user"
	"time"
)

type Middleware func(HandlerFunc) HandlerFunc

func ApplyMiddleware(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func JwtMiddleware(spotifyService *spotify.Service, userService *user.Service) Middleware {
	errPrefix := "JWT middleware error:"

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := contextx.NewContextX(r.Context())
			a, err := ctx.App()
			if err != nil {
				HandleErrorResponse(ctx, w, HandleErrorResponseProps{
					Status: http.StatusInternalServerError,
					Err:    fmt.Errorf("%s failed to get app: %w", errPrefix, err),
				})
				return
			}

			claims, err := app.ValidateClaimsFromRequest(r, a.Config().JwtSecret)
			if err != nil {
				a.DeleteClaims(w)
			} else if claims == nil || claims.UserID == nil {
				err = fmt.Errorf("missing user ID")
			}

			if err != nil {
				err = fmt.Errorf("%s Invalid or expired claims: %s", errPrefix, err.Error())
				HandleUnauthorized(ctx, w, r, err)
				return
			}

			if claims.UserID != nil {
				ctx = ctx.WithUserId(*claims.UserID)
			}

			// Add claims to request context
			err = a.SetClaims(w, claims)
			if err != nil {
				HandleErrorResponse(ctx, w, HandleErrorResponseProps{
					Status: http.StatusInternalServerError,
					Err:    fmt.Errorf("%s failed to set JWT: %w", errPrefix, err),
				})
				return
			}
			ctx = ctx.WithApp(a)

			user, err := userService.GetUserFromCtx(ctx)
			if err != nil {
				err = fmt.Errorf("%s failed to get user: %w", errPrefix, err)
				HandleUnauthorized(ctx, w, r, err)
				return
			}
			ctx = ctx.WithUserId(user.ID)

			if user.SpotifyRefreshToken(a.Config().SpotifyTokenSecret) == nil {
				err = fmt.Errorf("%s missing Spotify refresh token", errPrefix)
				HandleUnauthorized(ctx, w, r, err)
				return
			}

			claims.UserID = &user.ID
			err = a.SetClaims(w, claims)
			if err != nil {
				HandleErrorResponse(ctx, w, HandleErrorResponseProps{
					Status: http.StatusInternalServerError,
					Err:    fmt.Errorf("%s failed to set JWT: %w", errPrefix, err),
				})
				return
			}
			ctx = ctx.WithApp(a)

			r = r.WithContext(ctx)
			// Call next handler
			next(w, r)
		}
	}
}

type RequestLoggingMiddlewareResponseWriter struct {
	http.ResponseWriter
	statusCode int
	startTime  time.Time
}

func (w *RequestLoggingMiddlewareResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *RequestLoggingMiddlewareResponseWriter) Duration() time.Duration {
	return time.Since(w.startTime)
}

func RequestLoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ww := &RequestLoggingMiddlewareResponseWriter{ResponseWriter: w, statusCode: 200, startTime: time.Now()}
		next(ww, r)
		slog.InfoContext(r.Context(), "Request", "status", ww.statusCode, "method", r.Method, "path", r.URL.Path, "url", r.URL.String(), "duration", ww.Duration())
	}
}
