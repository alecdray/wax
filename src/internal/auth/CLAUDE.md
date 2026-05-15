# auth — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Owns JWT issuance, login orchestration, and the Spotify OAuth callback flow. `Service` composes `spotify.AuthService`, `user.Service`, and `feed.Service` to bootstrap a user on first login.
- Routes (`/`, `/logout`, `/unauthorized`, `/spotify/callback`) mount on the root mux — they sit outside JWT middleware. `core/httpx` middleware redirects to `/unauthorized` on auth failure rather than rendering the page inline.
