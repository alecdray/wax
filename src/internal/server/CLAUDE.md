# server — composition root (singleton)

This directory is the **composition root** of the application. There is exactly one of it; it has no archetype because an archetype describes a category and there is only one server.

*This is the target state. Server today still owns route registration inline in `Start`; see `docs/architecture/refactor-backlog.md` for the migration.*

## Responsibilities

- Build all services in `NewServices(app, db)` (manual DI).
- Set up the root `*httpx.Mux` and any sub-muxes (e.g. authenticated `/app/` sub-mux with JWT middleware, mounted via `rootMux.Use("/app/", appMux)`).
- Register cron tasks with the `core/task` task manager.
- Call each domain module's `adapters.RegisterRoutes(mux, handler)` to register routes — one call per module.
- Run lifecycle: open DB, start task manager, start HTTP listener, handle shutdown.

## Rules

- **No domain logic.** Server wires things together; it does not implement features.
- **No URL patterns** beyond mounting sub-muxes (`rootMux.Use("/app/", appMux)`). Concrete paths live in each domain module's `adapters/routes.go`.
- **Allowed imports:** every domain module, every external client, all of `core/*`. Server is the only place this is allowed.
- **No tests** in this package — the application is integration tested via e2e tests in `e2e/`.

## Why server is not an archetype

An archetype describes a category of modules with multiple instances and shared rules. There is exactly one server. Trying to fit it into the `utility` archetype would require carving out exceptions to utility's import rules (utility forbids importing domain modules; server requires importing all of them). A singleton is documented here directly.
