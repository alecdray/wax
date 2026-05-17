package adapters

import (
	"github.com/alecdray/wax/src/internal/core/httpx"
)

// RegisterRoutes mounts all /app/review/... routes on the given mux. The mux
// is expected to be the authenticated app sub-mux (JWT middleware applied).
func RegisterRoutes(mux *httpx.Mux, h *HttpHandler) {
	mux.Handle("GET /app/review/rating-recommender", httpx.HandlerFunc(h.GetRatingRecommender))
	mux.Handle("GET /app/review/rating-recommender/questions", httpx.HandlerFunc(h.GetRatingRecommenderQuestions))
	mux.Handle("POST /app/review/rating-recommender/questions", httpx.HandlerFunc(h.SubmitRatingRecommenderQuestions))
	mux.Handle("POST /app/review/rating-recommender/rating", httpx.HandlerFunc(h.SubmitRatingRecommenderRating))
	mux.Handle("POST /app/review/rating-recommender/confirm", httpx.HandlerFunc(h.GetRatingConfirm))
	mux.Handle("DELETE /app/review/rating-log/{id}", httpx.HandlerFunc(h.DeleteRatingLogEntry))
}
