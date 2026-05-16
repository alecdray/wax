# Testing

## Approach

Testing in Wax is **agent-driven** — AI agents participate in the test authoring and execution cycle, not just humans. Tests should be machine-readable and machine-runnable, with intent expressed through their structure.

Unit tests target the service layer — business logic isolated from HTTP and the database, with dependencies injected so they can be substituted. E2E tests drive the real application through a browser; because Wax renders server-side HTML and uses HTMX for interactivity, E2E is the primary way to verify the full stack works together from a user's perspective.

See [`e2e/README.md`](../e2e/README.md) for E2E directory structure, selectors, authentication helpers, and debugging.

## BDD style

E2E scenarios are expressed in terms of user-observable behaviour rather than implementation details. This makes them readable by non-engineers and makes the intent clear to agents generating or reviewing them.

## Dev flow

Testing is part of the development loop, not a separate phase. When a feature is built or changed, the corresponding unit and E2E tests are written or updated as part of the same change.
