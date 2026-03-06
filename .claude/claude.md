# Claude Guidelines

## Code Generation

### Templates
- After modifying `.templ` files, run: `templ generate`
- Generated files end in `_templ.go`

### Database
- After modifying `.sql` files in `db/queries/`, run: `sqlc generate`
- After creating migrations in `db/migrations/`, run: `task db/up`
- Use `goose -dir db/migrations create migration_name sql` to create new migrations

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
- Environment variables documented in `.env.template`
- Run `task` without arguments to list available commands

## Documentation

- After editing or adding significant logic to a module, review and update the module's README if needed
- Keep README focused on high-level concepts, architecture, and workflows
- Avoid exhaustive lists or overly specific descriptions of package contents that will become outdated as code evolves
- Only add inline code comments when they provide context not evident from the code itself
- Avoid comments that simply restate what the code does
- If you notice a pattern or convention that should be documented here, ask the user if it should be added to this file
