package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/library"
	libraryAdapters "github.com/alecdray/wax/src/internal/library/adapters"
	"github.com/alecdray/wax/src/internal/tags"
)

type HttpHandler struct {
	libraryService *library.Service
	tagsService    *tags.Service
	discogsService *discogs.Service
}

func NewHttpHandler(libraryService *library.Service, tagsService *tags.Service, discogsService *discogs.Service) *HttpHandler {
	return &HttpHandler{
		libraryService: libraryService,
		tagsService:    tagsService,
		discogsService: discogsService,
	}
}

func (h *HttpHandler) fetchGenreSuggestions(ctx contextx.ContextX, album *library.AlbumDTO) []string {
	artist := ""
	if len(album.Artists) > 0 {
		artist = album.Artists[0].Name
	}
	return h.discogsService.GetAlbumGenreSuggestions(ctx, album.Title, artist)
}

func (h *HttpHandler) GetTagsModal(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	albumId := r.URL.Query().Get("albumId")
	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return
	}

	// Ensure default groups exist on first use
	tagGroups, err := h.tagsService.GetOrCreateDefaultGroups(ctx, userId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to bootstrap tag groups: %w", err),
		})
		return
	}

	allTags, err := h.tagsService.GetUserTags(ctx, userId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to get user tags: %w", err),
		})
		return
	}

	suggestions := h.fetchGenreSuggestions(ctx, album)

	err = TagsModal(*album, allTags, tagGroups, suggestions).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render response: %w", err),
		})
	}
}

func (h *HttpHandler) SubmitAlbumTags(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	albumId := r.URL.Query().Get("albumId")
	if albumId == "" {
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

	// Parse tag[] form values: each value is "name|groupId"
	rawTags := r.Form["tag[]"]
	inputs := make([]tags.TagInput, 0, len(rawTags))
	for _, raw := range rawTags {
		parts := strings.SplitN(raw, "|", 2)
		name := strings.TrimSpace(parts[0])
		groupID := ""
		if len(parts) == 2 {
			groupID = strings.TrimSpace(parts[1])
		}
		if name != "" {
			inputs = append(inputs, tags.TagInput{Name: name, GroupID: groupID})
		}
	}

	newTags, err := h.tagsService.SetAlbumTags(ctx, userId, albumId, inputs)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to set album tags: %w", err),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return
	}
	// GetAlbumInLibrary already calls GetAlbumTags, but use our freshly computed tags
	// to avoid an extra DB round-trip.
	album.Tags = newTags

	err = CloseTagsModal().Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = libraryAdapters.AlbumTagsCell(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = libraryAdapters.AlbumRowTagsSection(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
	}
}
