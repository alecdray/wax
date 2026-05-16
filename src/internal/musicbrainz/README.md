# musicbrainz

A secondary metadata source used to enrich what Spotify provides. No authentication required.

## Role

Queried by name (title + artist) to find canonical recordings that Spotify doesn't expose or exposes ambiguously. Results are matched fuzzily against the requested title and artist credit; ambiguous or missing matches return nil rather than guessing.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/external-client.md`](../../../docs/architecture/archetypes/external-client.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
