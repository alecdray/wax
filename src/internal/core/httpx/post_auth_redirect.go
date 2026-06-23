package httpx

import (
	"net/http"
	"strings"
)

const postAuthRedirectCookie = "wax_post_auth_redirect"

// SetPostAuthRedirect records a local path to return to once the next Spotify
// (re)authentication completes. Used when an action needs a scope the user must
// re-grant (e.g. enabling the radar inbox), so the OAuth round trip returns the
// user to where they started instead of the default landing page.
func SetPostAuthRedirect(w http.ResponseWriter, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     postAuthRedirectCookie,
		Value:    path,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// TakePostAuthRedirect returns the stored post-auth redirect path and clears the
// cookie. It returns "" when none is set or the path is not a safe local app
// path (must begin with "/app/" and not "//"), guarding against open redirects.
func TakePostAuthRedirect(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(postAuthRedirectCookie)
	if err != nil || cookie.Value == "" {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     postAuthRedirectCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	if !strings.HasPrefix(cookie.Value, "/app/") || strings.HasPrefix(cookie.Value, "//") {
		return ""
	}
	return cookie.Value
}
