# Bugs

Known bugs, regressions, and operational tech debt. Include repro steps or a pointer to them when known.

## Active

### E2E test baseline — pre-existing Playwright failures

After the design-system icons + colors work (PR #9, merged as `7f1eb5e`), 11 of 39 Playwright tests were failing on the feature branch. Spot-checks during that PR suggested the failures are **pre-existing on `main`** and unrelated to that branch's scope.

Next: run `task test/e2e` against `main` to confirm the same 11 failures reproduce. If they do, file fixes in a dedicated PR. If any of the 11 turn out to be regressions, prioritise those first.

### E2E suite SQLite contention / shared-test-data flakiness

Specs share a single SQLite DB with no per-test isolation, and helpers issue direct `sqlite3` CLI writes. `playwright.config.ts` now pins `workers: 1`, which makes the suite deterministic — but only by giving up parallelism.

Next: enable WAL + `busy_timeout` on the SQLite DSN and isolate per-spec test data (or shard a DB per worker), then lift the `workers: 1` pin.

### `htmx:oobErrorNoTarget` when rating from the album detail page

Saving or finalizing a rating from the album *detail* page logs two `htmx:oobErrorNoTarget` console errors. The rating handlers broadcast `album-changed`, and library's surface-refresh responds with OOB swaps for dashboard surfaces whose target ids aren't present on the detail page. Harmless — the rating still saves and the detail page updates — but it's console noise.

Next: decide whether the surface-refresh should scope its OOB swaps to targets present on the current page, or whether the detail page should host (hidden) the same surface ids the dashboard does.

## Tech debt

### Reusable pattern for migration-side ID generation

Backfill migrations need to generate row IDs in SQL, but the codebase has no consistent approach. `lower(hex(randomblob(16)))` (32-char hex) appears in two prior migrations; an inline UUID-v4 construction was used in `20260517202814`. Neither matches runtime, which uses `uuid.NewString()` (36-char dashed UUID v4).

Next: when a third backfill is on the horizon, pick a convention (probably a custom goose binary + `RegisterFunc("uuid_v4", ...)`) and document it.

### Migrate raw `HX-Trigger` header writes to `httpx.SetHXTrigger`

`library/adapters/http.go` still sets several `HX-Trigger` headers via raw `w.Header().Set(...)`. Consider migrating them to the new `httpx.SetHXTrigger` helper for consistency.

### `auth` module missing its archetype-required README

The `auth` domain module has `service.go` and `AGENTS.md` but no `README.md`, which the domain-module archetype requires. Pre-existing on `main`.

Next: write `auth/README.md` covering the module's responsibility.

### `core/AGENTS.md` enumerates sub-packages (exhaustive-list rot)

`src/internal/core/AGENTS.md` lists every `core/*` sub-package under "Sub-packages". This is the exact "no exhaustive lists" violation that `docs/architecture/AGENTS.md` warns against.

Next: replace the enumerated list with a conceptual description, and let the live tree be the source of truth.

### E2E suite-rule duplication not registered as synchronized

The e2e suite rules appear in both `e2e/README.md` (detailed guide) and `e2e/AGENTS.md` (auto-loaded quick reference), but this intentional duplication is not listed in `AGENTS.md`'s "Synchronized content" section, so the two can drift.

Next: either register the duplication in "Synchronized content" or consolidate to full rules in `e2e/README.md` with `e2e/AGENTS.md` referring to it.