# notes — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Pure data + markdown-rendering service for album notes; no HTTP entrypoints. Persistence type is `AlbumNote`; "sleeve note" is a UI label only — don't introduce parallel naming.

## Domain docs

| Doc | What | When |
|---|---|---|
| [notes](../../../docs/domain/notes.md) | album annotation model | working on note reads/writes |

## Product docs

| Doc | What | When |
|---|---|---|
| [notes](../../../docs/product/notes.md) | freeform album annotations | deciding annotation behaviour or how notes should work |
