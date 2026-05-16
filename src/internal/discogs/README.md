# discogs

A metadata source for physical media. Wax queries Discogs to identify and enrich vinyl, CD, and cassette releases that aren't well represented in Spotify.

## Role

Discogs provides catalog data (label, release date, format details) and a stable Discogs ID for physical releases, plus genre and style information that's normalized through the genre DAG before being surfaced as tag suggestions.

## Constraints

- Discogs returns compound genre and style strings (e.g. `"Funk / Soul"`, `"Folk, World, & Country"`); these are split and resolved against the wax genre DAG, and unrecognised tokens are dropped silently.
- Genre suggestions are best-effort. Search failures are logged and swallowed; callers always receive a (possibly empty) slice.

## See also

- Architecture rules: [`../../../docs/architecture/archetypes/external-client.md`](../../../docs/architecture/archetypes/external-client.md)
- Module-specific notes: [`./CLAUDE.md`](./CLAUDE.md)
