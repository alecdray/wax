# auth — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Login orchestration only — `HttpHandler` composes `spotify.AuthService`, `user.Service`, and `feed.Service` to bootstrap the user on first Spotify callback.
- Routes `/`, `/logout`, and `/spotify/callback` are registered in `server/server.go` (root mux, not `/app`).
- NOT YET COMPLIANT: no `service.go` or `repo.go` — everything lives in `http.go` and `components.templ` at the module root. See `docs/architecture/refactor-backlog.md`.
