package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecdray/wax/src/internal/auth"
	appConfig "github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/core/task"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/feed"
	"github.com/alecdray/wax/src/internal/genregraph"
	"github.com/alecdray/wax/src/internal/library"
	"github.com/alecdray/wax/src/internal/listeninghistory"
	"github.com/alecdray/wax/src/internal/musicbrainz"
	"github.com/alecdray/wax/src/internal/genres"
	"github.com/alecdray/wax/src/internal/notes"
	"github.com/alecdray/wax/src/internal/review"
	"github.com/alecdray/wax/src/internal/spotify"
	"github.com/alecdray/wax/src/internal/tags"
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
	genres           *genres.Service
	notes            *notes.Service
	auth             *auth.Service
}

func NewServices(app appConfig.App, db *db.DB) *services {
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
	genreDAG, err := genregraph.Load()
	if err != nil {
		slog.Warn("Failed to load genre DAG; tag suggestions will be unavailable", "error", err)
	}
	s.discogs = discogs.NewService(discogsClient, genreDAG)

	s.spotifyAuth = spotify.NewAuthService(
		app.Config().SpotifyClientId,
		app.Config().SpotifyClientSecret,
		fmt.Sprintf("%s/spotify/callback", app.Config().Host),
		spotifyauth.ScopeUserLibraryRead,
		spotifyauth.ScopeUserLibraryModify,
		spotifyauth.ScopeUserReadRecentlyPlayed,
		spotifyauth.ScopePlaylistReadPrivate,
		spotifyauth.ScopePlaylistModifyPrivate,
	)

	s.spotify = spotify.NewService(s.user, s.spotifyAuth)

	s.listeningHistory = listeninghistory.NewService(db, s.spotify)

	s.tags = tags.NewService(db)

	s.genres = genres.NewService(db, s.discogs, genreDAG)

	s.notes = notes.NewService(db)

	s.review = review.NewService(db)

	s.library = library.NewService(db, s.spotify, s.listeningHistory, s.tags, s.genres, s.notes, s.review)

	s.feed = feed.NewService(db, s.spotify, s.library)

	s.auth = auth.NewService(s.spotifyAuth, s.user, s.feed)

	// Cron tasks poll the Spotify Web API. Every non-prod instance shares the
	// same Spotify app credentials and therefore the same rate-limit budget, so
	// only prod runs them — local/dev rely on manual (ad-hoc) syncs instead.
	if app.Config().Env == appConfig.EnvProd {
		s.taskManager.RegisterCronTask(
			listeninghistory.NewSyncListeningHistoryTask(s.listeningHistory),
		)
		s.taskManager.RegisterCronTask(
			feed.NewSyncStaleSpotifyFeedsTask(s.feed),
		)
		s.taskManager.RegisterCronTask(
			feed.NewSyncStaleSpotifyRadarFeedsTask(s.feed),
		)
		s.taskManager.RegisterCronTask(
			genres.NewEnrichGenresTask(s.genres, s.library),
		)
	}

	return s
}
