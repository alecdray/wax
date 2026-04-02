package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/labels"
	"github.com/alecdray/wax/src/internal/library"
	libraryAdapters "github.com/alecdray/wax/src/internal/library/adapters"
)

type HttpHandler struct {
	libraryService *library.Service
	labelsService  *labels.Service
	discogsService *discogs.Service
}

func NewHttpHandler(libraryService *library.Service, labelsService *labels.Service, discogsService *discogs.Service) *HttpHandler {
	return &HttpHandler{
		libraryService: libraryService,
		labelsService:  labelsService,
		discogsService: discogsService,
	}
}

func (h *HttpHandler) SearchGenres(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	results := h.labelsService.SearchGenres(q)

	type suggestion struct {
		ID          string `json:"id"`
		Label       string `json:"label"`
		ParentLabel string `json:"parentLabel,omitempty"`
	}
	out := make([]suggestion, 0, len(results))
	for _, s := range results {
		out = append(out, suggestion{ID: s.ID, Label: s.Label, ParentLabel: s.ParentLabel})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *HttpHandler) GetLabelsModal(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return
	}

	artist := ""
	if len(album.Artists) > 0 {
		artist = album.Artists[0].Name
	}
	suggestions := h.discogsService.GetAlbumGenreSuggestions(ctx, album.Title, artist)
	genreSuggestions := make([]labels.GenreDTO, 0, len(suggestions))
	for _, n := range suggestions {
		genreSuggestions = append(genreSuggestions, labels.GenreDTO{ID: n.ID, Label: n.Label})
	}

	allMoods, err := h.labelsService.GetDistinctUserMoods(ctx, userID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get moods: %w", err),
		})
		return
	}

	allUserTags, err := h.labelsService.GetDistinctUserTags(ctx, userID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get user tags: %w", err),
		})
		return
	}

	if err := LabelsModal(*album, genreSuggestions, allMoods, allUserTags).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render labels modal: %w", err),
		})
	}
}

func (h *HttpHandler) SubmitAlbumLabels(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	genreIDs := r.Form["genre[]"]
	moods := r.Form["mood[]"]
	userTags := r.Form["tag[]"]

	newGenres, err := h.labelsService.SetAlbumGenres(ctx, userID, albumID, genreIDs)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to set album genres: %w", err),
		})
		return
	}

	newMoods, err := h.labelsService.SetAlbumMoods(ctx, userID, albumID, moods)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to set album moods: %w", err),
		})
		return
	}

	newUserTags, err := h.labelsService.SetAlbumUserTags(ctx, userID, albumID, userTags)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to set album user tags: %w", err),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return
	}
	// Use freshly computed labels to avoid extra DB round-trip.
	album.Genres = newGenres
	album.Moods = newMoods
	album.UserTags = newUserTags

	if err := CloseLabelsModal().Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	if err := libraryAdapters.AlbumTagsCell(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	if err := libraryAdapters.AlbumRowTagsSection(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
	}
}
