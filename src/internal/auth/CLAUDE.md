# auth — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Login orchestration: `HttpHandler` composes `spotify.AuthService`, `user.Service`, and `feed.Service` to bootstrap the user on first Spotify callback.
- Routes `/`, `/logout`, and `/spotify/callback` mount on the root mux (not `/app`) since they're unauthenticated.
