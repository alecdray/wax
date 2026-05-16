# Backlog

Follow-up items not currently scheduled. Each entry: one paragraph of context + a "Next" line. Promote to a spec when picked up.

## E2E test baseline — investigate pre-existing Playwright failures

After the design-system icons + colors work (PR #9, merged as `7f1eb5e`), 11 of 39 Playwright tests were failing on the feature branch. Spot-checks during that PR suggested the failures are **pre-existing on `main`** and unrelated to that branch's scope: test-ids referenced in tests that never existed in the templates, an active/inactive nav split that predates the design-system work, and an HTMX modal flow the branch didn't touch.

Next: run `task test/e2e` against `main` (requires the app on `:4691`) to confirm the same 11 failures reproduce. If they do, file fixes in a dedicated PR — likely a mix of test corrections (stale test-ids) and one or two real bug fixes (modal flow). If any of the 11 turn out to be regressions introduced by PR #9 after all, prioritise those first.
