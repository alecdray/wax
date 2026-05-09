# notes — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Pure data + markdown-rendering service for album notes; no HTTP entrypoints. Persistence type is `AlbumNote`; "sleeve note" is a UI label only — don't introduce parallel naming.
