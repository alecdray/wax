# Backlog

Follow-up items not currently scheduled. Each entry: one paragraph of context + a "Next" line. Promote to a spec when picked up.

## E2E test baseline — investigate pre-existing Playwright failures

After the design-system icons + colors work (PR #9, merged as `7f1eb5e`), 11 of 39 Playwright tests were failing on the feature branch. Spot-checks during that PR suggested the failures are **pre-existing on `main`** and unrelated to that branch's scope: test-ids referenced in tests that never existed in the templates, an active/inactive nav split that predates the design-system work, and an HTMX modal flow the branch didn't touch.

Next: run `task test/e2e` against `main` (requires the app on `:4691`) to confirm the same 11 failures reproduce. If they do, file fixes in a dedicated PR — likely a mix of test corrections (stale test-ids) and one or two real bug fixes (modal flow). If any of the 11 turn out to be regressions introduced by PR #9 after all, prioritise those first.

## Simplify review flow

Drop the time-based aspects of the review flow — let the user manually move an item from unrated → provisional → final instead. Also simplify the rating flow so the rating modal opens directly rather than forcing a Q-and-A step first.

Next: spec the state transitions (including what happens to any existing time-based metadata) and decide whether the Q-and-A path is removed entirely or kept as opt-in.

## Rethink tagging system

The current tagging system isn't pulling its weight and needs a rethink. No replacement design yet — fully open.

Next: brainstorm desired UX and constraints (likely via `/brainstorming`) before drafting a spec.

## Reusable pattern for migration-side ID generation

Backfill migrations need to generate row IDs in SQL, but the codebase has no consistent approach. `lower(hex(randomblob(16)))` (32-char hex) appears in two prior migrations; an inline UUID-v4 construction was used in `20260517202814`. Neither matches runtime, which uses `uuid.NewString()` (36-char dashed UUID v4). The clean fixes each have a cost — inline UUID-v4 SQL is verbose and easy to typo; a Go migration or driver-registered `uuid_v4()` function both need a custom goose binary (`src/cmd/goose/main.go`) because the vanilla CLI can't load Go code or use a custom-registered driver.

Next: when a third backfill is on the horizon, pick a convention (probably the custom goose binary + `RegisterFunc("uuid_v4", ...)` — works in both the CLI and the app, reusable forever) and document it. Until then, inline SQL is the local optimum.

## E2E suite SQLite contention / shared-test-data flakiness

Specs share a single SQLite DB with no per-test isolation, and helpers issue direct `sqlite3` CLI writes (no `busy_timeout`/WAL on the app connection; many specs share one test album/user). Run across parallel workers the suite both deadlocks (`database is locked`) and races rating/library tests on the same user's state. `playwright.config.ts` now pins `workers: 1`, which makes the suite deterministic — but only by giving up parallelism; the underlying lack of isolation remains.

Next: to restore parallelism, enable WAL + `busy_timeout` on the SQLite DSN in `core/db` and isolate per-spec test data (or shard a DB per worker), then lift the `workers: 1` pin. Until then, serial is the safe default.

## Migrate existing raw `HX-Trigger` header writes to `httpx.SetHXTrigger`

`library/adapters/http.go` still sets several `HX-Trigger` headers via raw `w.Header().Set(...)` (e.g. `libraryUpdated`, `radarUpdated`). Consider migrating them to the new `httpx.SetHXTrigger` helper for consistency. Low priority.

## Auth-aware feed dormancy (reconnect-Spotify UX)

The Spotify rate-limit hardening ([ADR 0006](adr/0006-spotify-rate-limit-guard.md)) backs every failed feed sync off uniformly — transient errors (429/5xx/network) and un-fixable auth errors (revoked refresh token, lost scope) alike. A feed whose token is dead therefore retries forever at the backoff cap (~hourly) instead of telling the user anything. That is a large improvement over the prior every-minute loop, but it is still pointless polling and offers no fix path.

Next: classify failures — keep exponential backoff for transient errors, but give feeds a dormant/needs-reconnect state for auth/permission failures that stops auto-polling and surfaces a "reconnect Spotify" prompt in the UI. Requires a new feed state, an adapter surface to show it, and a small error classifier over the existing `ErrInsufficientScope` / token-refresh failures.

## `htmx:oobErrorNoTarget` when rating from the album detail page

Saving or finalizing a rating from the album *detail* page logs two `htmx:oobErrorNoTarget` console errors. The rating handlers broadcast `album-changed`, and library's surface-refresh responds with out-of-band swaps for dashboard surfaces (the provisional carousel, the album list row) whose target ids aren't present on the detail page, so the OOB swaps land nowhere. Harmless — the rating still saves and the detail page's own readout updates — but it's console noise on every rate-from-detail action. Pre-existing; surfaced during manual verification of the rating-modal rework.

Next: decide whether the surface-refresh should scope its OOB swaps to targets present on the current page, or whether the detail page should host (hidden) the same surface ids the dashboard does. Low priority.

## Visible toast for rate-limited user actions

When the shared Spotify guard is paused ([ADR 0006](adr/0006-spotify-rate-limit-guard.md)), a user-initiated action (search, save/unsave, radar setup) now fails fast: `httpx.HandleErrorResponse` maps `spotify.ErrRateLimited` to HTTP 429 + `Retry-After`. But HTMX does not swap 4xx responses by default (`responseHandling` in `htmx.min.js`), so the user currently sees no inline message — the action simply doesn't apply. The fail-fast contract holds (no hang, no added load); the *visible* feedback does not.

Next: decide the app-wide transient-error display mechanism (a global `htmx:responseError` listener that raises a toast keyed off the 429, or the response-targets extension) and render a "Spotify is rate-limiting us, try again shortly" toast. This is a cross-cutting error-UX decision, not Spotify-specific, so it should be designed once for all transient errors.

## `auth` module is missing its archetype-required README

The `auth` domain module has `service.go` and `CLAUDE.md` but no `README.md`, which the domain-module archetype requires. Pre-existing on `main`; surfaced by the audit during the Spotify rate-limit branch (which added `user/README.md` but left `auth` untouched, since it did not change auth's behaviour).

Next: write `auth/README.md` covering the module's responsibility — JWT issuance, login orchestration, and the Spotify OAuth callback flow.

## `core/CLAUDE.md` enumerates sub-packages (exhaustive-list rot)

`src/internal/core/CLAUDE.md` lists every `core/*` sub-package (app, contextx, cryptox, db, httpx, sqlx, task, templates, timex, utils) under "Sub-packages". This is the exact "no exhaustive lists" violation that `docs/architecture/CLAUDE.md` warns against — the list goes stale silently the moment a sub-package is added or renamed. Pre-existing on `main`; surfaced by the docs-audit during the Spotify rate-limit branch.

Next: replace the enumerated list with a conceptual description of what `core/` holds (focused, framework-level utilities used by 2+ modules — time, DB, HTTP, context, encryption, task scheduling, SQL helpers, shared UI), and let the live tree be the source of truth. Low priority.

## E2E suite-rule duplication not registered as synchronized

The e2e suite rules (feature↔spec 1:1, no orphan testids, testid naming, selectors, wait signals, single auth path, real backend, fail loud) appear in both `e2e/README.md` (detailed guide) and `e2e/CLAUDE.md` (auto-loaded quick reference), but this intentional duplication is not listed in `.claude/CLAUDE.md`'s "Synchronized content" section, so the two can drift. Pre-existing on `main`; surfaced by the docs-audit during the Spotify rate-limit branch.

Next: either register the duplication in "Synchronized content" (README = detailed, CLAUDE.md = quick reference; update both together) or consolidate to full rules in `e2e/README.md` with `e2e/CLAUDE.md` referring to it. Low priority.
