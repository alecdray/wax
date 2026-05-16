# core

Shared infrastructure: framework-level utilities used by 2+ modules. There is exactly one `core/` — it has no archetype because an archetype describes a category and `core/` is unique.

`core/` holds no business concepts. If a sub-package mentions albums, ratings, tags, or users, it belongs in a domain module instead.

## Adding to core

- **Used by 2+ modules.** Single-consumer code stays in the consumer.
- **Framework-level, not domain.** No business concepts.
- **`x` suffix** marks extension packages over a stdlib counterpart.

## See also

- Singleton rationale and sub-package roster: [`./CLAUDE.md`](./CLAUDE.md)
- Architecture entry point: [`../../../docs/architecture/README.md`](../../../docs/architecture/README.md)
