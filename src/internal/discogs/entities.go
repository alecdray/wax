package discogs

type Genre string

const (
	GenreRock             Genre = "Rock"
	GenreElectronic       Genre = "Electronic"
	GenrePop              Genre = "Pop"
	GenreFolkWorldCountry Genre = "Folk, World, & Country"
	GenreJazz             Genre = "Jazz"
	GenreFunkSoul         Genre = "Funk / Soul"
	GenreClassical        Genre = "Classical"
	GenreHipHop           Genre = "Hip Hop"
	GenreLatin            Genre = "Latin"
	GenreStageScreen      Genre = "Stage & Screen"
	GenreReggae           Genre = "Reggae"
	GenreBlues            Genre = "Blues"
	GenreNonMusic         Genre = "Non-Music"
	GenreChildrens        Genre = "Children's"
	GenreBrassMilitary    Genre = "Brass & Military"
)

type Style string

const (
	StyleAbstract         Style = "Abstract"
	StyleAcoustic         Style = "Acoustic"
	StyleAfrobeat         Style = "Afrobeat"
	StyleAltPop           Style = "Alt-Pop"
	StyleAlternativeRock  Style = "Alternative Rock"
	StyleAmbient          Style = "Ambient"
	StyleArtRock          Style = "Art Rock"
	StyleBallad           Style = "Ballad"
	StyleBaroquePop       Style = "Baroque Pop"
	StyleBassMusic        Style = "Bass Music"
	StyleBayouFunk        Style = "Bayou Funk"
	StyleBluesRock        Style = "Blues Rock"
	StyleBoomBap          Style = "Boom Bap"
	StyleBossaNova        Style = "Bossa Nova"
	StyleChiptune         Style = "Chiptune"
	StyleCloudRap         Style = "Cloud Rap"
	StyleConscious        Style = "Conscious"
	StyleContemporaryJazz Style = "Contemporary Jazz"
	StyleContemporaryRnB  Style = "Contemporary R&B"
	StyleCountry          Style = "Country"
	StyleCountryBlues     Style = "Country Blues"
	StyleDancePop         Style = "Dance-pop"
	StyleDeepHouse        Style = "Deep House"
	StyleDisco            Style = "Disco"
	StyleDowntempo        Style = "Downtempo"
	StyleDreamPop         Style = "Dream Pop"
	StyleDrumNBass        Style = "Drum n Bass"
	StyleDubstep          Style = "Dubstep"
	StyleElectro          Style = "Electro"
	StyleExperimental     Style = "Experimental"
	StyleFavelaFunk       Style = "Favela Funk"
	StyleFolk             Style = "Folk"
	StyleFolkRock         Style = "Folk Rock"
	StyleFunk             Style = "Funk"
	StyleFusion           Style = "Fusion"
	StyleFutureBass       Style = "Future Bass"
	StyleFutureJazz       Style = "Future Jazz"
	StyleGFunk            Style = "G-Funk"
	StyleGangsta          Style = "Gangsta"
	StyleGarageRock       Style = "Garage Rock"
	StyleGoGo             Style = "Go-Go"
	StyleGrime            Style = "Grime"
	StyleHardcoreHipHop   Style = "Hardcore Hip-Hop"
	StyleHipHop           Style = "Hip Hop"
	StyleHonkyTonk        Style = "Honky Tonk"
	StyleHouse            Style = "House"
	StyleHyperpop         Style = "Hyperpop"
	StyleIDM              Style = "IDM"
	StyleIndiePop         Style = "Indie Pop"
	StyleIndieRock        Style = "Indie Rock"
	StyleInstrumental     Style = "Instrumental"
	StyleJazzRock         Style = "Jazz-Rock"
	StyleJazzyHipHop      Style = "Jazzy Hip-Hop"
	StyleLatinJazz        Style = "Latin Jazz"
	StyleLeftfield        Style = "Leftfield"
	StyleNeoSoul          Style = "Neo Soul"
	StyleNuMetal          Style = "Nu Metal"
	StylePopPunk          Style = "Pop Punk"
	StylePopRap           Style = "Pop Rap"
	StylePopRock          Style = "Pop Rock"
	StylePostRock         Style = "Post Rock"
	StylePostPunk         Style = "Post-Punk"
	StylePsychedelic      Style = "Psychedelic"
	StylePsychedelicRock  Style = "Psychedelic Rock"
	StylePunk             Style = "Punk"
	StyleReggaeton        Style = "Reggaeton"
	StyleRhythmAndBlues   Style = "Rhythm & Blues"
	StyleRnBSwing         Style = "RnB/Swing"
	StyleSamba            Style = "Samba"
	StyleSka              Style = "Ska"
	StyleSlowcore         Style = "Slowcore"
	StyleSoftRock         Style = "Soft Rock"
	StyleSoul             Style = "Soul"
	StyleSoulJazz         Style = "Soul-Jazz"
	StyleSpaceRock        Style = "Space Rock"
	StyleSynthPop         Style = "Synth-pop"
	StyleTexasBlues       Style = "Texas Blues"
	StyleTrap             Style = "Trap"
	StyleTripHop          Style = "Trip Hop"
	StyleVocal            Style = "Vocal"
)

type Pagination struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	PerPage int `json:"per_page"`
	Items   int `json:"items"`
}

type SearchResult struct {
	Pagination Pagination   `json:"pagination"`
	Results    []SearchItem `json:"results"`
}

type SearchItem struct {
	ID          int      `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Year        string   `json:"year"`
	Country     string   `json:"country"`
	Format      []string `json:"format"`
	Label       []string `json:"label"`
	Genre       []string `json:"genre"`
	Style       []string `json:"style"`
	CoverImage  string   `json:"cover_image"`
	Thumb       string   `json:"thumb"`
	ResourceURL string   `json:"resource_url"`
	MasterID    int      `json:"master_id"`
}

type Artist struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ResourceURL string `json:"resource_url"`
	Role        string `json:"role,omitempty"`
	Anv         string `json:"anv,omitempty"`
	Join        string `json:"join,omitempty"`
}

type Track struct {
	Position string   `json:"position"`
	Type     string   `json:"type_"`
	Title    string   `json:"title"`
	Duration string   `json:"duration"`
	Artists  []Artist `json:"artists,omitempty"`
}

type Image struct {
	Type        string `json:"type"`
	URI         string `json:"uri"`
	ResourceURL string `json:"resource_url"`
	URI150      string `json:"uri150"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

type Release struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Year        int      `json:"year"`
	Country     string   `json:"country"`
	Released    string   `json:"released"`
	Notes       string   `json:"notes,omitempty"`
	Artists     []Artist `json:"artists"`
	Genres      []string `json:"genres"`
	Styles      []string `json:"styles"`
	Tracklist   []Track  `json:"tracklist"`
	Images      []Image  `json:"images,omitempty"`
	ResourceURL string   `json:"resource_url"`
	MasterID    int      `json:"master_id"`
	MasterURL   string   `json:"master_url"`
}

type Master struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Year        int      `json:"year"`
	Artists     []Artist `json:"artists"`
	Genres      []string `json:"genres"`
	Styles      []string `json:"styles"`
	Tracklist   []Track  `json:"tracklist"`
	Images      []Image  `json:"images,omitempty"`
	ResourceURL string   `json:"resource_url"`
	MainRelease int      `json:"main_release"`
}
