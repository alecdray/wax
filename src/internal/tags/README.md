# tags

User-defined tags applied to albums, optionally organized into tag groups.

## Responsibility

`tags` owns the user's tag vocabulary and the relationship between tags and albums. A tag is a normalized, lowercase label that can belong to a tag group (e.g. *Sound* or *Mood*). The module handles tag normalization, get-or-create on write, and bulk lookup of tags by album for rendering in `library`'s album views.

## Tagging surface

The tag editor opens from the album detail page; the library dashboard shows tags as read-only badges with no inline editing. Genre suggestions inside the editor come from `discogs.Service`.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
