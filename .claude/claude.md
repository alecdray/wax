# Claude Guidelines

## Code Generation

### Templates
- After modifying `.templ` files, run: `task build/templ`
- Generated files end in `_templ.go`

### Database
- After modifying `.sql` files in `db/queries/`, run: `task build/sqlc`
- After creating migrations in `db/migrations/`, run: `task db/up`
- Use `task db/create -- migration_name` to create new migrations

## Architecture Patterns

`src/internal/` is organized by archetype. Every directory under `src/internal/` has a `CLAUDE.md` declaring its archetype (or, for `server` and `core`, documenting singleton rules). When working in a module, the relevant rules will auto-load with that directory's `CLAUDE.md`.

Full rules: [`docs/architecture/`](../docs/architecture/).

For agents adding new code:
- New code under `src/internal/<module>/` must follow the rules of that module's archetype. Read the module's `CLAUDE.md` first.
- New modules: pick an archetype before writing code. If unsure, see [`docs/architecture/README.md`](../docs/architecture/README.md).

## Design

Every `.templ` file is one of three design archetypes (page templ, fragment templ, primitive), determined by location. Cross-cutting design principles (HTMX-first, fragments over pages, inline errors, theme tokens) and the visual vocabulary (Tailwind + DaisyUI `wax` theme) live alongside the archetype docs.

Full rules: [`docs/design/`](../docs/design/).

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

- **Strategy, unit-test conventions, the dev flow, and the gate** — [`docs/testing.md`](../docs/testing.md).
- **E2E suite rules + 8-step authoring recipe** — [`e2e/README.md`](../e2e/README.md) (and [`e2e/CLAUDE.md`](../e2e/CLAUDE.md), which auto-loads for agents working in `e2e/`).

## Documentation map

Where to find / update docs:

| Topic | Location |
|---|---|
| Product vision & philosophy | `docs/vision.md` |
| Roadmap, ideas, open questions | `docs/roadmap.md` |
| Operational follow-ups | `docs/backlog.md` |
| Testing strategy & gate | `docs/testing.md` |
| E2E authoring, debugging, and suite rules | `e2e/README.md` (auto-loads `e2e/CLAUDE.md`) |
| Architecture rules | `docs/architecture/` |
| Cross-cutting data model | `docs/architecture/data-model.md` |
| Design rules | `docs/design/` |
| Decision log (ADRs) | `docs/adr/` |
| Per-module behaviour, entities, key types | `src/internal/<module>/README.md` |
| Per-module agent rules | `src/internal/<module>/CLAUDE.md` |
| External integrations (auth, constraints, API shape) | `src/internal/<spotify\|musicbrainz\|discogs>/README.md` |

### Synchronized content

A few topics intentionally live in more than one place. **Edit both when changing either:**

- **Data model** — cross-cutting design decisions live in `docs/architecture/data-model.md`; per-entity meaning and key types live in each owning module's `README.md`. When adding, renaming, or removing an entity, update both.
- **Design tokens** — token and utility definitions live in `static/src/main.css` (truth); their conceptual roles live in `docs/design/design-system.md`. Update the doc when a token group or named-role utility changes, not when individual values shift.

Anything else that ends up duplicated should be removed from one location, not kept in sync.

### Working artifacts (not committed)

Spec, plan, and research files produced by skills (`/build`, `/to-issues`, `/grill-me`, etc.) are scratch artifacts. They live under `tmp/` (gitignored) and **must not be committed**. When the work merges, fold any durable learnings into the appropriate permanent home:

| Type of learning | Goes to |
|---|---|
| A reusable architectural rule | `docs/architecture/` (or a module's `CLAUDE.md`) |
| A reusable design rule or token | `docs/design/` (and `static/src/main.css` if applicable) |
| User-facing behaviour of a feature | the owning module's `README.md` |
| A decision worth preserving the "why" of | `docs/adr/NNNN-short-slug.md` |
| Operational follow-up | `docs/backlog.md` |
| Future direction | `docs/roadmap.md` |

If a learning doesn't fit any of these, it probably isn't worth persisting — let it die with the working file.

## Documentation practices

- After editing or adding significant logic to a module, review and update the module's README if needed.
- Keep READMEs focused on high-level concepts, behaviour, and boundaries.
- After editing a module, review its `CLAUDE.md` and update the module-specific notes if anything changed. Keep it tight — it's auto-loaded into context.
- A module's `CLAUDE.md` describes **current state only**. No historical context, no forward-looking "should eventually" plans, no comparative claims about other modules. History lives in commit messages. If a module is mid-migration and temporarily non-compliant, a brief transitional note is acceptable until the migration lands.
- Avoid exhaustive lists or overly specific descriptions of package contents that will become outdated as code evolves. The "no exhaustive lists" rule in [`docs/architecture/CLAUDE.md`](../docs/architecture/CLAUDE.md) and [`docs/design/CLAUDE.md`](../docs/design/CLAUDE.md) applies equally to module READMEs and to `docs/`.
- Only add inline code comments when they provide context not evident from the code itself.
- Avoid comments that simply restate what the code does.
- If you notice a pattern or convention that should be documented here, ask the user if it should be added to this file.
