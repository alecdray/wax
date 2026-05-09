# notes — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Persistence + markdown rendering for album notes. Persistence type is `AlbumNote`; "sleeve note" is a UI label only — don't introduce parallel naming.
- Consumed by `library/adapters/` to render sleeve notes inline on the album view; library calls `notes.Service` for upsert/fetch and `notes.RenderMarkdown` for HTML.
