# Spotify calls flow through a shared rate-limit guard that honors Retry-After

Every Spotify Web API call passes through a single, process-wide guard that paces requests and, on a `429`, stops *all* Spotify calls until the `Retry-After` window elapses. Background syncs that fail back off instead of retrying each tick, and the OAuth access token is cached until it actually expires.

Spotify's rate limit is per-app over a rolling 30-second window; exceeding it returns `429` with a `Retry-After`, and continuing to call *during* that window escalates the penalty — the wait can grow from seconds to hours. Wax previously had no rate-limit handling at all: no pacing, no `Retry-After`, a token refresh issued before every operation, and cron tasks that re-hit the same failing feed every minute. A single overrun therefore became a self-sustaining outage that never recovered on its own. Because the limit is counted per-app, the guard must be one shared instance spanning every call path — a per-call limiter cannot bound the app-wide total.

The decisions that follow from this:

- **User-initiated operations fail fast while the window is open** — search, saving an album, radar setup, manual sync — surfacing a "rate-limited, try again shortly" signal rather than hanging, since the wait may exceed any reasonable request lifetime. Background syncs simply defer to their next eligible run.
- **A failed sync backs off** rather than staying eligible for the next tick. The prior design left a failed feed immediately re-syncable, which is what sustained the penalty. The recurring poll is also unified: both feed kinds run on one incremental cadence instead of the prior split (saved albums hourly, the radar inbox every tick), and the expensive full-library backfill is reserved for first sync and reconnects rather than the recurring poll.
- **The guard is in-process.** Surviving a restart is not worth persisting: a restart mid-window costs at most one probe call, which immediately re-trips the guard — not a storm.
- **The access token is cached until it expires** rather than re-exchanged before every call — the prior refresh-per-call roughly doubled request volume against the budget. The cached token is persisted with the same encryption-at-rest already used for the refresh token, so a restart resumes from it instead of re-exchanging for every user at once.

Separately, only the production deployment runs the periodic Spotify polling: every non-prod instance shares the same Spotify app credentials and therefore the same budget, so local and dev rely on manual syncs instead.

Rejected: per-request retry without a shared breaker — concurrent callers still stampede the API during an open window, and nothing bounds the per-app total. Also rejected: blocking a user's request until the window clears — a multi-hour `Retry-After` makes that indistinguishable from a hang.
