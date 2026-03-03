package review

import (
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/httpx"
	"shmoopicks/src/internal/core/utils"
	"strconv"
)

type HttpHandler struct {
}

func NewHttpHandler() *HttpHandler {
	return &HttpHandler{}
}

func (h *HttpHandler) GetRatingRecommender(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	err := RatingRecommender(RatingRecommenderQuestions, nil).Render(ctx, w)
	if err != nil {
		err = fmt.Errorf("failed to render response: %w", err)
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusInternalServerError,
			Err:    err,
		})
		return
	}
}

func (h *HttpHandler) SubmitRatingRecommender(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	err := r.ParseForm()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    err,
		})
	}

	questionsWithValues := make(RatingQuestions, len(RatingRecommenderQuestions))
	for i, question := range RatingRecommenderQuestions {
		if !r.Form.Has(question.Key.String()) {
			err = fmt.Errorf("missing form value for key %s", question.Key.String())
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    err,
			})
			return
		}

		val, err := strconv.Atoi(r.Form.Get(question.Key.String()))
		if err != nil {
			err = fmt.Errorf("failed to parse form value: %w", err)
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    err,
			})
		}

		questionsWithValues[i] = question.WithValue(val)
	}

	err = RatingRecommenderForm(questionsWithValues, utils.NewPointer(questionsWithValues.Score())).Render(ctx, w)
	httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
		Status: http.StatusInternalServerError,
		Err:    err,
	})
	return
}
