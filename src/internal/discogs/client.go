package discogs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/alecdray/wax/src/internal/core/contextx"
)

const origin = "https://api.discogs.com"

type rateLimit struct {
	mu        sync.Mutex
	remaining int
}

func (r *rateLimit) update(resp *http.Response) {
	remaining, err := strconv.Atoi(resp.Header.Get("X-Discogs-Ratelimit-Remaining"))
	if err != nil {
		return
	}
	r.mu.Lock()
	r.remaining = remaining
	r.mu.Unlock()
}

func (r *rateLimit) release() {
	r.mu.Lock()
	r.remaining++
	r.mu.Unlock()
}

// acquire claims one request slot. If none are available it waits one second
// and retries, so concurrent callers queue up rather than all firing at once.
func (r *rateLimit) acquire(ctx contextx.ContextX) error {
	for {
		r.mu.Lock()
		if r.remaining > 0 {
			r.remaining--
			r.mu.Unlock()
			return nil
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
	if err := c.rateLimit.acquire(ctx); err != nil {
		return nil, err
	}

	reqURL, err := url.Parse(origin + path)
	if err != nil {
		return nil, err
	}
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Discogs key=%s, secret=%s", c.consumerKey, c.consumerSecret))
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		c.rateLimit.release()
		return nil, err
	}

	c.rateLimit.update(resp)
	return resp, nil
}

type SearchProps struct {
	Query   string
	Type    string
	PerPage int
	Page    int
}

func (c *Client) SearchDatabase(ctx contextx.ContextX, props SearchProps) (*SearchResult, error) {
	query := url.Values{}
	if props.Query != "" {
		query.Set("q", props.Query)
	}
	if props.Type != "" {
		query.Set("type", props.Type)
	}
	if props.PerPage > 0 {
		query.Set("per_page", strconv.Itoa(props.PerPage))
	}
	if props.Page > 0 {
		query.Set("page", strconv.Itoa(props.Page))
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
