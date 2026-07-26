# Development Process

How a feature ships. The four phases are **fixed**; *how* you carry out each one is at the
agent/user's discretion. Work happens on a branch off `main`.

**Core principle — the repo holds two layers of docs.** *Canonical docs* (ADRs, architecture
docs, the data model, module READMEs) are the source of truth for **current state** and are kept
live. *Spec docs* under [`docs/spec/<timestamp>-<name>/`](./spec/) are the **point-in-time record of one
chunk of work** — its scope, goals, spec, and supporting documents — reconciled to what actually
shipped and then **frozen at merge**. Canonical docs answer *"how is it now?"*; a spec folder
answers *"what did we set out to do for this piece, and how did it land?"* Genuinely throwaway
scratch (raw exploration, dead-end notes) still lives outside the repo (`/tmp`, …) and is never
committed — only the deliberate working docs for the chunk go in `docs/spec/`.

## 1. Spec

Run `/app-spec <name>` — creates `docs/spec/<timestamp>-<name>/` with a scope doc and surfaces relevant planning
skills. See [`docs/spec/README.md`](./spec/README.md) for the folder convention.

Codify the design in the canonical docs (ADRs, module READMEs, data model) on the branch during this
phase. Validate with a grilling pass (`/grill-with-docs`) against the full set of canonical docs —
architecture rules, design rules, the testing gate and e2e conventions, the ADR log, and the
affected module READMEs — not only the files the feature edits.

## 2. Implement

Run `/app-implement` — enforces the canonical-docs-first rule, surfaces relevant build skills, and
confirms both doc layers are reconciled before handoff.

Gate: `task test` green — unit (`task test/unit`) and e2e (`task test/e2e`, with `task dev` running on
port 4691). See [testing.md](./testing.md).

## 3. Audit

Run `/app-audit` — the pre-merge gate covering both code and docs. Fix what it finds; repeat until clean.

## 4. Merge (PR → merge)

`main` is protected: every change lands via PR. Push the branch and open a pull request (use `/gh-pr` for
the canonical PR body). Once the audit is clean and review passes, **squash-merge the PR** to `main`.
The spec folder is now **frozen** — an immutable record of that chunk of work. A later change that
revisits the same area gets its own new `docs/spec/<timestamp>-<name>/`; the frozen folder is never
edited again.