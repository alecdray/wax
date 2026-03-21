package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sort"

	"github.com/alecdray/wax/src/internal/core/app"
	"github.com/alecdray/wax/src/internal/core/contextx"
	"github.com/alecdray/wax/src/internal/core/db"
	"github.com/alecdray/wax/src/internal/discogs"
	"github.com/alecdray/wax/src/internal/genres"
)

const (
	limit      = 50
	outputFile = "tmp/discogs_records.json"
	stylesFile = "tmp/discogs_styles.json"
)

type ResolvedGenre struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Record struct {
	AlbumID      string          `json:"album_id"`
	AlbumTitle   string          `json:"album_title"`
	DiscogsID    int             `json:"discogs_id"`
	DiscogsTitle string          `json:"discogs_title"`
	Artists      []string        `json:"artists"`
	Year         int             `json:"year"`
	Genres       []ResolvedGenre `json:"genres"`
	Styles       []ResolvedGenre `json:"styles"`
	ResourceURL  string          `json:"resource_url"`
	MainRelease  int             `json:"main_release"`
}

func nodesToResolved(nodes []*genres.Node) []ResolvedGenre {
	out := make([]ResolvedGenre, len(nodes))
	for i, n := range nodes {
		out[i] = ResolvedGenre{ID: n.ID, Label: n.Label}
	}
	return out
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	dag, err := genres.Load()
	if err != nil {
		slog.Error("Failed to load genre DAG", "error", err)
		os.Exit(1)
	}

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
	seenStyles := make(map[string]struct{})

	for _, album := range albums {
		artistRows, err := database.Queries().GetAlbumArtistByAlbumId(ctx, album.ID)
		if err != nil || len(artistRows) == 0 {
			slog.Warn("No artists found for album, skipping", "album", album.Title)
			continue
		}

		var matchedItem *discogs.SearchItem

		for _, artistRow := range artistRows {
			artistName := artistRow.Artist.Name
			slog.Debug("Searching Discogs masters", "album", album.Title, "artist", artistName)

			item, err := discogsSvc.SearchMasterByAlbum(ctx, album.Title, artistName)
			if err != nil {
				slog.Warn("Discogs master search failed", "album", album.Title, "artist", artistName, "error", err)
				continue
			}

			if item != nil {
				matchedItem = item
				break
			}

			slog.Debug("No master results, trying next artist", "album", album.Title, "artist", artistName)
		}

		if matchedItem == nil {
			slog.Debug("No master found, falling back to release search", "album", album.Title)
			for _, artistRow := range artistRows {
				artistName := artistRow.Artist.Name
				slog.Debug("Searching Discogs releases", "album", album.Title, "artist", artistName)

				item, err := discogsSvc.SearchReleaseByAlbum(ctx, album.Title, artistName)
				if err != nil {
					slog.Warn("Discogs release search failed", "album", album.Title, "artist", artistName, "error", err)
					continue
				}

				if item != nil {
					matchedItem = item
					break
				}

				slog.Debug("No release results, trying next artist", "album", album.Title, "artist", artistName)
			}
		}

		if matchedItem == nil {
			slog.Warn("No Discogs results for any artist", "album", album.Title)
			continue
		}

		var (
			discogsID      int
			discogsTitle   string
			artistNames    []string
			year           int
			resolvedGenres []ResolvedGenre
			resolvedStyles []ResolvedGenre
			resourceURL    string
			mainRelease    int
		)

		if matchedItem.MasterID != 0 {
			master, err := discogsSvc.GetMaster(ctx, matchedItem.MasterID)
			if err != nil {
				slog.Warn("Failed to fetch master", "album", album.Title, "master_id", matchedItem.MasterID, "error", err)
				continue
			}
			discogsID = master.ID
			discogsTitle = master.Title
			year = master.Year
			resolvedGenres = nodesToResolved(discogs.Resolve(dag, master.Genres))
			resolvedStyles = nodesToResolved(discogs.Resolve(dag, master.Styles))
			resourceURL = master.ResourceURL
			mainRelease = master.MainRelease
			for _, a := range master.Artists {
				artistNames = append(artistNames, a.Name)
			}
		} else {
			release, err := discogsSvc.GetRelease(ctx, matchedItem.ID)
			if err != nil {
				slog.Warn("Failed to fetch release", "album", album.Title, "release_id", matchedItem.ID, "error", err)
				continue
			}
			discogsID = release.ID
			discogsTitle = release.Title
			year = release.Year
			resolvedGenres = nodesToResolved(discogs.Resolve(dag, release.Genres))
			resolvedStyles = nodesToResolved(discogs.Resolve(dag, release.Styles))
			resourceURL = release.ResourceURL
			for _, a := range release.Artists {
				artistNames = append(artistNames, a.Name)
			}
		}

		records = append(records, Record{
			AlbumID:      album.ID,
			AlbumTitle:   album.Title,
			DiscogsID:    discogsID,
			DiscogsTitle: discogsTitle,
			Artists:      artistNames,
			Year:         year,
			Genres:       resolvedGenres,
			Styles:       resolvedStyles,
			ResourceURL:  resourceURL,
			MainRelease:  mainRelease,
		})

		slog.Info("Matched", "album", album.Title, "discogs", discogsTitle, "year", year)
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
	enc.SetEscapeHTML(false)
	if err := enc.Encode(records); err != nil {
		slog.Error("Failed to write output", "error", err)
		os.Exit(1)
	}

	styles := make([]string, 0, len(seenStyles))
	for s := range seenStyles {
		styles = append(styles, s)
	}
	sort.Strings(styles)

	sf, err := os.Create(stylesFile)
	if err != nil {
		slog.Error("Failed to create styles file", "error", err)
		os.Exit(1)
	}
	defer sf.Close()

	senc := json.NewEncoder(sf)
	senc.SetIndent("", "  ")
	senc.SetEscapeHTML(false)
	if err := senc.Encode(styles); err != nil {
		slog.Error("Failed to write styles", "error", err)
		os.Exit(1)
	}

	slog.Info("Done", "records", len(records), "styles", len(styles), "output", outputFile)
}
