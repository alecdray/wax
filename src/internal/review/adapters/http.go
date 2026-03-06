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

	queryScore := query.Get("score")
	if queryScore != "" {
		score, err := strconv.ParseFloat(queryScore, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse score query: %w", err)
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    err,
			})
			return
		}

		err = RatingModal(*album, RatingModalProps{
			Score: &score,
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

	score := questionsWithValues.Score()

	err = RatingRecommenderConfirm(*album, score).Render(ctx, w)
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

	queryScore := formData.Get("score")
	score, err := strconv.ParseFloat(queryScore, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse score query: %w", err)
		errText := "Invalid rating"

		if queryScore == "" {
			errText = "A rating is required"
		}

		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status:   http.StatusBadRequest,
			Err:      err,
			Response: *httpx.NewErrorResponse().SetComponent(RatingRecommenderConfirmError(errText)),
		})
		return
	}

	if score < 0 || score > 10 {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status:   http.StatusBadRequest,
			Err:      err,
			Response: *httpx.NewErrorResponse().SetComponent(RatingRecommenderConfirmError("Rating must be between 0 and 10")),
		})
		return
	}

	err = RatingRecommenderConfirm(*album, score).Render(ctx, w)
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

	queryScore := formData.Get("score")
	score, err := strconv.ParseFloat(queryScore, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse score query: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}

	rating, err := h.reviewService.UpdateRating(ctx, userId, albumId, score)
	if err != nil {
		err = fmt.Errorf("failed to update rating: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
		return
	}
	album.Rating = rating

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
