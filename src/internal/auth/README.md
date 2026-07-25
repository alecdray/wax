# auth

User authentication: JWT session issuance, login orchestration, and the Spotify OAuth callback flow.

## Responsibility

`auth` owns session issuance and the credential store for third-party tokens. It is the single entry point for authentication — every authenticated request flows through a JWT validated by this module.

## Session model

- Sessions are JWTs issued on login and validated on every authenticated request via middleware in `server/`.
- The JWT carries the user ID and expiry; no server-side session store.
- The Spotify OAuth callback flow is owned here — it exchanges the authorization code for tokens, persists the encrypted credential via the [user](../user/README.md) module, and issues the application JWT.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./AGENTS.md`](./AGENTS.md)