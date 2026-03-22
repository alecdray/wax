package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/library/adapters"
	"github.com/alecdray/wax/src/internal/review"
	"strconv"
)

type HttpHandler struct {
	libraryService *library.Service
	reviewService  *review.Service
}

func NewHttpHandler(libraryService *library.Service, reviewService *review.Service) *HttpHandler {
	return &HttpHandler{
		libraryService: libraryService,
		reviewService:  reviewService,
	}
}

func (h *HttpHandler) GetRatingRecommender(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		err = fmt.Errorf("failed to get user ID: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	query := r.URL.Query()
	albumId := query.Get("albumId")
	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		err = fmt.Errorf("failed to get album: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	var rating *float64
	if query.Has("rating") {
		queryRating := query.Get("rating")
		parsedRating, err := strconv.ParseFloat(queryRating, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse rating query: %w", err)
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    err,
			})
			return
		}
		rating = &parsedRating
	}

	err = RatingModal(*album, RatingModalProps{
		Rating: rating,
	}).Render(ctx, w)
	if err != nil {
		err = fmt.Errorf("failed to render response: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	return
}

func (h *HttpHandler) SubmitRatingRecommenderQuestions(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		err = fmt.Errorf("failed to get user ID: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	query := r.URL.Query()
	albumId := query.Get("albumId")
	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		err = fmt.Errorf("failed to get album: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	err = r.ParseForm()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}
	formData := r.Form

	questionsWithValues := make(review.RatingQuestions, len(review.RatingRecommenderQuestions))
	for i, question := range review.RatingRecommenderQuestions {
		if !formData.Has(question.Key.String()) {
			err = fmt.Errorf("missing form value for key %s", question.Key.String())
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    err,
			})
			return
		}

		val, err := strconv.Atoi(formData.Get(question.Key.String()))
		if err != nil {
			err = fmt.Errorf("failed to parse form value: %w", err)
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    err,
			})
			return
		}

		questionsWithValues[i] = question.WithValue(val)
	}

	rating := questionsWithValues.Rating()

	err = RatingRecommenderConfirm(*album, &rating).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
}


func (h *HttpHandler) GetRatingRecommenderQuestions(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumId := r.URL.Query().Get("albumId")
	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	err := RatingRecommenderQuestions(albumId, review.RatingRecommenderQuestions).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to render response: %w", err),
		})
	}
}

func (h *HttpHandler) SubmitRatingRecommenderRating(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		err = fmt.Errorf("failed to get user ID: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	query := r.URL.Query()
	albumId := query.Get("albumId")

	if albumId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing album ID"),
		})
		return
	}

	err = r.ParseForm()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}
	formData := r.Form

	queryRating := formData.Get("rating")
	rating, err := strconv.ParseFloat(queryRating, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse rating query: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	note := formData.Get("note")
	if len(note) > 2000 {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("note exceeds 2000 character limit"),
		})
		return
	}

	albumRating, err := h.reviewService.AddRating(ctx, userId, albumId, rating, note)
	if err != nil {
		err = fmt.Errorf("failed to add rating: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userId, albumId)
	if err != nil {
		err = fmt.Errorf("failed to get album: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}
	album.Rating = albumRating

	err = CloseRatingModal().Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = adapters.AlbumListRating(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = adapters.AlbumRating(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = adapters.AlbumRatingHistory(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = adapters.AlbumRowTagsSection(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
}

func (h *HttpHandler) DeleteRatingLogEntry(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userId, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get user ID: %w", err),
		})
		return
	}

	entryId := r.PathValue("id")
	if entryId == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    errors.New("missing entry ID"),
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

	err = h.reviewService.DeleteRatingEntry(ctx, userId, entryId)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    fmt.Errorf("failed to delete rating entry: %w", err),
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

	err = adapters.AlbumListRating(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = adapters.AlbumRating(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = adapters.AlbumRatingHistory(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}

	err = adapters.AlbumRowTagsSection(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
}
