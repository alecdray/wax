package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"

	"github.com/alecdray/wax/src/internal/spotify"
)

// unauthorizedRedirectPath is the public auth route that renders the
// "Unauthorized" page. Owned by the `auth` module; referenced here as a
// URL string so `core/httpx` doesn't depend on the auth package.
const unauthorizedRedirectPath = "/unauthorized"

type ErrorResponseKind int

const (
	ErrorResponseKindNone ErrorResponseKind = iota
	ErrorResponseKindJSON
	ErrorResponseKindComponent
)

type ErrorResponse struct {
	json      *json.RawMessage
	component templ.Component
}

func NewErrorResponse() *ErrorResponse {
	return &ErrorResponse{}
}

func (e ErrorResponse) Kind() ErrorResponseKind {
	if e.JSON() != nil {
		return ErrorResponseKindJSON
	}
	if e.Component() != nil {
		return ErrorResponseKindComponent
	}
	return ErrorResponseKindNone
}

func (e *ErrorResponse) SetJSON(json *json.RawMessage) *ErrorResponse {
	e.json = json
	return e
}

func (e ErrorResponse) JSON() *json.RawMessage {
	return e.json
}

func (e *ErrorResponse) SetComponent(component templ.Component) *ErrorResponse {
	e.component = component
	return e
}

func (e ErrorResponse) Component() templ.Component {
	return e.component
}

type HandleErrorResponseProps struct {
	Status   int
	Err      error
	Response ErrorResponse
}

func HandleErrorResponse(ctx context.Context, w http.ResponseWriter, props HandleErrorResponseProps) {
	if props.Status == 0 {
		props.Status = http.StatusInternalServerError
	}

	// A refusal from the shared Spotify rate-limit guard is transient and
	// user-actionable, not a server fault: surface it as 429 + Retry-After so a
	// user-initiated action fails fast with a clear signal rather than a generic
	// 500. Background syncs never reach here — they back off via the feed. ADR 0006.
	var rateLimited *spotify.ErrRateLimited
	if errors.As(props.Err, &rateLimited) {
		props.Status = http.StatusTooManyRequests
		if secs := int(rateLimited.RetryAfter.Round(time.Second).Seconds()); secs > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(secs))
		}
	}

	w.WriteHeader(props.Status)
	slog.ErrorContext(ctx, "http error", "error", props.Err, "status", props.Status)

	switch props.Response.Kind() {
	case ErrorResponseKindJSON:
		json.NewEncoder(w).Encode(props.Response.JSON())
	case ErrorResponseKindComponent:
		props.Response.Component().Render(ctx, w)
	case ErrorResponseKindNone:
		// Do nothing
	}
}

// HandleUnauthorized logs the auth failure and redirects the client to
// the dedicated unauthorized page. For HTMX requests it emits an
// `HX-Redirect` header so HTMX performs a full-page navigation rather
// than swapping a fragment; for plain browser navigation it issues a
// 303 See Other.
func HandleUnauthorized(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(ctx, "http unauthorized", "error", err)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", unauthorizedRedirectPath)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, unauthorizedRedirectPath, http.StatusSeeOther)
}
