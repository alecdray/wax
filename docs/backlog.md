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

Running `task test/e2e` at the default (multi-worker) parallelism intermittently fails with `database is locked` because spec files run in parallel against one SQLite DB while helpers issue direct `sqlite3` CLI writes (no `busy_timeout`/WAL on the app connection; many specs share one test album/user). This is pre-existing and orthogonal to feature work.

Next: consider enabling WAL + `busy_timeout` on the SQLite DSN in `core/db`, setting `workers: 1` (or sharding) in `playwright.config.ts`, or isolating per-spec test data. The suite passes deterministically at `--workers=1`.

## Migrate existing raw `HX-Trigger` header writes to `httpx.SetHXTrigger`

`library/adapters/http.go` still sets several `HX-Trigger` headers via raw `w.Header().Set(...)` (e.g. `libraryUpdated`, `radarUpdated`). Consider migrating them to the new `httpx.SetHXTrigger` helper for consistency. Low priority.

## `htmx:oobErrorNoTarget` when rating from the album detail page

Saving or finalizing a rating from the album *detail* page logs two `htmx:oobErrorNoTarget` console errors. The rating handlers broadcast `album-changed`, and library's surface-refresh responds with out-of-band swaps for dashboard surfaces (the provisional carousel, the album list row) whose target ids aren't present on the detail page, so the OOB swaps land nowhere. Harmless — the rating still saves and the detail page's own readout updates — but it's console noise on every rate-from-detail action. Pre-existing; surfaced during manual verification of the rating-modal rework.

Next: decide whether the surface-refresh should scope its OOB swaps to targets present on the current page, or whether the detail page should host (hidden) the same surface ids the dashboard does. Low priority.
