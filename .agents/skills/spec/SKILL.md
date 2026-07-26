---
name: spec
description: Starts phase 1 (Spec) — creates the spec folder with a scope doc, then surfaces eligible planning skills for the user to choose from. Keywords: spec, phase 1, new feature, start work, plan, brainstorm, scoping.
argument-hint: "<name of the work — kebab-case slug, prefixed with current Unix timestamp>"
---

Phase 1 of the process in [`docs/process.md`](../../../docs/process.md). Full convention: [`docs/spec/README.md`](../../../docs/spec/README.md).

## Steps

1. **Create the spec folder.** Get the current Unix timestamp, then make `docs/spec/<timestamp>-<name>/` with a single `scope.md` — a heading and one short paragraph describing the work at a high level. The slug comes from the user's argument; if not supplied, ask for one before proceeding.

2. **Surface eligible skills.** Look through your available skills and present the ones relevant to planning and design work as a short list with their one-line descriptions. Ask the user which, if any, they want to use next.