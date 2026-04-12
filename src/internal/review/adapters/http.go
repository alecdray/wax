package adapters

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/library"
	libAdapters "github.com/alecdray/wax/src/internal/library/adapters"
	"github.com/alecdray/wax/src/internal/review"
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

	props := RatingModalProps{Album: *album}

	if album.RatingState == nil {
		props.ContentType = RatingModalContentQuestions
		props.Mode = review.RatingModeProvisional
	} else if album.RatingState.State == review.RatingStateFinalized {
		props.ContentType = RatingModalContentConfirm
		props.Mode = review.RatingModeFinalized
		if album.Rating != nil {
			props.Rating = album.Rating.Rating
		}
	} else {
		props.ContentType = RatingModalContentReratePrompt
		props.RerateIsDue = album.RatingState.IsRerateDue()
		props.RerateIsStalled = album.RatingState.State == review.RatingStateStalled
	}

	if err := RatingModal(props).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

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

	if err := BaseQuestionsForm(albumID, mode, review.AllBaseQuestions).Render(ctx, w); err != nil {
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
		if mode == review.RatingModeProvisional && q.Key == review.QuestionReturnRate {
			continue
		}
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

	baseScore := questions.Score(mode)

	if err := ModifiersForm(albumID, mode, baseScore, questionValues).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

func (h *HttpHandler) SubmitModifiers(w http.ResponseWriter, r *http.Request) {
	ctx := contextx.NewContextX(r.Context())

	albumID := r.URL.Query().Get("albumId")

	if err := r.ParseForm(); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	mode := review.RatingMode(r.Form.Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
	}

	baseScore, err := strconv.ParseFloat(r.Form.Get("base_score"), 64)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid base_score: %w", err)})
		return
	}

	mods := make(review.Modifiers, len(review.AllModifiers))
	copy(mods, review.AllModifiers)
	for i, m := range mods {
		rawVal := r.Form.Get(string(m.Key))
		if rawVal == "" {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("missing value for modifier %s", m.Key)})
			return
		}
		val, err := strconv.Atoi(rawVal)
		if err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid value for modifier %s: %w", m.Key, err)})
			return
		}
		mods[i] = m.WithValue(val)
	}

	finalScore := review.FinalScore(baseScore, mods.Adjustment(), mode)

	questions := make(review.BaseQuestions, len(review.AllBaseQuestions))
	copy(questions, review.AllBaseQuestions)
	for i, q := range questions {
		rawVal := r.Form.Get(string(q.Key))
		if rawVal != "" {
			val, _ := strconv.Atoi(rawVal)
			questions[i] = q.WithValue(val)
		}
	}

	if review.DetectContradictions(questions, mods, baseScore, mode) {
		allValues := make(map[string]string)
		for _, q := range questions {
			k := string(q.Key)
			if v := r.Form.Get(k); v != "" {
				allValues[k] = v
			}
		}
		for _, m := range mods {
			k := string(m.Key)
			allValues[k] = r.Form.Get(k)
		}
		if err := ConfidenceInterstitial(albumID, mode, finalScore, allValues).Render(ctx, w); err != nil {
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

	if err := RatingConfirmForm(*album, mode, &finalScore).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

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

	finalScore, err := strconv.ParseFloat(r.Form.Get("final_score"), 64)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid final_score: %w", err)})
		return
	}

	album, ok := h.getAlbum(ctx, w, userID, albumID)
	if !ok {
		return
	}

	if err := RatingConfirmForm(*album, mode, &finalScore).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
	}
}

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

	mode := review.RatingMode(r.Form.Get("mode"))
	if mode != review.RatingModeProvisional && mode != review.RatingModeFinalized {
		mode = review.RatingModeProvisional
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

	logState := review.RatingStateProvisional
	if mode == review.RatingModeFinalized {
		logState = review.RatingStateFinalized
	}

	albumRating, err := h.reviewService.AddRating(ctx, userID, albumID, ratingVal, note, logState)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to add rating: %w", err)})
		return
	}
	_ = albumRating

	currentState, err := h.reviewService.GetRatingState(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}

	if mode == review.RatingModeFinalized {
		if currentState == nil {
			if _, err := h.reviewService.CreateRatingState(ctx, userID, albumID); err != nil {
				httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
				return
			}
		}
		if _, err := h.reviewService.FinalizeRating(ctx, userID, albumID, currentState); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
			return
		}
	} else if mode == review.RatingModeProvisional && currentState == nil {
		if _, err := h.reviewService.CreateRatingState(ctx, userID, albumID); err != nil {
			httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
			return
		}
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to get album: %w", err)})
		return
	}

	if err := CloseRatingModal().Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumListRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRatingHistory(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRowTagsSection(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
}

func (h *HttpHandler) SnoozeRating(w http.ResponseWriter, r *http.Request) {
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

	if _, err := h.reviewService.SnoozeRating(ctx, userID, albumID); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: fmt.Errorf("failed to snooze: %w", err)})
		return
	}

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: err})
		return
	}

	if err := CloseRatingModal().Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumListRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
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

	album, err := h.libraryService.GetAlbumInLibrary(ctx, userID, albumID)
	if err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusBadRequest, Err: fmt.Errorf("failed to get album: %w", err)})
		return
	}

	if err := libAdapters.AlbumListRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRating(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRatingHistory(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
	if err := libAdapters.AlbumRowTagsSection(*album, true).Render(ctx, w); err != nil {
		httpx.HandleErrorResponse(ctx, w, httpx.HandleErrorResponseProps{Status: http.StatusInternalServerError, Err: err})
		return
	}
}
