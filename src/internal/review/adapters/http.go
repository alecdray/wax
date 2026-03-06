package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/httpx"
	"shmoopicks/src/internal/library"
	"shmoopicks/src/internal/library/adapters"
	"shmoopicks/src/internal/review"
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

	queryRating := query.Get("rating")
	if queryRating != "" {
		rating, err := strconv.ParseFloat(queryRating, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse rating query: %w", err)
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    err,
			})
			return
		}

		err = RatingModal(*album, RatingModalProps{
			Rating: &rating,
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

	err = RatingModal(*album, RatingModalProps{
		Questions: &review.RatingRecommenderQuestions,
	}).Render(ctx, w)
	if err != nil {
		err = fmt.Errorf("failed to render response: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
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

	err = RatingRecommenderConfirm(*album, rating).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
}

func (h *HttpHandler) UpdateRatingRecommenderRating(w http.ResponseWriter, r *http.Request) {
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

	queryRating := formData.Get("rating")
	rating, err := strconv.ParseFloat(queryRating, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse rating query: %w", err)
		errText := "Invalid rating"

		if queryRating == "" {
			errText = "A rating is required"
		}

		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status:   http.StatusBadRequest,
			Err:      err,
			Response: *httpx.NewErrorResponse().SetComponent(RatingRecommenderConfirmError(errText)),
		})
		return
	}

	if rating < 0 || rating > 10 {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status:   http.StatusBadRequest,
			Err:      err,
			Response: *httpx.NewErrorResponse().SetComponent(RatingRecommenderConfirmError("Rating must be between 0 and 10")),
		})
		return
	}

	err = RatingRecommenderConfirm(*album, rating).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
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

	albumRating, err := h.reviewService.UpdateRating(ctx, userId, albumId, rating)
	if err != nil {
		err = fmt.Errorf("failed to update rating: %w", err)
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

	err = adapters.AlbumRating(*album, true).Render(ctx, w)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
}
