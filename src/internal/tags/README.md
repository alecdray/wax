# tags

User-defined tags applied to albums, optionally organized into tag groups.

## Responsibility

`tags` owns the user's tag vocabulary and the relationship between tags and albums. A `TagDTO` is a normalized, lowercase label that can belong to a `TagGroupDTO` (e.g. *Sound* or *Mood*). The module handles tag normalization, get-or-create on write, and bulk lookup of tags by album for rendering in `library`'s album views.

## Key types

- `TagDTO` — one tag with optional `*TagGroupDTO`. Names are normalized (lowercase, trimmed, alphanumerics + `-` and `&` only).
- `TagGroupDTO` — a named bucket of tags scoped to one user.
- `TagInput` — write-side shape: a raw name plus optional group ID, taken by `SetAlbumTags`.

## Boundaries

- **Inbound:** consumed by `library.Service` (so `AlbumDTO.Tags` populates) and by `tags/adapters` (the tag editor modal).
- **Outbound:** none — the module depends only on `core/db`.
- **Adapters:** own the `/app/tags/album` modal and form submission. The handler also calls into `library.Service` to load the album, and `discogs.Service` for genre suggestions.

## Background tasks

None.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
