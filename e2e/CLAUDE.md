# E2E tests

Playwright-driven end-to-end tests. Read [`README.md`](./README.md) before writing or modifying tests — it covers structure, the 8-step recipe, helpers, watch modes, and debugging.

## Suite rules

These are the invariants the suite holds itself to. Honor them when adding or editing specs:

- **Feature ↔ spec 1:1.** Every `e2e/feat/<name>.feature` has a matching `e2e/spec/<name>.spec.ts`. Every `Scenario:` in a feature has a `test()` with the exact same name in the paired spec. The feature file is the source of truth — edit it first.
- **No orphan testids.** Every `getByTestId('foo')` must resolve to a `data-testid="foo"` declared under `src/internal/`. An orphan testid silently makes `expect(...).not.toBeVisible()` pass for the wrong reason. After adding a `getByTestId`, grep `src/internal/` to confirm the testid exists.
- **Testid naming** follows [`docs/design/testids.md`](../docs/design/testids.md).
- **Selectors are `data-testid` only.** Allowed factories: `getByTestId`, `getByRole`, `getByLabel`. No `getByText` / `getByPlaceholder` / `getByAltText` / `getByTitle` — even for content assertions, add a testid first. `page.locator(...)` is allowed only for `'dialog[open]'` (modal scoping) and `[data-testid="..."]` attribute selectors.
- **Wait on observable DOM signals, never on time.** No `page.waitForTimeout(...)`, `sleep(N)`, `delay(N)`, or `setTimeout(...)` in specs. HTMX swaps complete when the new DOM is present — assert on that.
- **Single auth path.** Authenticated specs reach the authenticated state only through `loginAs(context, userId)` from `helpers/auth.ts`. No raw `wax_token` cookie injection anywhere else.
- **Real backend.** No `page.route(...)`, `page.unroute(...)`, `page.fulfill(...)`, no MSW/nock/sinon, no `jest.mock` / `vi.mock`. If a test needs specific data, insert it into the SQLite DB directly.
- **Fail loud on missing required data.** When a test depends on an env var or fixture, guard with `expect(value, '<msg>').toBeTruthy()`, not `test.skip()` — skipping hides misconfigured environments.

## Gate

`task test/e2e` (with `task dev` in another terminal) must pass before considering a test or change done.
