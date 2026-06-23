package spotify

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Spotify rate-limits per app over a rolling 30-second window and publishes no
// remaining-quota header — the only signal is a 429 with Retry-After. Dev-mode
// apps get roughly 180 requests/minute, so the proactive pace is set well under
// that; the Retry-After pause is the authoritative backstop. See ADR 0006.
const (
	guardRatePerSec  = 3.0
	guardBurst       = 10.0
	defaultRetryWait = 5 * time.Second
)

// ErrRateLimited reports that a Spotify call was refused by the shared guard —
// either Spotify returned 429, or the guard is still pausing all calls for a
// prior Retry-After window. User-initiated callers should surface a "try again
// shortly" message (RetryAfter says how long); background syncs defer to their
// next run.
type ErrRateLimited struct {
	RetryAfter time.Duration
}

func (e *ErrRateLimited) Error() string {
	return fmt.Sprintf("spotify rate limited; retry after %s", e.RetryAfter.Round(time.Second))
}

// guard is the process-wide rate-limit guard every Spotify HTTP call routes
// through. It is an http.RoundTripper so the same instance wraps both the
// vendor SDK's per-user client and the raw client.go calls, giving the
// per-app limit a single shared view. It paces normal traffic with a token
// bucket and, on a 429, pauses all calls until the Retry-After window elapses.
type guard struct {
	base   http.RoundTripper
	bucket *tokenBucket
	now    func() time.Time

	mu          sync.Mutex
	pausedUntil time.Time
}

func newGuard(base http.RoundTripper) *guard {
	if base == nil {
		base = http.DefaultTransport
	}
	now := time.Now
	return &guard{
		base:   base,
		bucket: newTokenBucket(guardRatePerSec, guardBurst, now),
		now:    now,
	}
}

func (g *guard) RoundTrip(req *http.Request) (*http.Response, error) {
	// Breaker: while an earlier Retry-After window is open, refuse every call
	// without touching Spotify — issuing requests during the window is what
	// escalates the penalty.
	g.mu.Lock()
	pausedUntil := g.pausedUntil
	g.mu.Unlock()
	if remaining := pausedUntil.Sub(g.now()); remaining > 0 {
		return nil, &ErrRateLimited{RetryAfter: remaining}
	}

	// Proactive pacing — a bounded, cancellable wait.
	if err := g.bucket.wait(req.Context()); err != nil {
		return nil, err
	}

	resp, err := g.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		g.mu.Lock()
		g.pausedUntil = g.now().Add(retryAfter)
		g.mu.Unlock()
		resp.Body.Close()
		return nil, &ErrRateLimited{RetryAfter: retryAfter}
	}
	return resp, nil
}

// parseRetryAfter reads a Retry-After delta-seconds value (the form Spotify
// sends). Missing or unparseable headers fall back to a conservative default
// rather than zero, so the guard always pauses for something on a 429.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return defaultRetryWait
	}
	secs, err := strconv.Atoi(header)
	if err != nil || secs <= 0 {
		return defaultRetryWait
	}
	return time.Duration(secs) * time.Second
}

// tokenBucket is a minimal refilling token bucket. capacity tokens accumulate
// at perSec; wait() blocks until one is available or the context is cancelled.
type tokenBucket struct {
	perSec   float64
	capacity float64
	now      func() time.Time

	mu     sync.Mutex
	tokens float64
	last   time.Time
}

func newTokenBucket(perSec, capacity float64, now func() time.Time) *tokenBucket {
	return &tokenBucket{
		perSec:   perSec,
		capacity: capacity,
		now:      now,
		tokens:   capacity,
		last:     now(),
	}
}

func (b *tokenBucket) wait(ctx context.Context) error {
	for {
		b.mu.Lock()
		now := b.now()
		b.tokens = min(b.capacity, b.tokens+now.Sub(b.last).Seconds()*b.perSec)
		b.last = now
		if b.tokens >= 1 {
			b.tokens--
			b.mu.Unlock()
			return nil
		}
		deficit := (1 - b.tokens) / b.perSec
		b.mu.Unlock()

		timer := time.NewTimer(time.Duration(deficit * float64(time.Second)))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
