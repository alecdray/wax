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

### HTMX (frontend convention)
- Forms use `hx-post`, `hx-put` for submissions; responses are HTML fragments.
- Use `hx-swap` to control content replacement.
- Return error components for inline display.

## Development

- Use `task` command for all build/run operations (see `taskfile.yml`)
- Always prefer `task <name>` over invoking tools directly (e.g. `task build/templ` not `templ generate`)
- All `go build` commands must output to `./bin/` using `-o ./bin/<name>` — never build to the project root
- Environment variables documented in `.env.template`
- Run `task` without arguments to list available commands
- Worktrees don't have a `.env` — copy from the main project: `cp /Users/shmoopy/workshop/projects/wax/.env .env`
- Also run `npm install` in the worktree before `task dev` if `node_modules` is missing
- Copy the DB from the main project to avoid 500s from missing users: `cp /Users/shmoopy/workshop/projects/wax/tmp/db.sql ./tmp/db.sql`

## Testing

- **Always write tests for new logic.** Any non-trivial function should have a corresponding test.
- For pure logic functions in `cmd` packages, extract them into a separate file (e.g. `pull.go`) so they can be unit tested without I/O — keep `main.go` as thin orchestration only
- Group tests by the function under test using a top-level `Test<FuncName>` function
- Use `t.Run` subtests to describe specific behaviours — makes output scannable and serves as documentation
- Name subtests as plain descriptions of the expected behaviour (e.g. `"returns empty for nonsense query"`)
- Test behaviour, not implementation — each subtest should assert one specific outcome
- Use `t.Skip` when a test condition may legitimately not be met in all dataset states

## Documentation

- After editing or adding significant logic to a module, review and update the module's README if needed
- Keep README focused on high-level concepts, architecture, and workflows
- After editing a module, review its `CLAUDE.md` and update the module-specific notes if anything changed. Keep it tight — it's auto-loaded into context.
- A module's `CLAUDE.md` describes **current state only**. No historical context, no transitional "not yet compliant" notes, no forward-looking "should eventually" plans, no comparative claims about other modules. Compliance gaps and future work live in `docs/architecture/refactor-backlog.md`; history lives in commit messages.
- Avoid exhaustive lists or overly specific descriptions of package contents that will become outdated as code evolves
- Only add inline code comments when they provide context not evident from the code itself
- Avoid comments that simply restate what the code does
- If you notice a pattern or convention that should be documented here, ask the user if it should be added to this file
