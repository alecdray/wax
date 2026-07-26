# Agent Guidelines

> **Convention note:** This file is the tool-agnostic entry point. Claude Code loads it via the sibling `CLAUDE.md` symlink; other agents load `AGENTS.md` directly. The same pattern holds throughout the repo — every `AGENTS.md` has a generated `CLAUDE.md` symlink beside it, and `.claude/` is a symlink to the canonical `.agents/`. Git hooks in `.githooks/` regenerate these symlinks automatically on checkout/merge/rewrite.

## Starting new work

All new work ships through the four-phase process in [`docs/process.md`](./docs/process.md) — read it before planning a feature. **Spec is phase 1:** create the work's spec folder [`docs/spec/<timestamp>-<name>/`](./docs/spec/) (scope, goals, spec) and plan the work. Codifying the design into the canonical docs (ADRs, module READMEs) is the opening move of phase 2 (Implement); `/build` and similar tools come after, never as the entry point.

## Code Generation

### Templates
- After modifying `.templ` files, run: `task build/templ`
- Generated files end in `_templ.go`

### Database
- After modifying `.sql` files in `db/queries/`, run: `task build/sqlc`
- After creating migrations in `db/migrations/`, run: `task db/up`
- Use `task db/create -- migration_name` to create new migrations

## Architecture Patterns

`src/internal/` is organized by archetype. Every directory under `src/internal/` has an `AGENTS.md` declaring its archetype (or, for `server` and `core`, documenting singleton rules). When working in a module, the relevant rules will auto-load with that directory's `AGENTS.md`.

Full rules: [`docs/architecture/`](./docs/architecture/).

For agents adding new code:
- New code under `src/internal/<module>/` must follow the rules of that module's archetype. Read the module's `AGENTS.md` first.
- New modules: pick an archetype before writing code. If unsure, see [`docs/architecture/README.md`](./docs/architecture/README.md).

## Design

Every `.templ` file is one of three design archetypes (page templ, fragment templ, primitive), determined by location. Cross-cutting design principles (HTMX-first, fragments over pages, inline errors, theme tokens) and the visual vocabulary (Tailwind + DaisyUI `wax` theme) live alongside the archetype docs.

Full rules: [`docs/design/`](./docs/design/).

## Development

- Use `task` command for all build/run operations (see `taskfile.yml`)
- Always prefer `task <name>` over invoking tools directly (e.g. `task build/templ` not `templ generate`)
- All `go build` commands must output to `./bin/` using `-o ./bin/<name>` — never build to the project root
- Environment variables documented in `.env.template`
- Run `task` without arguments to list available commands
- Worktrees don't have a `.env` — copy from the main project: `cp /Users/shmoopy/workshop/projects/wax/.env .env`
- Also run `npm install` in the worktree before `task dev` if `node_modules` is missing
- Copy the DB from the main project to avoid 500s from missing users: `cp /Users/shmoopy/workshop/projects/wax/tmp/db.sql ./tmp/db.sql`
- `main` is protected — direct pushes are rejected. All changes (including docs-only) must land via PR. Use `/gh-pr` after committing.

## Testing

- **Strategy, unit-test conventions, the dev flow, and the gate** — [`docs/testing.md`](./docs/testing.md).
- **E2E suite rules + 8-step authoring recipe** — [`e2e/README.md`](./e2e/README.md) (and [`e2e/AGENTS.md`](./e2e/AGENTS.md), which auto-loads for agents working in `e2e/`).

## Documentation map

Where to find / update docs:

| Topic | Location |
|---|---|
| Development process (spec→implement→audit→merge) | `docs/process.md` |
| Per-chunk-of-work record (scope, goals, spec; frozen at merge) | `docs/spec/<timestamp>-<name>/` |
| Product vision & philosophy | `docs/product/vision.md` |
| Backlog (features and bugs) | `docs/backlog/` |
| Product features (present-state, what + why) | `docs/product/` |
| Domain model & glossary | `docs/domain/` |
| Testing strategy & gate | `docs/testing.md` |
| E2E authoring, debugging, and suite rules | `e2e/README.md` (auto-loads `e2e/AGENTS.md`) |
| Architecture rules | `docs/architecture/` |
| Cross-cutting data model | `docs/architecture/data-model.md` |
| Design rules | `docs/design/` |
| Decision log (ADRs) | `docs/adr/` |
| Per-module behaviour, entities, key types | `src/internal/<module>/README.md` |
| Per-module agent rules | `src/internal/<module>/AGENTS.md` |
| External integrations (auth, constraints, API shape) | `src/internal/<spotify\|musicbrainz\|discogs>/README.md` |

### Synchronized content

A few topics intentionally live in more than one place. **Edit every listed location when changing any of them:**

- **Data model** — cross-cutting design decisions live in `docs/architecture/data-model.md`; per-entity meaning and key types live in each owning module's `README.md`; the domain glossary in `docs/domain/README.md` is canonical for term meaning. When adding, renaming, or removing an entity, update all three.
- **Design tokens** — token and utility definitions live in `static/src/main.css` (truth); their conceptual roles live in `docs/design/design-system.md`. Update the doc when a token group or named-role utility changes, not when individual values shift.

Anything else that ends up duplicated should be removed from one location, not kept in sync.

### Working docs and durable outputs

The working docs for a chunk of work — scope, goals, spec, supporting documents — are **committed** under [`docs/spec/<timestamp>-<name>/`](./docs/spec/), reconciled to what shipped, and frozen at merge (see [`docs/spec/README.md`](./docs/spec/README.md)). Genuinely throwaway scratch (raw exploration, dead-end notes) still stays outside the repo (`/tmp`, …) and is never committed.

A spec folder is a point-in-time record, **not** a canonical home. During Implement, durable outputs are still codified into their canonical homes — that's what keeps the current-state docs live and auto-loadable:

| Type of output | Canonical home |
|---|---|
| A reusable architectural rule | `docs/architecture/` (or a module's `AGENTS.md`) |
| A reusable design rule or token | `docs/design/` (and `static/src/main.css` if applicable) |
| User-facing behaviour of a feature | the owning module's `README.md` |
| A decision worth preserving the "why" of | `docs/adr/NNNN-short-slug.md` |
| A subtle invariant a refactor could silently break | a doc-comment next to the code it guards |
| A known architectural divergence | `docs/architecture/known-gaps.md` |
| Backlog items or future direction | `docs/backlog/` |

The spec folder is preserved beside these as the narrative of how the work landed — not a substitute for them.

## Documentation practices

- **Structure over prose; compress what's left.** Prefer bullets, tables, diagrams, and code blocks over paragraphs — docs should be scannable, not read line-by-line. Where prose is unavoidable, keep it compact: cut filler, preamble, and hedging so every sentence earns its place. A wall of prose is a rewrite signal — run `/prose-compact` to tighten it without losing facts.
- After editing or adding significant logic to a module, review and update the module's README if needed.
- Keep READMEs focused on high-level concepts, behaviour, and boundaries.
- After editing a module, review its `AGENTS.md` and update the module-specific notes if anything changed. Keep it tight — it's auto-loaded into context.
- A module's `AGENTS.md` describes **current state only**. No historical context, no forward-looking "should eventually" plans, no comparative claims about other modules. History lives in commit messages. If a module is mid-migration and temporarily non-compliant, a brief transitional note is acceptable until the migration lands.
- Avoid exhaustive lists or overly specific descriptions of package contents that will become outdated as code evolves. The "no exhaustive lists" rule in [`docs/architecture/AGENTS.md`](./docs/architecture/AGENTS.md) and [`docs/design/AGENTS.md`](./docs/design/AGENTS.md) applies equally to module READMEs and to `docs/`.
- Only add inline code comments when they provide context not evident from the code itself.
- Avoid comments that simply restate what the code does.
- If you notice a pattern or convention that should be documented here, ask the user if it should be added to this file.

### Doc references in module AGENTS.md

Every module `AGENTS.md` links to the docs an agent needs when working in that module. Three optional sections, each a table with **what** (one phrase) and **when** (one phrase):

- **Domain docs** — `docs/domain/` files for this module's domain model; open when reading/writing domain entities or rules. Add for every module that has a domain doc (including stubs — a pointer is useful even before the doc is filled in).
- **Product docs** — `docs/product/` files for features this module implements; open when deciding behaviour or how a flow should work. Omit where no product doc exists yet.
- **Relevant ADRs** — decisions in `docs/adr/` that directly constrain this module's behaviour; open when questioning a rule or constraint. Only link ADRs that are load-bearing for the module — skip cosmetic or naming decisions that don't affect implementation.

Module-specific notes should **reference** these docs rather than restate their content. If an implementation detail (e.g. a function name or migration reference) sheds light on why a doc is relevant, a brief note is fine; full restatement is not.

```markdown
## Domain docs

| Doc | What | When |
|---|---|---|
| [library](../../../docs/domain/library.md) | ownership state machine, radar eligibility rules | working on ownership transitions or radar logic |

## Product docs

| Doc | What | When |
|---|---|---|
| [library](../../../docs/product/library.md) | user-facing feature description | deciding what to expose or how a flow should behave |

## Relevant ADRs

| ADR | Decision | When |
|---|---|---|
| [ADR 0005](../../../docs/adr/0005-radar-eligibility-excludes-only-owned-wishlisted.md) | removed albums are radar-eligible | questioning why a removed album can be re-radared |
```