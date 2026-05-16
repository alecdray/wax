# notes

User-authored notes attached to albums — "sleeve notes" in the UI. Markdown-based, one note per (user, album).

## Role

A free-form text surface a user can attach to any album in their library. Notes are stored as markdown and rendered to HTML on read; the same note is editable in place from the album detail page. There is exactly one current note per user/album — saving replaces the prior content.

## Constraints

- Notes have a length cap (`MaxSleeveNoteLength`).
- Markdown is rendered to HTML at read time; rendered links open in a new tab.
- The persistence type is `AlbumNote`; "sleeve note" is a UI label only — don't introduce parallel naming.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/domain-module.md`](../../../docs/architecture/archetypes/domain-module.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
