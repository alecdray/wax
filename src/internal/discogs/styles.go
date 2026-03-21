package discogs

import (
	"log/slog"

	"github.com/alecdray/wax/src/internal/tags"
)

// normalizedStyleMap is a normalized version of StyleMap for case-insensitive lookup.
var normalizedStyleMap = func() map[string]tags.Subgenre {
	m := make(map[string]tags.Subgenre, len(StyleMap))
	for k, v := range StyleMap {
		m[normalizeGenre(string(k))] = v
	}
	return m
}()

// ToSubgenres converts a slice of Discogs style strings to subgenre tags.
// Unrecognized styles map to tags.SubgenreUnknown.
func ToSubgenres(discogsStyles []string) []tags.Subgenre {
	seen := make(map[tags.Subgenre]struct{})
	var result []tags.Subgenre
	for _, s := range discogsStyles {
		key := normalizeGenre(s)
		mapped, ok := normalizedStyleMap[key]
		if !ok {
			slog.Warn("unrecognized discogs style", "style", s)
			mapped = tags.SubgenreUnknown
		}
		if _, exists := seen[mapped]; !exists {
			seen[mapped] = struct{}{}
			result = append(result, mapped)
		}
	}
	return result
}

// StyleMap maps a Discogs style to a subgenre tag.
var StyleMap = map[Style]tags.Subgenre{
	StyleAbstract:         tags.SubgenreAbstract,
	StyleAcoustic:         tags.SubgenreAcoustic,
	StyleAfrobeat:         tags.SubgenreAfrobeat,
	StyleAltPop:           tags.SubgenreAltPop,
	StyleAlternativeRock:  tags.SubgenreAlternativeRock,
	StyleAmbient:          tags.SubgenreAmbient,
	StyleArtRock:          tags.SubgenreArtRock,
	StyleBallad:           tags.SubgenreBallad,
	StyleBaroquePop:       tags.SubgenreBaroquePop,
	StyleBassMusic:        tags.SubgenreBassMusic,
	StyleBayouFunk:        tags.SubgenreBayouFunk,
	StyleBluesRock:        tags.SubgenreBluesRock,
	StyleBoomBap:          tags.SubgenreBoomBap,
	StyleBossaNova:        tags.SubgenreBossaNova,
	StyleChiptune:         tags.SubgenreChiptune,
	StyleCloudRap:         tags.SubgenreCloudRap,
	StyleConscious:        tags.SubgenreConscious,
	StyleContemporaryJazz: tags.SubgenreContemporaryJazz,
	StyleContemporaryRnB:  tags.SubgenreContemporaryRnB,
	StyleCountry:          tags.SubgenreCountry,
	StyleCountryBlues:     tags.SubgenreCountryBlues,
	StyleDancePop:         tags.SubgenreDancePop,
	StyleDeepHouse:        tags.SubgenreDeepHouse,
	StyleDisco:            tags.SubgenreDisco,
	StyleDowntempo:        tags.SubgenreDowntempo,
	StyleDreamPop:         tags.SubgenreDreamPop,
	StyleDrumNBass:        tags.SubgenreDrumNBass,
	StyleDubstep:          tags.SubgenreDubstep,
	StyleElectro:          tags.SubgenreElectro,
	StyleExperimental:     tags.SubgenreExperimental,
	StyleFavelaFunk:       tags.SubgenreFavelaFunk,
	StyleFolk:             tags.SubgenreFolk,
	StyleFolkRock:         tags.SubgenreFolkRock,
	StyleFunk:             tags.SubgenreFunk,
	StyleFusion:           tags.SubgenreFusion,
	StyleFutureBass:       tags.SubgenreFutureBass,
	StyleFutureJazz:       tags.SubgenreFutureJazz,
	StyleGFunk:            tags.SubgenreGFunk,
	StyleGangsta:          tags.SubgenreGangsta,
	StyleGarageRock:       tags.SubgenreGarageRock,
	StyleGoGo:             tags.SubgenreGoGo,
	StyleGrime:            tags.SubgenreGrime,
	StyleHardcoreHipHop:   tags.SubgenreHardcoreHipHop,
	StyleHipHop:           tags.SubgenreHipHop,
	StyleHonkyTonk:        tags.SubgenreHonkyTonk,
	StyleHouse:            tags.SubgenreHouse,
	StyleHyperpop:         tags.SubgenreHyperpop,
	StyleIDM:              tags.SubgenreIDM,
	StyleIndiePop:         tags.SubgenreIndiePop,
	StyleIndieRock:        tags.SubgenreIndieRock,
	StyleInstrumental:     tags.SubgenreInstrumental,
	StyleJazzRock:         tags.SubgenreJazzRock,
	StyleJazzyHipHop:      tags.SubgenreJazzyHipHop,
	StyleLatinJazz:        tags.SubgenreLatinJazz,
	StyleLeftfield:        tags.SubgenreLeftfield,
	StyleNeoSoul:          tags.SubgenreNeoSoul,
	StyleNuMetal:          tags.SubgenreNuMetal,
	StylePopPunk:          tags.SubgenrePopPunk,
	StylePopRap:           tags.SubgenrePopRap,
	StylePopRock:          tags.SubgenrePopRock,
	StylePostRock:         tags.SubgenrePostRock,
	StylePostPunk:         tags.SubgenrePostPunk,
	StylePsychedelic:      tags.SubgenrePsychedelic,
	StylePsychedelicRock:  tags.SubgenrePsychedelicRock,
	StylePunk:             tags.SubgenrePunk,
	StyleReggaeton:        tags.SubgenreReggaeton,
	StyleRhythmAndBlues:   tags.SubgenreRhythmAndBlues,
	StyleRnBSwing:         tags.SubgenreRnBSwing,
	StyleSamba:            tags.SubgenreSamba,
	StyleSka:              tags.SubgenreSka,
	StyleSlowcore:         tags.SubgenreSlowcore,
	StyleSoftRock:         tags.SubgenreSoftRock,
	StyleSoul:             tags.SubgenreSoul,
	StyleSoulJazz:         tags.SubgenreSoulJazz,
	StyleSpaceRock:        tags.SubgenreSpaceRock,
	StyleSynthPop:         tags.SubgenreSynthPop,
	StyleTexasBlues:       tags.SubgenreTexasBlues,
	StyleTrap:             tags.SubgenreTrap,
	StyleTripHop:          tags.SubgenreTripHop,
	StyleVocal:            tags.SubgenreVocal,
}
