package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/review/adapters/views"
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

func (h *HttpHandler) getAlbum(ctx contextx.ContextX, w http.ResponseWriter, userID, albumID string) (*library.AlbumDTO, bool) {
	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to get album: %w", err),
		})
		return nil, false
	}
	return album, true
}

// scoreEntryPrefill returns the score to pre-fill on the score-entry form for
// an album. The pre-fill comes from the most-recent rating-log entry; an album
// with no log entries opens the form empty.
func scoreEntryPrefill(album library.AlbumDTO) *float64 {
	if album.Rating != nil && album.Rating.Rating != nil {
		v := *album.Rating.Rating
		return &v
	}
	return nil
}

// GetRatingRecommender opens the rating modal directly to the score-entry
// form, regardless of the album's rating state. The score input is pre-filled
// from the most-recent rating-log entry when one exists, or left empty for an
// unrated album.
func (h *HttpHandler) GetRatingRecommender(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	album, ok := h.getAlbum(ctx, w, userID, albumID)
	if !ok {
		return
	}

	props := views.RatingModalProps{
		Album:  *album,
		Rating: scoreEntryPrefill(*album),
	}

	if err := views.RatingModalFrag(props).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// GetRatingRecommenderQuestions renders the questionnaire fragment. The
// optional priorRating query param carries the value currently sitting in the
// score-entry form so dismissing the questionnaire can restore it.
func (h *HttpHandler) GetRatingRecommenderQuestions(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	mode := review.RatingMode(r.URL.Query().Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	priorRating := optionalFloatParam(r.URL.Query().Get("priorRating"))

	if err := views.BaseQuestionsFormFrag(albumID, mode, review.AllBaseQuestions, priorRating).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

func (h *HttpHandler) SubmitRatingRecommenderQuestions(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.URL.Query().Get("albumId")
	mode := review.RatingMode(r.URL.Query().Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	questions := make(review.BaseQuestions, len(review.AllBaseQuestions))
	copy(questions, review.AllBaseQuestions)

	questionValues := make(map[string]string)
	for i, q := range questions {
		rawVal := r.Form.Get(string(q.Key))
		if rawVal == "" {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("missing value for question %s", q.Key),
			})
			return
		}
		val, err := strconv.Atoi(rawVal)
		if err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid value for %s: %w", q.Key, err)})
			return
		}
		questions[i] = q.WithValue(val)
		questionValues[string(q.Key)] = rawVal
	}

	finalScore := review.FinalScore(questions.Score())

	if review.DetectContradictions(questions, mode) {
		questionValues["mode"] = string(mode)
		questionValues["final_score"] = strconv.FormatFloat(finalScore, 'f', 2, 64)
		if err := views.ConfidenceInterstitialFrag(albumID, mode, finalScore, questionValues).Render(ctx, w); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		}
		return
	}

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	album, ok := h.getAlbum(ctx, w, userID, albumID)
	if !ok {
		return
	}

	if err := views.RatingConfirmFormFrag(*album, mode, &finalScore).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// GetRatingConfirm re-renders the score-entry form with the supplied score
// pre-filled. Used by the questionnaire's dismiss affordance (to restore the
// prior pre-fill) and by the contradiction interstitial's proceed action.
func (h *HttpHandler) GetRatingConfirm(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.URL.Query().Get("albumId")

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	mode := review.RatingMode(r.Form.Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	rawScore := r.Form.Get("final_score")
	var ratingPtr *float64
	if rawScore != "" {
		v, err := strconv.ParseFloat(rawScore, 64)
		if err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid final_score: %w", err)})
			return
		}
		ratingPtr = &v
	}

	album, ok := h.getAlbum(ctx, w, userID, albumID)
	if !ok {
		return
	}

	if err := views.RatingConfirmFormFrag(*album, mode, ratingPtr).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

// SubmitRatingRecommenderRating saves a new rating-log entry and sets the
// album's rating state to provisional — creating the state row on first save,
// and demoting a finalized album back to provisional. The resulting state is
// always provisional.
func (h *HttpHandler) SubmitRatingRecommenderRating(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	ratingVal, err := strconv.ParseFloat(r.Form.Get("rating"), 64)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid rating: %w", err)})
		return
	}

	note := r.Form.Get("note")
	if len(note) > 2000 {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("note exceeds 2000 character limit")})
		return
	}

	if _, _, err := h.reviewService.SaveRating(ctx, userID, albumID, ratingVal, note); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to save rating: %w", err)})
		return
	}

	h.renderRatingSaveResponse(ctx, w, albumID)
}

// SubmitRatingRecommenderFinalize saves the score from the score-entry form and
// sets the album's rating state to finalized, from any prior state.
func (h *HttpHandler) SubmitRatingRecommenderFinalize(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	ratingVal, err := strconv.ParseFloat(r.Form.Get("rating"), 64)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid rating: %w", err)})
		return
	}

	note := r.Form.Get("note")
	if len(note) > 2000 {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("note exceeds 2000 character limit")})
		return
	}

	if _, _, err := h.reviewService.FinalizeWithRating(ctx, userID, albumID, ratingVal, note); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to finalize: %w", err)})
		return
	}

	h.renderRatingSaveResponse(ctx, w, albumID)
}

// renderRatingSaveResponse closes the rating modal and broadcasts album-changed
// so library refreshes the surfaces that depend on the rating state.
func (h *HttpHandler) renderRatingSaveResponse(ctx contextx.ContextX, w http.ResponseWriter, albumID string) {
	if err := httpx.SetHXTrigger(w, "album-changed", map[string]string{"albumId": albumID, "scope": "rating"}); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := views.CloseRatingModalFrag().Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

func (h *HttpHandler) DeleteRatingLogEntry(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	userID, err := ctx.UserId()
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to get user ID: %w", err)})
		return
	}

	entryID := r.PathValue("id")
	if entryID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing entry ID")})
		return
	}

	albumID := r.URL.Query().Get("albumId")
	if albumID == "" {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: errors.New("missing album ID")})
		return
	}

	if err := h.reviewService.DeleteRatingEntry(ctx, userID, entryID); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to delete: %w", err)})
		return
	}

	if err := httpx.SetHXTrigger(w, "album-changed", map[string]string{"albumId": albumID, "scope": "rating"}); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	// No body: HX-Trigger fires album-changed; library refreshes the surfaces.
}

// optionalFloatParam parses a query-string float that may be absent or empty;
// returns nil when no value is supplied or the value fails to parse.
func optionalFloatParam(raw string) *float64 {
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &v
}
