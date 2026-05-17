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
