package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecdray/wax/src/internal/auth"
	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/core/task"
	"github.com/alecdray/wax/src/internal/core/templates"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/feed"
	"github.com/alecdray/wax/src/internal/genres"
	"github.com/alecdray/wax/src/internal/library"
	libraryAdapters "github.com/alecdray/wax/src/internal/library/adapters"
	"github.com/alecdray/wax/src/internal/listeninghistory"
	"github.com/alecdray/wax/src/internal/musicbrainz"
	"github.com/alecdray/wax/src/internal/review"
	reviewAdapters "github.com/alecdray/wax/src/internal/review/adapters"
	"github.com/alecdray/wax/src/internal/spotify"
	"github.com/alecdray/wax/src/internal/tags"
	tagsAdapters "github.com/alecdray/wax/src/internal/tags/adapters"
	"github.com/alecdray/wax/src/internal/user"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type services struct {
	taskManager      *task.TaskManager
	user             *user.Service
	musicbrainz      *musicbrainz.Service
	discogs          *discogs.Service
	spotifyAuth      *spotify.AuthService
	spotify          *spotify.Service
	library          *library.Service
	feed             *feed.Service
	review           *review.Service
	listeningHistory *listeninghistory.Service
	tags             *tags.Service
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

	discogsClient, err := discogs.NewClient(app.Config().DiscogsKey, app.Config().DiscogsSecret, app.Config().AppName)
	if err != nil {
		slog.Error("Failed to create Discogs client", "error", err)
		os.Exit(1)
	}
	genreDAG, err := genres.Load()
	if err != nil {
		slog.Warn("Failed to load genre DAG; tag suggestions will be unavailable", "error", err)
	}
	s.discogs = discogs.NewService(discogsClient, genreDAG)

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

	s.tags = tags.NewService(db)

	s.library = library.NewService(db, s.listeningHistory, s.tags)

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

	templates.InitCSSVersion("static/public/main.css")

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
	appMux.Handle("GET /app/library/dashboard/albums-page", httpx.HandlerFunc(libraryHandler.GetAlbumsPage))
	appMux.Handle("GET /app/library/dashboard/carousel", httpx.HandlerFunc(libraryHandler.GetCarousel))
	appMux.Handle("GET /app/library/albums/{albumId}", httpx.HandlerFunc(libraryHandler.GetAlbumDetailPage))

	tagsHandler := tagsAdapters.NewHttpHandler(services.library, services.tags, services.discogs)
	appMux.Handle("GET /app/tags/album", httpx.HandlerFunc(tagsHandler.GetTagsModal))
	appMux.Handle("POST /app/tags/album", httpx.HandlerFunc(tagsHandler.SubmitAlbumTags))

	reviewHandler := reviewAdapters.NewHttpHandler(services.library, services.review)
	appMux.Handle("GET /app/review/rating-recommender", httpx.HandlerFunc(reviewHandler.GetRatingRecommender))
	appMux.Handle("GET /app/review/rating-recommender/questions", httpx.HandlerFunc(reviewHandler.GetRatingRecommenderQuestions))
	appMux.Handle("POST /app/review/rating-recommender/questions", httpx.HandlerFunc(reviewHandler.SubmitRatingRecommenderQuestions))
	appMux.Handle("POST /app/review/rating-recommender/rating", httpx.HandlerFunc(reviewHandler.SubmitRatingRecommenderRating))
	appMux.Handle("DELETE /app/review/rating-log/{id}", httpx.HandlerFunc(reviewHandler.DeleteRatingLogEntry))

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
