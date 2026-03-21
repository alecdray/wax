package discogs

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
