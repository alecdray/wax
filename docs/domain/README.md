# Domain

The model of the **problem space** — organized by concept and mutation-ownership. It holds the domain language, the entities and their invariants, and the bounded domains (who is allowed to change what).

**Domain is the deep reference; [product](../product/) is the entry point.** A reader meets a feature in product — why it exists, what it does, its headline rules — then comes here for the complete, authoritative model: exact rules, edge cases, entity ownership, lifecycle. Product summarizes; domain is the full spec it summarizes.

**The test that keeps this doc honest: it is UI-agnostic and implementation-agnostic.** An entity and its rules exist the same way whether shown on a page, returned by an API, or printed by a CLI — and regardless of how the code is structured. If a sentence names a screen or flow, it belongs in [product](../product/); if it names how the code is built, it belongs in [architecture](../architecture/) or a module README. Domain is the detailed *what and why-rules*, never the *how-it's-built*.

Every term with a specific meaning is defined once here (in the glossary) and linked from wherever it is used — never re-explained in context. When a term's meaning changes, update it here; references follow automatically.

## Structure

Shared, cross-cutting content lives in this README; each **bounded domain gets its own file** (`docs/domain/<domain>.md`) so no single doc grows unbounded.

- **This README** — the entities (shared nouns), the glossary (shared language), the cross-domain write ledger (spans domains by definition), and the index of per-domain docs below.
- **`docs/domain/<domain>.md`** — one file per bounded domain: the fields it owns, its rules/policies, state machines, and operations.

## Entities

The shared nouns — what exists in the domain. Each is cross-domain (any domain may read it); exactly one domain writes each field.

| Entity | What it is |
|---|---|
| Album | A musical release. The central anchor — almost every user interaction is scoped to an album. |
| Release | A specific format of an album (vinyl, CD, digital, cassette, streaming). |
| Rating | A user's scored assessment of an album (0.0–10.0), backed by an append-only rating-log. |
| Tag | A label in a user-defined tag group, assigned to an album. |
| TagGroup | A user-defined category of tags (e.g. "mood," "era"). |
| Note | Freeform text annotation on an album. |
| Genre | A curated, app-owned primary genre bucket, auto-assigned to albums from Discogs data. |
| RadarEntry | An album on the user's radar — watched but not owned or wishlisted. |
| InboxItem | A Spotify-side track awaiting ingestion into radar, per a user's opt-in inbox playlist. |
| ListeningRecord | A timestamped record of the user playing an album. |
| User | A person with a wax account, owning all of the above. |
| Session | A JWT session representing an authenticated user. |

## Glossary

Terms with specific or confusable meanings — defined once, linked from everywhere else.

| Term | Meaning |
|---|---|
| Owned | The user possesses the album in at least one format. |
| Wishlisted | The user intends to acquire the album; it appears in the library marked as such. |
| Removed | The album was previously owned or wishlisted but has been removed from the library. No format relationships survive removal. |
| Radar-eligible | An album that is neither owned nor wishlisted. A removed album is radar-eligible. |
| Provisional rating | A rating the user has saved but not yet finalized — intended to be revisited. |
| Finalized rating | A settled rating — the user's definitive score for the album. |
| Release vs. Format | A release is a specific format (vinyl, CD, etc.) of an album. The user's ownership is per-release. |
| Tag vs. Genre | Tags are user-defined and per-user. Genres are app-curated, shared, and auto-assigned. Tags have no global taxonomy; genres are the deliberate exception. |
| Primary genre | One of ~20 curated genre buckets derived from the Wikidata genre graph via Discogs data. |
| Spotify inbox | A user-created Spotify playlist that wax watches for radar candidates. |

## Domains

| Domain | Doc | Owns (writes) |
|---|---|---|
| Library | [library.md](library.md) | Albums, releases, ownership state, radar entries, album-surfaces |
| Review | [review.md](review.md) | Ratings, rating-log entries, rating lifecycle state |
| Genres | [genres.md](genres.md) | Primary genre taxonomy, album-genre assignments |
| Tags | [tags.md](tags.md) | Tag groups, tag assignments |
| Notes | [notes.md](notes.md) | Album annotations |
| Feed | [feed.md](feed.md) | Feed sync state, inbox playlist handles |
| Listening History | [listeninghistory.md](listeninghistory.md) | Play records |
| User | [user.md](user.md) | User profiles, settings |
| Auth | [auth.md](auth.md) | Session issuance, credential store |

## Cross-domain write ledger

Every write that crosses a domain boundary — named, with owning operation and reason.

| Write | Owning operation | Why it crosses |
|---|---|---|
| Ownership change clears radar entry | library — acquiring or wishlisting an album | Owning or wishlisting settles the radar question — the user acted, so the watch entry is obsolete. |
| Radar inbox sync creates radar entries | feed inbox sync → library radar-add | The feed ingests external data; library owns the radar relationship. |
| Saved-albums sync creates ownership | feed saved-albums sync → library ownership | Feed ingests from Spotify; library owns ownership. |
| Rating save/finalize broadcasts album-changed | review — saving or finalizing a rating | Library owns the album surfaces that re-render in response. |
| Genre auto-assignment from genregraph | genres — album enrichment | Genres derive assignments from the genre graph and Discogs data; no user action involved. |