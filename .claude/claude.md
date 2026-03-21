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

### Module Structure
- Service layer: Business logic in `Service` struct
- Adapters: HTTP handlers and templates in `adapters/` subdirectory
- Domain models: Core types in module root
- Check module README files for detailed documentation

### Context
- Use `contextx.ContextX` instead of `context.Context`
- Extract user ID with `ctx.UserId()`

### Error Handling
- Use `httpx.HandleErrorResponse()` for consistent error responses
- Return HTML fragments for HTMX error display

### HTMX
- Forms use `hx-post`, `hx-put` for submissions
- Responses are HTML fragments
- Use `hx-swap` to control content replacement
- Return error components for inline display

## Development

- Use `task` command for all build/run operations (see `taskfile.yml`)
- Always prefer `task <name>` over invoking tools directly (e.g. `task build/templ` not `templ generate`)
- Environment variables documented in `.env.template`
- Run `task` without arguments to list available commands

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
- Avoid exhaustive lists or overly specific descriptions of package contents that will become outdated as code evolves
- Only add inline code comments when they provide context not evident from the code itself
- Avoid comments that simply restate what the code does
- If you notice a pattern or convention that should be documented here, ask the user if it should be added to this file
