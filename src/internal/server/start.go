package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	authAdapters "github.com/alecdray/wax/src/internal/auth/adapters"
	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/httpx"
	"github.com/alecdray/wax/src/internal/core/templates"
	libraryAdapters "github.com/alecdray/wax/src/internal/library/adapters"
	reviewAdapters "github.com/alecdray/wax/src/internal/review/adapters"
	tagsAdapters "github.com/alecdray/wax/src/internal/tags/adapters"
)

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

	authHandler := authAdapters.NewHttpHandler(services.auth)
	authAdapters.RegisterRoutes(rootMux, authHandler)

	appMux := httpx.NewMux(app, httpx.JwtMiddleware(services.spotify, services.user))
	rootMux.Use("/app/", appMux)

	libraryHandler := libraryAdapters.NewHttpHandler(
		services.spotifyAuth,
		services.musicbrainz,
		services.feed,
		services.library,
		services.taskManager,
		services.discogs,
		services.notes,
	)
	libraryAdapters.RegisterRoutes(appMux, libraryHandler)

	tagsHandler := tagsAdapters.NewHttpHandler(services.library, services.tags, services.discogs)
	tagsAdapters.RegisterRoutes(appMux, tagsHandler)

	reviewHandler := reviewAdapters.NewHttpHandler(services.library, services.review)
	reviewAdapters.RegisterRoutes(appMux, reviewHandler)

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
