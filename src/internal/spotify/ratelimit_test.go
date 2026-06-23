package spotify

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fakeRoundTripper struct {
	calls   int
	respond func(*http.Request) (*http.Response, error)
}

func (f *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls++
	return f.respond(req)
}

func makeResp(status int, retryAfter string) *http.Response {
	h := http.Header{}
	if retryAfter != "" {
		h.Set("Retry-After", retryAfter)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

// newTestGuard builds a guard with an injectable clock and a token bucket large
// enough never to block, so tests exercise only the breaker.
func newTestGuard(base http.RoundTripper, clock *time.Time) *guard {
	now := func() time.Time { return *clock }
	return &guard{
		base:   base,
		bucket: newTokenBucket(1_000_000, 1_000_000, now),
		now:    now,
	}
}

func mustReq(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://api.spotify.com/v1/me", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	return req
}

func TestGuardPassesThroughNon429(t *testing.T) {
	clock := time.Unix(0, 0)
	base := &fakeRoundTripper{respond: func(*http.Request) (*http.Response, error) {
		return makeResp(http.StatusOK, ""), nil
	}}
	g := newTestGuard(base, &clock)

	resp, err := g.RoundTrip(mustReq(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if base.calls != 1 {
		t.Fatalf("base calls = %d, want 1", base.calls)
	}
}

func TestGuard429OpensBreakerAndFailsFast(t *testing.T) {
	clock := time.Unix(0, 0)
	base := &fakeRoundTripper{respond: func(*http.Request) (*http.Response, error) {
		return makeResp(http.StatusTooManyRequests, "30"), nil
	}}
	g := newTestGuard(base, &clock)

	// First call hits Spotify, gets 429, opens the breaker, returns ErrRateLimited.
	_, err := g.RoundTrip(mustReq(t))
	var rl *ErrRateLimited
	if !errors.As(err, &rl) {
		t.Fatalf("first call error = %v, want *ErrRateLimited", err)
	}
	if rl.RetryAfter != 30*time.Second {
		t.Fatalf("RetryAfter = %s, want 30s", rl.RetryAfter)
	}
	if base.calls != 1 {
		t.Fatalf("base calls after first = %d, want 1", base.calls)
	}

	// Second call while the window is open must NOT touch Spotify — fail fast.
	_, err = g.RoundTrip(mustReq(t))
	if !errors.As(err, &rl) {
		t.Fatalf("second call error = %v, want *ErrRateLimited", err)
	}
	if base.calls != 1 {
		t.Fatalf("base calls after fail-fast = %d, want 1 (no new request)", base.calls)
	}

	// Once the window elapses, calls resume.
	clock = clock.Add(30 * time.Second)
	base.respond = func(*http.Request) (*http.Response, error) {
		return makeResp(http.StatusOK, ""), nil
	}
	resp, err := g.RoundTrip(mustReq(t))
	if err != nil {
		t.Fatalf("post-window error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("post-window status = %d, want 200", resp.StatusCode)
	}
	if base.calls != 2 {
		t.Fatalf("base calls after window = %d, want 2", base.calls)
	}
}

func TestParseRetryAfter(t *testing.T) {
	cases := []struct {
		header string
		want   time.Duration
	}{
		{"30", 30 * time.Second},
		{"1", 1 * time.Second},
		{"", defaultRetryWait},
		{"abc", defaultRetryWait},
		{"0", defaultRetryWait},
		{"-5", defaultRetryWait},
	}
	for _, c := range cases {
		if got := parseRetryAfter(c.header); got != c.want {
			t.Errorf("parseRetryAfter(%q) = %s, want %s", c.header, got, c.want)
		}
	}
}
