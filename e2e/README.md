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

The app must be running before executing E2E tests.

```bash
# Terminal 1 — start the app
task dev

# Terminal 2 — run E2E tests
task test/e2e

# Run a specific spec file
task test/e2e -- e2e/spec/login.spec.ts
```

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

**1. Write the feature file first** (`e2e/feat/<name>.feature`):

```gherkin
Feature: <Name>

  <One-sentence description of the area being tested.>

  Scenario: <What the user does and expects>
    Given <starting state>
    When <action>
    Then <expected outcome>
```

Keep scenarios focused on user-observable behaviour. Avoid implementation details like CSS classes or internal IDs.

**2. Create the matching spec file** (`e2e/spec/<name>.spec.ts`):

```ts
import { test, expect } from '@playwright/test';

// Scenarios from e2e/feat/<name>.feature

test('<Scenario name verbatim>', async ({ page }) => {
  // ...
});
```

Test names must match the scenario names in the feature file exactly. The comment at the top links the spec back to its feature file.

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
- **One spec per feature file**: `login.feature` → `login.spec.ts`.
- **Feature file is the source of truth**: if a scenario changes, update the feature file first, then the spec.
- **No shared state between tests**: each test must be self-contained. Use `page.goto` to set up starting state.
- **Use `data-testid` for selectors**: specs select elements via `page.getByTestId('...')`. Add `data-testid` attributes to `.templ` files for any element a test needs to interact with or assert on. Never use CSS classes, roles, or text content as selectors — these are implementation details that change for non-test reasons.

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
