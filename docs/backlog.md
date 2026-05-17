# Backlog

Follow-up items not currently scheduled. Each entry: one paragraph of context + a "Next" line. Promote to a spec when picked up.

## E2E test baseline — investigate pre-existing Playwright failures

After the design-system icons + colors work (PR #9, merged as `7f1eb5e`), 11 of 39 Playwright tests were failing on the feature branch. Spot-checks during that PR suggested the failures are **pre-existing on `main`** and unrelated to that branch's scope: test-ids referenced in tests that never existed in the templates, an active/inactive nav split that predates the design-system work, and an HTMX modal flow the branch didn't touch.

Next: run `task test/e2e` against `main` (requires the app on `:4691`) to confirm the same 11 failures reproduce. If they do, file fixes in a dedicated PR — likely a mix of test corrections (stale test-ids) and one or two real bug fixes (modal flow). If any of the 11 turn out to be regressions introduced by PR #9 after all, prioritise those first.

## Simplify review flow

Drop the time-based aspects of the review flow — let the user manually move an item from unrated → provisional → final instead. Also simplify the rating flow so the rating modal opens directly rather than forcing a Q-and-A step first.

Next: spec the state transitions (including what happens to any existing time-based metadata) and decide whether the Q-and-A path is removed entirely or kept as opt-in.

## Discover results panel doesn't actually scroll

The Discover results panel (`#discover-results` in `src/internal/library/adapters/views/discover_page.templ`) is wrapped in a flex chain (`h-dvh` ancestor + nested `flex-1 min-h-0 overflow-y-auto`) that computes to `overflow: visible` with `scrollHeight === clientHeight` at every viewport tested — so the whole page scrolls instead of the panel. The htmx `hx-swap="innerHTML scroll:#discover-results:top"` directive still runs and sets `scrollTop = 0` on the target (PR #28's PC1 scenario forces overflow inline-style to verify this wire contract), but the user-visible effect is masked. Surfaced during the search-clear build's integration tests.

Next: pick one — either fix the flex chain so the panel actually scrolls within the viewport, or drop the `scroll:#discover-results:top` directive (and the corresponding PC1 scenario) since it's a no-op today.

## Rethink tagging system

The current tagging system isn't pulling its weight and needs a rethink. No replacement design yet — fully open.

Next: brainstorm desired UX and constraints (likely via `/brainstorming`) before drafting a spec.
