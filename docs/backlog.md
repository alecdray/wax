# Backlog

Follow-up items not currently scheduled. Each entry: one paragraph of context + a "Next" line. Promote to a spec when picked up.

## E2E test baseline — investigate pre-existing Playwright failures

After the design-system icons + colors work (PR #9, merged as `7f1eb5e`), 11 of 39 Playwright tests were failing on the feature branch. Spot-checks during that PR suggested the failures are **pre-existing on `main`** and unrelated to that branch's scope: test-ids referenced in tests that never existed in the templates, an active/inactive nav split that predates the design-system work, and an HTMX modal flow the branch didn't touch.

Next: run `task test/e2e` against `main` (requires the app on `:4691`) to confirm the same 11 failures reproduce. If they do, file fixes in a dedicated PR — likely a mix of test corrections (stale test-ids) and one or two real bug fixes (modal flow). If any of the 11 turn out to be regressions introduced by PR #9 after all, prioritise those first.

## Simplify review flow

Drop the time-based aspects of the review flow — let the user manually move an item from unrated → provisional → final instead. Also simplify the rating flow so the rating modal opens directly rather than forcing a Q-and-A step first.

Next: spec the state transitions (including what happens to any existing time-based metadata) and decide whether the Q-and-A path is removed entirely or kept as opt-in.

## Add a clear button to discover search bar

The discover search bar has no one-click clear, so resetting the query means manually deleting characters.

Next: add a clear (×) affordance inside the search input that empties the query and re-runs the search.

## Rethink tagging system

The current tagging system isn't pulling its weight and needs a rethink. No replacement design yet — fully open.

Next: brainstorm desired UX and constraints (likely via `/brainstorming`) before drafting a spec.
