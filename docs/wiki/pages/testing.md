---
description: >
  How Wax is tested — testing strategy, tooling, and the development workflow around quality.
  Belongs here: unit testing approach, e2e testing, agent-driven testing, BDD conventions, and
  test tooling decisions. Does not belong here: specific test implementations (→ code), feature
  descriptions (→ features), UI rendering details (→ frontend), or specific commands, file paths,
  and configuration values (→ e2e/README.md).
links:
  - architecture
  - frontend
  - features
---

[Parent: wiki](../wiki.md)

# Testing

How Wax is tested and the philosophy behind the testing strategy.

## Approach

Testing in Wax is intended to be **agent-driven** — AI agents (e.g. Claude Code) participate in the test authoring and execution cycle, not just humans. This means tests should be written in a way that is machine-readable and machine-runnable, with clear intent expressed through the test structure itself.

## Layers

| Layer | Scope | Tooling |
|---|---|---|
| Unit | Individual service methods and domain logic | Go standard library |
| E2E | Full user workflows through the browser | Playwright (Chromium) |

### Unit Tests

Unit tests target the service layer — business logic isolated from HTTP and the database. Dependencies (external services, database) are injected, making them substitutable in tests.

### E2E Tests

End-to-end tests drive the real application through a browser. Because Wax renders server-side HTML and uses HTMX for interactivity, E2E tests are the primary way to verify that the full stack works together correctly from a user's perspective.

See [e2e/README.md](../../../e2e/README.md) for directory structure, running tests, selectors, authentication helpers, and debugging.

## BDD Style

E2E tests follow a **BDD (Behaviour-Driven Development)** convention: scenarios are expressed in terms of user-observable behaviour rather than implementation details. This makes tests readable by non-engineers and makes the intent clear to agents generating or reviewing them.

A scenario describes:
1. A starting state (e.g. a user with albums in their library)
2. An action the user takes (e.g. adds a rating)
3. The expected outcome (e.g. the rating appears in the UI)

## Dev Flow

Testing is part of the development loop, not a separate phase. When a feature is built or changed, the corresponding unit and E2E tests are written or updated as part of the same change. Agent-driven tooling can assist in generating test scaffolding and running the suite.
