# tags — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Tag normalization (lowercase, trim, strip non-letter/digit/`-&`) is a private helper in `service.go` — domain logic, not a utility.
- After saving tags, the handler broadcasts the `album-changed` HTMX event (via `httpx.SetHXTrigger`, detail `{"albumId": <id>}`) instead of rendering library views; library owns the refresh via its `GET /app/library/album-surfaces` endpoint.

## Product docs

| Doc | What | When |
|---|---|---|
| [tags](../../../docs/product/tags.md) | user-defined tag groups and vocabulary for organizing albums | deciding tag behaviour or how tagging should work |
