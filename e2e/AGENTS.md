# E2E tests

Playwright-driven end-to-end tests. Read [`README.md`](./README.md) before writing or modifying tests — it covers structure, the 8-step recipe, helpers, watch modes, and debugging.

## Suite rules

Defined in [`README.md` §Conventions](./README.md#conventions) — that's the source of truth. Read it before writing or editing specs.

## Gate

`task test/e2e` (with `task dev` in another terminal) must pass before considering a test or change done.
