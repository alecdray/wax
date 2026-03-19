package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/discogs"
)

const (
	limit      = 50
	outputFile = "tmp/discogs_records.json"
)

type Record struct {
	AlbumID      string   `json:"album_id"`
	AlbumTitle   string   `json:"album_title"`
	DiscogsID    int      `json:"discogs_id"`
	DiscogsTitle string   `json:"discogs_title"`
	Artists      []string `json:"artists"`
	Year         int      `json:"year"`
	Genres       []string `json:"genres"`
	Styles       []string `json:"styles"`
	ResourceURL  string   `json:"resource_url"`
	MainRelease  int      `json:"main_release"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	config := app.LoadConfig()

	database, err := db.NewDB(config.DbPath)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	discogsClient, err := discogs.NewClient(
		config.DiscogsKey,
		config.DiscogsSecret,
		config.AppName+"/"+config.AppVersion+" +"+config.ContactEmail,
	)
	if err != nil {
		slog.Error("Failed to create Discogs client", "error", err)
		os.Exit(1)
	}
	discogsSvc := discogs.NewService(discogsClient)

	ctx := contextx.NewContextX(context.Background())

	albums, err := database.Queries().ListAlbums(ctx, int64(limit))
	if err != nil {
		slog.Error("Failed to list albums", "error", err)
		os.Exit(1)
	}

	slog.Info("Fetched albums", "count", len(albums))

	var records []Record

	for _, album := range albums {
		artistRows, err := database.Queries().GetAlbumArtistByAlbumId(ctx, album.ID)
		if err != nil || len(artistRows) == 0 {
			slog.Warn("No artists found for album, skipping", "album", album.Title)
			continue
		}

		var matchedItem *discogs.SearchItem

		for _, artistRow := range artistRows {
			artistName := artistRow.Artist.Name
			slog.Debug("Searching Discogs", "album", album.Title, "artist", artistName)

			item, err := discogsSvc.SearchMasterByAlbum(ctx, album.Title, artistName)
			if err != nil {
				slog.Warn("Discogs search failed", "album", album.Title, "artist", artistName, "error", err)
				continue
			}

			if item != nil {
				matchedItem = item
				break
			}

			slog.Debug("No results, trying next artist", "album", album.Title, "artist", artistName)
		}

		if matchedItem == nil {
			slog.Warn("No Discogs results for any artist", "album", album.Title)
			continue
		}

		if matchedItem.MasterID == 0 {
			slog.Warn("Search result has no master ID, skipping", "album", album.Title, "discogs_title", matchedItem.Title)
			continue
		}

		master, err := discogsSvc.GetMaster(ctx, matchedItem.MasterID)
		if err != nil {
			slog.Warn("Failed to fetch master", "album", album.Title, "master_id", matchedItem.MasterID, "error", err)
			continue
		}

		var artistNames []string
		for _, a := range master.Artists {
			artistNames = append(artistNames, a.Name)
		}

		records = append(records, Record{
			AlbumID:      album.ID,
			AlbumTitle:   album.Title,
			DiscogsID:    master.ID,
			DiscogsTitle: master.Title,
			Artists:      artistNames,
			Year:         master.Year,
			Genres:       master.Genres,
			Styles:       master.Styles,
			ResourceURL:  master.ResourceURL,
			MainRelease:  master.MainRelease,
		})

		slog.Info("Matched", "album", album.Title, "discogs", master.Title, "year", master.Year)
	}

	if err := os.MkdirAll("tmp", 0755); err != nil {
		slog.Error("Failed to create tmp dir", "error", err)
		os.Exit(1)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		slog.Error("Failed to create output file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(records); err != nil {
		slog.Error("Failed to write output", "error", err)
		os.Exit(1)
	}

	slog.Info("Done", "records", len(records), "output", outputFile)
}
