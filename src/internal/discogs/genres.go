package discogs

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/alecdray/wax/src/internal/tags"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9 ]+`)

func normalizeGenre(g string) string {
	g = strings.ToLower(strings.TrimSpace(g))
	return strings.TrimSpace(nonAlphanumeric.ReplaceAllString(g, ""))
}

// normalizedGenreMap is a normalized version of GenreMap for case-insensitive lookup.
var normalizedGenreMap = func() map[string][]tags.Genre {
	m := make(map[string][]tags.Genre, len(GenreMap))
	for k, v := range GenreMap {
		m[normalizeGenre(string(k))] = v
	}
	return m
}()

// ToMasterGenres converts a slice of Discogs genre strings to master genre tags.
// Unrecognized genres map to tags.GenreUnknown.
func ToMasterGenres(discogsGenres []string) []tags.Genre {
	seen := make(map[tags.Genre]struct{})
	var result []tags.Genre
	for _, g := range discogsGenres {
		key := normalizeGenre(g)
		mapped, ok := normalizedGenreMap[key]
		if !ok {
			slog.Warn("unrecognized discogs genre", "genre", g)
			mapped = []tags.Genre{tags.GenreUnknown}
		}
		for _, mg := range mapped {
			if _, exists := seen[mg]; !exists {
				seen[mg] = struct{}{}
				result = append(result, mg)
			}
		}
	}
	return result
}

// GenreMap maps a Discogs genre to one or more master genre tags.
var GenreMap = map[Genre][]tags.Genre{
	GenreRock:             {tags.GenreRock},
	GenreElectronic:       {tags.GenreElectronic},
	GenrePop:              {tags.GenrePop},
	GenreFolkWorldCountry: {tags.GenreFolk, tags.GenreWorld, tags.GenreCountry},
	GenreJazz:             {tags.GenreJazz},
	GenreFunkSoul:         {tags.GenreFunk, tags.GenreSoul},
	GenreClassical:        {tags.GenreClassical},
	GenreHipHop:           {tags.GenreHipHop},
	GenreLatin:            {tags.GenreLatin},
	GenreStageScreen:      {tags.GenreTheater},
	GenreReggae:           {tags.GenreReggae},
	GenreBlues:            {tags.GenreBlues},
	GenreNonMusic:         {tags.GenreOther},
	GenreChildrens:        {tags.GenreOther},
	GenreBrassMilitary:    {tags.GenreOther},
}
