package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db/models"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/core/templates"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/library"
)

// GetFormatsModal renders the formats management modal for an album.
func (h *HttpHandler) GetFormatsModal(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.PathValue("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to get album: %w", err)})
		return
	}

	formats, err := h.libraryService.GetAlbumFormats(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to get formats: %w", err)})
		return
	}

	defaultSearch := album.Title
	if len(album.Artists) > 0 {
		defaultSearch = album.Title + " " + album.Artists[0].Name
	}

	if err := FormatsModal(albumID, defaultSearch, formats).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// PutFormats saves format ownership and optional Discogs data.
func (h *HttpHandler) PutFormats(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.PathValue("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	physicalFormats := []models.ReleaseFormat{
		models.ReleaseFormatVinyl,
		models.ReleaseFormatCD,
		models.ReleaseFormatCassette,
	}

	inputs := make([]library.SaveFormatInput, 0, len(physicalFormats))
	for _, format := range physicalFormats {
		owned := r.FormValue(string(format)+"_owned") == "true"
		discogsID := r.FormValue(string(format) + "_discogs_id")
		label := r.FormValue(string(format) + "_label")
		releasedAtStr := r.FormValue(string(format) + "_released_at")

		input := library.SaveFormatInput{
			Format:    format,
			Owned:     owned,
			DiscogsID: discogsID,
			Label:     label,
		}
		if releasedAtStr != "" {
			t, err := time.Parse("2006-01-02", releasedAtStr)
			if err != nil {
				httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid released_at for %s: %w", format, err)})
				return
			}
			input.ReleasedAt = &t
		}
		inputs = append(inputs, input)
	}

	if err := h.libraryService.SaveAlbumFormats(ctx, userID, albumID, inputs); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to save formats: %w", err)})
		return
	}

	// Re-fetch releases for OOB swap
	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}

	// Update format icons on album detail page and close the modal (both OOB).
	if err := FormatsReleasesOOB(albumID, album.Releases).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := templates.ForceCloseModal(FormatsModalId).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

var physicalFormatSet = map[models.ReleaseFormat]bool{
	models.ReleaseFormatVinyl:    true,
	models.ReleaseFormatCD:       true,
	models.ReleaseFormatCassette: true,
}

// discogsFormatName maps our internal format to the Discogs format search parameter.
var discogsFormatName = map[models.ReleaseFormat]string{
	models.ReleaseFormatVinyl:    "Vinyl",
	models.ReleaseFormatCD:       "CD",
	models.ReleaseFormatCassette: "Cassette",
}

// GetDiscogsSearch searches Discogs by query string and returns an inline results fragment.
func (h *HttpHandler) GetDiscogsSearch(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.PathValue("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	format := models.ReleaseFormat(r.PathValue("format"))
	if !physicalFormatSet[format] {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid format: %s", format)})
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	results, err := h.discogsService.SearchReleasesForFormat(ctx, q, discogsFormatName[format])
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("discogs search failed: %w", err)})
		return
	}

	items := results.Results
	if len(items) > 10 {
		items = items[:10]
	}

	if err := DiscogsSearchResults(albumID, format, items).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// GetDiscogsRelease uses label and year from query params (forwarded from the search result)
// because discogs.Release has no Labels field.
func (h *HttpHandler) GetDiscogsRelease(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	format := models.ReleaseFormat(r.PathValue("format"))
	if !physicalFormatSet[format] {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid format: %s", format)})
		return
	}

	discogsIDStr := r.PathValue("discogsId")
	discogsID, err := strconv.Atoi(discogsIDStr)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid Discogs ID: %w", err)})
		return
	}

	// label and year are forwarded from the search result that the user clicked.
	// discogs.Release has no Labels field, so we rely on the params.
	label := r.URL.Query().Get("label")
	year := r.URL.Query().Get("year")

	// Fetch full release to get the exact Released date if available.
	release, err := h.discogsService.GetRelease(ctx, discogsID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to fetch Discogs release: %w", err)})
		return
	}

	// Prefer the most precise date available from the Release object.
	// release.Released is "YYYY-MM-DD" when the full date is known, or just "YYYY".
	// Ensure the result is always a valid YYYY-MM-DD so PutFormats can parse it.
	releasedDate := year
	if len(release.Released) >= 10 {
		releasedDate = release.Released // full YYYY-MM-DD
	} else if len(release.Released) >= 4 {
		releasedDate = release.Released[:4] + "-01-01"
	} else if release.Year > 0 {
		releasedDate = strconv.Itoa(release.Year) + "-01-01"
	}
	// Bare 4-digit year from the search result fallback — expand to a full date.
	if len(releasedDate) == 4 {
		releasedDate += "-01-01"
	}

	item := discogs.SearchItem{
		ID:    discogsID,
		Title: release.Title,
		Year:  releasedDate,
	}
	if label != "" {
		item.Label = []string{label}
	}

	if err := DiscogsReleaseDetails(format, item).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}
