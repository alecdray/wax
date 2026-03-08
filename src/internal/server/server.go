package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"shmoopicks/src/internal/auth"
	"shmoopicks/src/internal/core/app"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/db"
	"shmoopicks/src/internal/core/httpx"
	"shmoopicks/src/internal/core/task"
	"shmoopicks/src/internal/core/templates"
	"shmoopicks/src/internal/feed"
	"shmoopicks/src/internal/library"
	libraryAdapters "shmoopicks/src/internal/library/adapters"
	"shmoopicks/src/internal/listeninghistory"
	"shmoopicks/src/internal/musicbrainz"
	"shmoopicks/src/internal/review"
	reviewAdapters "shmoopicks/src/internal/review/adapters"
	"shmoopicks/src/internal/spotify"
	"shmoopicks/src/internal/user"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type services struct {
	taskManager      *task.TaskManager
	user             *user.Service
	musicbrainz      *musicbrainz.Service
	spotifyAuth      *spotify.AuthService
	spotify          *spotify.Service
	library          *library.Service
	feed             *feed.Service
	review           *review.Service
	listeningHistory *listeninghistory.Service
}

func NewServices(app app.App, db *db.DB) *services {
	s := &services{}

	s.taskManager = task.NewTaskManager(db, slog.Default())

	mbClient, err := musicbrainz.NewClient(
		app.Config().AppName,
		app.Config().AppVersion,
		musicbrainz.WithContactEmail(app.Config().ContactEmail),
	)
	if err != nil {
		slog.Error("Failed to create MusicBrainz client", "error", err)
		os.Exit(1)
	}

	s.user = user.NewService(db)

	s.musicbrainz = musicbrainz.NewService(mbClient)

	s.spotifyAuth = spotify.NewAuthService(
		app.Config().SpotifyClientId,
		app.Config().SpotifyClientSecret,
		fmt.Sprintf("%s/spotify/callback", app.Config().Host),
		spotifyauth.ScopeUserLibraryRead,
		spotifyauth.ScopeUserReadRecentlyPlayed,
	)

	s.spotify = spotify.NewService(s.user, s.spotifyAuth)

	s.listeningHistory = listeninghistory.NewService(db, s.spotify)
	s.taskManager.RegisterCronTask(
		listeninghistory.NewSyncListeningHistoryTask(s.listeningHistory),
	)

	s.library = library.NewService(db, s.listeningHistory)

	s.feed = feed.NewService(db, s.spotify, s.library)
	s.taskManager.RegisterCronTask(
		feed.NewSyncStaleSpotifyFeedsTask(s.feed),
	)

	s.review = review.NewService(db)

	return s
}

func Start(ctx context.Context, app app.App) {
	db, err := db.NewDB(app.Config().DbPath)
	if err != nil {
		slog.Error("Failed to create database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	services := NewServices(app, db)
	services.taskManager.Start(contextx.NewContextX(ctx).WithApp(app))
	defer services.taskManager.Stop()

	rootMux := httpx.NewMux(app, httpx.RequestLoggingMiddleware)

	rootMux.Handle("/static/", httpx.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/public")))))

	authHandler := auth.NewHttpHandler(services.spotifyAuth, services.user, services.feed)
	rootMux.Handle("/{$}", httpx.HandlerFunc(authHandler.GetLoginPage))
	rootMux.Handle("/logout", httpx.HandlerFunc(authHandler.Logout))
	rootMux.Handle("/spotify/callback", httpx.HandlerFunc(authHandler.AuthorizeSpotify))

	appMux := httpx.NewMux(app, httpx.JwtMiddleware(services.spotify, services.user))
	rootMux.Use("/app/", appMux)

	libraryHandler := libraryAdapters.NewHttpHandler(
		services.spotifyAuth,
		services.musicbrainz,
		services.feed,
		services.library,
		services.taskManager,
	)
	appMux.Handle("/app/library/dashboard", httpx.HandlerFunc(libraryHandler.GetDashboardPage))
	appMux.Handle("/app/library/dashboard/feeds-dropdown-content", httpx.HandlerFunc(libraryHandler.GetFeedsDropdown))
	appMux.Handle("POST /app/library/dashboard/feeds/sync", httpx.HandlerFunc(libraryHandler.TriggerFeedSync))
	appMux.Handle("/app/library/dashboard/albums-table", httpx.HandlerFunc(libraryHandler.GetAlbumsTable))
	appMux.Handle("GET /app/library/dashboard/carousel", httpx.HandlerFunc(libraryHandler.GetCarousel))

	reviewHandler := reviewAdapters.NewHttpHandler(services.library, services.review)
	appMux.Handle("GET /app/review/rating-recommender", httpx.HandlerFunc(reviewHandler.GetRatingRecommender))
	appMux.Handle("GET /app/review/rating-recommender/questions", httpx.HandlerFunc(reviewHandler.GetRatingRecommenderQuestions))
	appMux.Handle("POST /app/review/rating-recommender/questions", httpx.HandlerFunc(reviewHandler.SubmitRatingRecommenderQuestions))
	appMux.Handle("POST /app/review/rating-recommender/rating", httpx.HandlerFunc(reviewHandler.SubmitRatingRecommenderRating))
	appMux.Handle("DELETE /app/review/rating-recommender/rating", httpx.HandlerFunc(reviewHandler.DeleteRatingRecommenderRating))
	appMux.Handle("GET /app/review/notes", httpx.HandlerFunc(reviewHandler.GetReviewNotes))
	appMux.Handle("POST /app/review/notes", httpx.HandlerFunc(reviewHandler.SubmitReviewNotes))

	// Not found handler, must be registered after all other handlers
	rootMux.HandleFunc("/", httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			templates.Redirect("/", 0)
			return
		}
	}))

	addr := fmt.Sprintf(":%s", app.Config().Port)
	slog.Info("Starting server", "addr", addr)
	err = http.ListenAndServe(addr, rootMux)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
