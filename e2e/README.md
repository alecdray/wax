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
6. **If the testid you need does not yet exist**, add `data-testid="..."` to the relevant `.templ` file following the [naming convention](#testid-naming-convention) below, then `task build/templ`.
7. **Wait on observable DOM signals, never on time.** HTMX swaps complete when the new DOM is present — assert on that. `page.waitForTimeout(...)` is banned; it produces flaky tests and hides race conditions.
8. **Run the alignment check and the suite.**
   ```bash
   npm run e2e:check        # static check: every spec testid is declared in some templ
   task test/e2e            # full Playwright run (needs `task dev` in another terminal)
   ```
   Both must pass before the test is done.

### Common pitfalls

- **SQLite `CURRENT_TIMESTAMP` is second-resolution.** Two rows inserted in the same second tie on `ORDER BY created_at DESC`. Tests that depend on insertion order should accept either ordering, or insert a deliberate gap.
- **A passing spec with an undeclared testid is silently wrong.** If `getByTestId('foo')` matches nothing, `expect(...).not.toBeVisible()` passes vacuously. Run `npm run e2e:check` to catch these.
- **Stale `_templ.go` after a branch switch.** `_templ.go` files are gitignored, so checking out a branch that added or renamed testids in `.templ` does **not** bring the matching generated Go with it. The static check (`npm run e2e:check`) reads `.templ` source and will pass, but the running server is built from the old `_templ.go` and won't render the new testids — specs time out waiting for elements that source says exist. Run `task build/templ` after any branch switch and before `task dev`. (The cold-start checklist above bakes this in.)

## Testid alignment check

`npm run e2e:check` is a fast, dependency-free static check that every `data-testid` referenced by a spec (via `getByTestId('...')`) is declared by at least one `.templ` under `src/internal/`. It reads files only — no dev server, DB, or browser. Source: [`check-testid-alignment.mjs`](./check-testid-alignment.mjs).

**Clean run** — every spec reference resolves, exit 0:

```
OK — 123 spec testid reference(s) across 5 spec file(s) all declared (scanned 87 templ file(s), 214 declared testid(s)).
```

**Dirty run** — one line per orphan in `<spec>:<line>: <testid>` form, a summary on stderr, exit 1. Example output from the current repo:

```
e2e/spec/library.spec.ts:46: album-row-menu
e2e/spec/library.spec.ts:47: album-row-tags-button
e2e/spec/reviews.spec.ts:151: rating-delete
e2e/spec/reviews.spec.ts:242: album-row-notes
e2e/spec/reviews.spec.ts:243: album-row-notes-button

5 orphan reference(s) across 5 distinct missing testid(s); spec testid is not declared in any templ under src/internal.
```

To fix an orphan, either add the missing `data-testid="..."` to the appropriate templ (and `task build/templ`) or update the spec to reference an existing testid. The check does not detect the reverse case (testids declared in templs but unused by specs) — those are not failures.

## Testid naming convention

`<surface>[-<element>][-<modifier>]`, kebab-case throughout.

- **surface** — derived from the declaring templ's filename, with `_page` / `_frag` / `_modal` suffixes dropped and underscores converted to hyphens. For example, `login_page.templ` → surface `login-page`; `album_score_readout_frag.templ` → surface `album-score-readout`.
- **element** — optional. Names the role of the specific node within the surface (`button`, `link`, `cover`, `title`, `state-icon`).
- **modifier** — optional. Names the variant or state (`rated`, `unrated`, `open`).

Concrete examples from the current codebase:

| Templ file | Surface | Declared testid |
|---|---|---|
| `auth/adapters/views/login_page.templ` | `login-page` | `login-page`, `login-page-button`, `login-page-link` |
| `library/adapters/views/album_score_readout_frag.templ` | `album-score-readout` | `album-score-readout-rated`, `album-score-readout-unrated`, `album-score-readout-state-icon` |
| `library/adapters/views/album_detail_page.templ` | `album-detail-page` | `album-detail-page-title`, `album-detail-page-cover`, `album-detail-page-rating` |

**Cross-surface composition is allowed.** A fragment that is consumed by exactly one page may declare testids using the consuming page's surface name. For example, `formats_releases_frag.templ` declares `album-detail-page-releases` because it is composed into `album_detail_page.templ`. The rule is "surface from filename" by default; explicit cross-composition is valid when the fragment is owned by a consuming page.

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

The grep-everything approach is the source of truth. Because of cross-surface composition, you cannot infer where a testid is declared from its name alone — always grep.

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

- **BDD style**: scenarios describe behaviour, not implementation. Write from the user's point of view.
- **One spec per feature file**: `login.feature` → `login.spec.ts`. Scenarios map 1:1 to tests by exact name.
- **Feature file is the source of truth**: if a scenario changes, update the feature file first, then the spec.
- **No shared state between tests**: each test must be self-contained. Use `page.goto` to set up starting state.
- **No mock backend**: tests run against the real Go server and SQLite database. `loginAs(context, userId)` is the only sanctioned bypass.
- **Selectors are `data-testid` only** (with narrow exceptions for `getByRole` / `getByLabel` on semantic form controls and `dialog[open]` for modal scoping). See [Testid naming convention](#testid-naming-convention).
- **Wait on observable DOM signals, never on time**: `page.waitForTimeout(...)` is banned. HTMX swaps complete when the new DOM is present — assert on that.

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
