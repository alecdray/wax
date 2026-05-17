# E2E Tests

End-to-end tests drive the real application through a browser using [Playwright](https://playwright.dev). They verify full-stack behaviour from a user's perspective.

## Structure

```
e2e/
├── feat/       # Gherkin-style feature files — what is being tested and why
├── helpers/    # Reusable logic shared across specs (auth, page setup, etc.)
└── spec/       # Playwright test files — how each scenario is implemented
```

Every feature file in `feat/` has a corresponding spec file in `spec/`. Scenarios in the feature file map 1:1 to test cases in the spec file, matched by name.

## Running

The app must be running before executing E2E tests. Use `task dev` (not `task dev/server` alone) — it also starts the `templ generate --watch` loop, so live edits to `.templ` files are reflected in the running server.

```bash
# Terminal 1 — start the app
task dev

# Terminal 2 — run E2E tests
task test/e2e

# Run a specific spec file
task test/e2e -- e2e/spec/login.spec.ts
```

Playwright targets `http://127.0.0.1:${PORT}`, where `PORT` is read from `.env` (default `4691`). Running each worktree on its own port via `.env` keeps multiple dev servers from colliding — the suite follows automatically.

### Cold start in a fresh worktree

When `task dev` is starting in a tree that hasn't been actively edited (a new worktree, or after switching branches), run these once before the first `task test/e2e`:

```bash
cp /Users/shmoopy/workshop/projects/wax/.env .env   # worktrees don't have .env
cp /Users/shmoopy/workshop/projects/wax/tmp/db.sql ./tmp/db.sql   # seed the fixture user/album rows
npm install                                          # if node_modules is missing
task build/templ                                     # regenerate _templ.go from .templ — see pitfall below
```

After that, `task dev` + `task test/e2e` is the steady-state loop.

### Watch modes

**`--ui`** — interactive GUI with a test sidebar, browser preview, and live re-run on file changes. Best for active development:
```bash
task test/e2e -- --ui
```

**`--headed`** — runs tests in a visible browser with no pause. Useful for a quick visual sanity check:
```bash
task test/e2e -- --headed
```

**`--debug`** — opens Playwright Inspector and steps through each action one at a time. Best for diagnosing a failing test:
```bash
task test/e2e -- --debug
task test/e2e -- --debug e2e/spec/login.spec.ts
```

## Writing a new test

Follow these steps in order. The last step is non-negotiable: **run the alignment check and the suite** before considering the test done.

1. **Decide what user-facing behaviour you are describing.** One feature file per surface (page or top-level fragment). If a behaviour spans two surfaces, that's two features.
2. **Write the feature file first** (`e2e/feat/<name>.feature`). Plain English, BDD style, from the user's point of view — no CSS classes, IDs, or implementation details:
   ```gherkin
   Feature: <Name>

     <One-sentence description of the area being tested.>

     Scenario: <What the user does and expects>
       Given <starting state>
       When <action>
       Then <expected outcome>
   ```
3. **Create the matching spec file** (`e2e/spec/<name>.spec.ts`, same base name as the feature). Add a header comment linking back to the feature:
   ```ts
   import { test, expect } from '@playwright/test';

   // Scenarios from e2e/feat/<name>.feature

   test('<Scenario name verbatim>', async ({ page }) => {
     // ...
   });
   ```
   Test names must match the scenario names in the feature file **exactly** — the feature↔spec mapping is name-based 1:1.
4. **Log in if the scenario is authenticated.** Use `loginAs(context, userId)` from `helpers/auth.ts`. It is the only sanctioned bypass of Spotify OAuth — there is no mock backend.
5. **Locate elements with `data-testid` only.** Use `page.getByTestId('...')`. The two narrow exceptions are `getByRole(...)` / `getByLabel(...)` for semantic assertions on standard form controls, and `dialog[open]` for scoping inside an open modal. Never select on CSS classes, raw text, or structural selectors — they change for non-test reasons.
6. **If the testid you need does not yet exist**, add `data-testid="..."` to the relevant `.templ` file following the [naming convention in `docs/design/testids.md`](../docs/design/testids.md), then `task build/templ`.
7. **Wait on observable DOM signals, never on time.** HTMX swaps complete when the new DOM is present — assert on that. `page.waitForTimeout(...)` is banned; it produces flaky tests and hides race conditions.
8. **Run the suite.** `task test/e2e` (with `task dev` in another terminal) must pass before the test is done. The suite-wide rules in [Conventions](#conventions) below aren't automated — verify them yourself, especially that every `getByTestId` you add resolves to a real declaration in `src/internal/`.

### Common pitfalls

- **SQLite `CURRENT_TIMESTAMP` is second-resolution.** Two rows inserted in the same second tie on `ORDER BY created_at DESC`. Tests that depend on insertion order should accept either ordering, or insert a deliberate gap.
- **A passing spec with an undeclared testid is silently wrong.** If `getByTestId('foo')` matches nothing, `expect(...).not.toBeVisible()` passes vacuously. Grep `src/internal/` for the testid after adding a new `getByTestId` call to confirm it actually exists.
- **Stale `_templ.go` after a branch switch.** `_templ.go` files are gitignored, so checking out a branch that added or renamed testids in `.templ` does **not** bring the matching generated Go with it. The `.templ` source will show the new testid, but the running server is built from the old `_templ.go` and won't render it — specs time out waiting for elements that source says exist. Run `task build/templ` after any branch switch and before `task dev`. (The cold-start checklist above bakes this in.)

## Discovering existing testids

Before adding a new `data-testid`, check what is already declared — there is often something you can reuse.

**List every declared testid in the codebase, sorted and deduplicated:**

```bash
grep -rhoE 'data-testid="[^"]+"' src/internal/ | sort -u
```

**List declared testids for one specific surface, with file and line:**

```bash
grep -rn 'data-testid' src/internal/auth/adapters/views/login_page.templ
```

**Find the templ that declares a given testid:**

```bash
grep -rn 'data-testid="album-score-readout-rated"' src/internal/
```

The grep-everything approach is the source of truth. A testid's prefix doesn't always match its file's own component name (a fragment scoped to one parent uses the parent's prefix — see [`docs/design/testids.md`](../docs/design/testids.md)), so always grep.

## Helpers

Shared logic lives in `helpers/`. Import from there rather than duplicating setup across specs.

### `auth.ts` — `loginAs(context, userId)`

Injects a signed `wax_token` JWT cookie into the browser context, bypassing the Spotify OAuth flow. Use this in any test that covers an authenticated page.

```ts
import { loginAs } from '../helpers/auth';

test('user sees their library', async ({ context, page }) => {
  await loginAs(context, 'a-real-user-id-from-the-db');
  await page.goto('/app/library/dashboard');
  // ...
});
```

`loginAs` reads `JWT_SECRET` from the environment (loaded automatically from `.env` by the test runner). The userId must exist in the database.

### Adding a new helper

Put new helpers in `helpers/<name>.ts` and export named functions. Keep helpers focused — one concern per file.

## Conventions

BDD style (scenarios describe behaviour, not implementation) is covered in [`docs/testing.md`](../docs/testing.md). The rules below are the suite-wide invariants — honor them when adding or editing specs. None of them are automated; the only gate is `task test/e2e` passing.

### Feature ↔ spec 1:1

Every `e2e/feat/<name>.feature` has a matching `e2e/spec/<name>.spec.ts`, and every `Scenario:` in a feature has a `test()` with the **exact same name** in the paired spec — no orphans in either direction.

The feature file is the source of truth: when behaviour changes, edit the feature first, then the spec. Renaming a scenario means editing two files in lockstep.

### No orphan testids

Every `getByTestId('foo')` in a spec must resolve to a `data-testid="foo"` declared by some templ under `src/internal/`. An orphan testid silently makes `expect(...).not.toBeVisible()` pass for the wrong reason — the element doesn't exist, so the negative assertion succeeds vacuously.

After adding a `getByTestId` call, grep `src/internal/` to confirm the testid exists. If it doesn't, add `data-testid="..."` to the relevant templ and run `task build/templ`.

### Testid naming

Naming follows [`docs/design/testids.md`](../docs/design/testids.md). The grep-everything approach (see [Discovering existing testids](#discovering-existing-testids)) is the source of truth — a testid's prefix doesn't always match its file's own component name (a fragment scoped to one parent uses the parent's prefix), so always grep before inventing a new one.

### Selectors are `data-testid` only

Allowed locator factories: `getByTestId`, `getByRole`, `getByLabel`. The latter two are reserved for semantic assertions on standard form controls.

Do not use `getByText`, `getByPlaceholder`, `getByAltText`, or `getByTitle` — even for content assertions. If you want to assert that a piece of text is visible, add a testid to its container and assert text on that. Copy, placeholders, alt attributes, and title attributes change for non-test reasons; testids are the only stable contract.

`page.locator(...)` is allowed only for `'dialog[open]'` (modal scoping) and `[data-testid="..."]` attribute selectors (including comma-separated alternation, semantically equivalent to chained `getByTestId`). No CSS classes, no `nth-of-type`, no XPath — anywhere.

### Wait on observable DOM signals, never on time

No `page.waitForTimeout(...)`, no `sleep(N)` / `delay(N)`, no `setTimeout(...)` in specs. HTMX swaps complete when the new DOM is present — assert on that with `expect(page.getByTestId(...)).toBeVisible()` or equivalent. For a specific network response, use `page.waitForResponse(...)`.

### Single auth path

Authenticated specs reach the authenticated state only through `loginAs(context, userId)` from `helpers/auth.ts`. No raw `wax_token` cookie injection anywhere else. The bypass needs a single owner — alternate paths would mean multiple things to update when JWT shape or cookie name changes, and would create unsanctioned ways to skip OAuth that drift from production.

### Real backend

Tests run against the real Go server and SQLite database. No mocking, no request interception, no test-double libraries. Specifically forbidden: `page.route(...)`, `page.unroute(...)`, `page.fulfill(...)`, `MockServiceWorker`, imports from `msw` / `nock` / `sinon`, `jest.mock|fn|spyOn`, `vi.mock|fn|spyOn`.

If a test needs a specific data shape, insert the fixture into the database directly. If it needs to exercise an external API call, that belongs in a unit test against the adapter, not in e2e.

### Fail loud on missing required data

When a test depends on an env var or fixture, guard with `expect(value, '<msg>').toBeTruthy()`, not `test.skip()`. Skipping silently hides a misconfigured environment; a failure surfaces the problem immediately.

### Self-contained tests

Each test must be self-contained. Use `page.goto` (and `loginAs` if authenticated) to set up starting state — never depend on what a previous test left behind.

## Logs

`task dev` writes output from each watcher to a file in `tmp/` alongside the console:

| File | Source |
|---|---|
| `tmp/dev-server.log` | Go server — HTTP requests, app errors, slog output |
| `tmp/dev-templ.log` | Templ compiler — template build errors |
| `tmp/dev-tailwind.log` | Tailwind — CSS build errors |

Logs are overwritten on each `task dev` restart. When debugging a failing E2E test, check `tmp/dev-server.log` first for server-side errors that the browser wouldn't surface.

## Maintenance

- When a feature changes, update the feature file and spec together in the same commit.
- If a scenario is no longer valid, remove it from both files.
- Keep feature files free of implementation detail — they should read like plain English.
