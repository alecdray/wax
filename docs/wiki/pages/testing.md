---
description: >
  How Wax is tested — testing strategy, tooling, and the development workflow around quality.
  Belongs here: unit testing approach, e2e testing, agent-driven testing, BDD conventions, and
  test tooling decisions. Does not belong here: specific test implementations (→ code), feature
  descriptions (→ features), or UI rendering details (→ frontend).
links:
  - architecture
  - frontend
  - features
---

[wiki](../wiki.md)

# Testing

How Wax is tested and the philosophy behind the testing strategy.

## Approach

Testing in Wax is intended to be **agent-driven** — AI agents (e.g. Claude Code) participate in the test authoring and execution cycle, not just humans. This means tests should be written in a way that is machine-readable and machine-runnable, with clear intent expressed through the test structure itself.

## Layers

| Layer | Scope | Tooling |
|---|---|---|
| Unit | Individual service methods and domain logic | Go `testing` package |
| E2E | Full user workflows through the browser | Playwright (Chromium) |

### Unit Tests

Unit tests target the service layer — business logic isolated from HTTP and the database. Dependencies (external services, database) are injected, making them substitutable in tests. No external test libraries are used — only the Go standard `testing` package.

Run with: `task test/unit`

### E2E Tests

End-to-end tests drive the real application through a browser. Because Wax renders server-side HTML and uses HTMX for interactivity, E2E tests are the primary way to verify that the full stack works together correctly from a user's perspective.

**Playwright** is the E2E framework, running against Chromium. Tests interact with the UI the way a real user would — navigating pages, submitting forms, and asserting on rendered output. The app must be running (`task dev`) before executing E2E tests.

Run with: `task test/e2e`

#### E2E Directory Structure

```
e2e/
├── feat/       # Gherkin-style feature files — what is being tested and why
├── helpers/    # Reusable TypeScript logic shared across specs
└── spec/       # Playwright test files — how each scenario is implemented
```

Feature files are the source of truth for what behaviour is covered. Each feature file has a corresponding spec file with test names matching scenario names exactly.

#### Selectors

Specs select elements exclusively via `data-testid` attributes (`page.getByTestId(...)`). CSS classes, roles, and text content are never used as selectors — they are presentation details that change for non-test reasons. The `data-testid` attribute is added to `.templ` files for any element a test needs to interact with or assert on.

#### Authenticated Tests

An `loginAs(context, userId)` helper in `e2e/helpers/auth.ts` injects a signed JWT cookie into the browser context, bypassing the Spotify OAuth flow for tests that cover authenticated pages. The `E2E_TEST_USER_ID` variable in `.env` holds the user ID to use. If not set, authenticated tests skip cleanly.

#### Watch and Debug Modes

- `task test/e2e -- --ui` — interactive GUI with live re-run on file changes (best for active development)
- `task test/e2e -- --headed` — visible browser, no pause
- `task test/e2e -- --debug` — Playwright Inspector, steps through each action one at a time

## BDD Style

E2E tests follow a **BDD (Behaviour-Driven Development)** convention: scenarios are expressed in terms of user-observable behaviour rather than implementation details. This makes tests readable by non-engineers and makes the intent clear to agents generating or reviewing them.

A scenario describes:
1. A starting state (e.g. a user with albums in their library)
2. An action the user takes (e.g. adds a rating)
3. The expected outcome (e.g. the rating appears in the UI)

## Dev Flow

Testing is part of the development loop, not a separate phase. When a feature is built or changed, the corresponding unit and E2E tests are written or updated as part of the same change. Agent-driven tooling can assist in generating test scaffolding and running the suite.

`task dev` writes server and build output to log files in `tmp/` (`dev-server.log`, `dev-templ.log`, `dev-tailwind.log`) alongside the console. When debugging a failing E2E test, `tmp/dev-server.log` is the first place to check for server-side errors.
