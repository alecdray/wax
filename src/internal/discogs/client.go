package discogs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
)

const origin = "https://api.discogs.com"

const (
	rateLimitWindow = 60 * time.Second
	maxRetries      = 3
)

type rateLimit struct {
	mu          sync.Mutex
	remaining   int
	exhaustedAt time.Time
}

func (r *rateLimit) update(resp *http.Response) {
	remaining, err := strconv.Atoi(resp.Header.Get("X-Discogs-Ratelimit-Remaining"))
	if err != nil {
		return
	}
	r.mu.Lock()
	r.remaining = remaining
	if remaining > 0 {
		r.exhaustedAt = time.Time{}
	}
	r.mu.Unlock()
}

func (r *rateLimit) release() {
	r.mu.Lock()
	r.remaining++
	r.mu.Unlock()
}

func (r *rateLimit) throttle() {
	r.mu.Lock()
	r.remaining = 0
	if r.exhaustedAt.IsZero() {
		r.exhaustedAt = time.Now()
	}
	r.mu.Unlock()
}

// acquire claims one request slot. If none are available it waits one second
// and retries, so concurrent callers queue up rather than all firing at once.
// When the limit has been exhausted for a full 60-second window, one probe
// request is allowed through so that update() can restore the counter from
// the server's response.
func (r *rateLimit) acquire(ctx contextx.ContextX) error {
	for {
		r.mu.Lock()
		if r.remaining > 0 {
			r.remaining--
			r.mu.Unlock()
			return nil
		}
		// Rate limit window has reset on the server side; let one probe through.
		if !r.exhaustedAt.IsZero() && time.Since(r.exhaustedAt) >= rateLimitWindow {
			r.exhaustedAt = time.Time{}
			r.mu.Unlock()
			return nil
		}
		if r.exhaustedAt.IsZero() {
			r.exhaustedAt = time.Now()
		}
		r.mu.Unlock()

		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type Client struct {
	consumerKey    string
	consumerSecret string
	userAgent      string
	rateLimit      rateLimit
}

func NewClient(consumerKey, consumerSecret, userAgent string) (*Client, error) {
	if consumerKey == "" {
		return nil, errors.New("consumerKey cannot be empty")
	}
	if consumerSecret == "" {
		return nil, errors.New("consumerSecret cannot be empty")
	}
	if userAgent == "" {
		return nil, errors.New("userAgent cannot be empty")
	}
	return &Client{
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		userAgent:      userAgent,
		rateLimit:      rateLimit{remaining: 60},
	}, nil
}

func (c *Client) makeRequest(ctx contextx.ContextX, method, path string, query url.Values) (*http.Response, error) {
	reqURL, err := url.Parse(origin + path)
	if err != nil {
		return nil, err
	}
	reqURL.RawQuery = query.Encode()

	for attempt := range maxRetries {
		if err := c.rateLimit.acquire(ctx); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Discogs key=%s, secret=%s", c.consumerKey, c.consumerSecret))
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.rateLimit.release()
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			c.rateLimit.throttle()
			slog.Debug("discogs rate limited, retrying", "attempt", attempt+1)
			continue
		}

		c.rateLimit.update(resp)
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: rate limited", maxRetries)
}

type SearchType string

const (
	SearchTypeRelease SearchType = "release"
	SearchTypeMaster  SearchType = "master"
	SearchTypeArtist  SearchType = "artist"
	SearchTypeLabel   SearchType = "label"
)

type PageProps struct {
	// Page is the page number to fetch (default 1).
	Page int
	// PerPage is the number of items per page, up to 100 (default 50).
	PerPage int
}

type SearchProps struct {
	// Query is a general search query (e.g. "nirvana").
	Query string
	// Type filters results to a specific entity type.
	Type SearchType
	// Title searches by combined "Artist Name - Release Title" title field.
	Title string
	// ReleaseTitle searches release titles.
	ReleaseTitle string
	// Credit searches release credits.
	Credit string
	// Artist searches artist names.
	Artist string
	// Anv searches artist ANV (artist name variation).
	Anv string
	// Label searches label names.
	Label string
	// Genre searches genres.
	Genre string
	// Style searches styles.
	Style string
	// Country searches release country.
	Country string
	// Year searches release year.
	Year string
	// Format searches formats (e.g. "album").
	Format string
	// Catno searches catalog numbers.
	Catno string
	// Barcode searches barcodes.
	Barcode string
	// Track searches track titles.
	Track string
	// Submitter searches by submitter username.
	Submitter string
	// Contributor searches by contributor username.
	Contributor string
	Page        PageProps
}

func (c *Client) SearchDatabase(ctx contextx.ContextX, props SearchProps) (*SearchResult, error) {
	query := url.Values{}
	setIfNonEmpty := func(key, val string) {
		if val != "" {
			query.Set(key, val)
		}
	}
	setIfNonEmpty("q", props.Query)
	setIfNonEmpty("type", string(props.Type))
	setIfNonEmpty("title", props.Title)
	setIfNonEmpty("release_title", props.ReleaseTitle)
	setIfNonEmpty("credit", props.Credit)
	setIfNonEmpty("artist", props.Artist)
	setIfNonEmpty("anv", props.Anv)
	setIfNonEmpty("label", props.Label)
	setIfNonEmpty("genre", props.Genre)
	setIfNonEmpty("style", props.Style)
	setIfNonEmpty("country", props.Country)
	setIfNonEmpty("year", props.Year)
	setIfNonEmpty("format", props.Format)
	setIfNonEmpty("catno", props.Catno)
	setIfNonEmpty("barcode", props.Barcode)
	setIfNonEmpty("track", props.Track)
	setIfNonEmpty("submitter", props.Submitter)
	setIfNonEmpty("contributor", props.Contributor)
	if props.Page.PerPage > 0 {
		query.Set("per_page", strconv.Itoa(props.Page.PerPage))
	}
	if props.Page.Page > 0 {
		query.Set("page", strconv.Itoa(props.Page.Page))
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, "/database/search", query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func (c *Client) GetRelease(ctx contextx.ContextX, id int) (*Release, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/releases/%d", id), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result Release
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func (c *Client) GetMaster(ctx contextx.ContextX, id int) (*Master, error) {
	resp, err := c.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/masters/%d", id), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result Master
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}
