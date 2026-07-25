# Library — domain

The user's relationship to albums through ownership, wishlisting, removal, and radar monitoring.

## Ownership state machine

A user-album relationship exists in one of three ownership states:

| State | Meaning |
|---|---|
| Owned | The user possesses the album in at least one format. |
| Wishlisted | The user intends to acquire the album. |
| Removed | The album was previously owned or wishlisted but has been removed. |

An album appears in the user's library when at least one of its formats is owned or wishlisted. A removed album disappears from the library.

### State transitions

```
          +---------+
          |  none   |  (no relationship row)
          +---------+
          /         \
    own/wishlist     remove
       /               \
      v                 v
+-----------+     +-----------+
| owned /   |     |  removed   |
| wishlisted|     +-----------+
+-----------+          ^
      |                |
      +--- remove ------+
```

- **Own/Wishlist** (from none or removed): creates an ownership row for the specified format. If this is the first owned/wishlisted format, the album enters the library.
- **Remove** (from owned/wishlisted): deletes all format ownership rows. The album leaves the library.
- An album can be owned in one format and wishlisted in another simultaneously — each format is an independent row.
- Re-owning or re-wishlisting a removed album starts fresh — no prior state is restored.

## Releases (formats)

A release is a specific format of an album (vinyl, CD, digital, cassette, streaming). Ownership is per-release:

- Each release carries its own ownership state (owned or wishlisted) for the current user.
- The streaming format is used for Spotify-synced albums — it behaves like any other format.
- An album's display format badges are derived from which releases the user owns or has wishlisted.

## Radar

A radar entry is a "watching this" bookmark on an album. It is independent of ownership — a user can radar an album they have no other relationship with.

### Eligibility rule

An album is radar-eligible unless it is **currently owned or wishlisted**. A removed album is radar-eligible ([ADR 0005](../adr/0005-radar-eligibility-excludes-only-owned-wishlisted.md)).

### Auto-clear on ownership change

When a user owns or wishlists an album, its radar entry is **automatically cleared** — the user acted on it, so it leaves the watch list. This is the library domain's single cross-domain write: acquiring or wishlisting an album clears the radar entry as a side effect.

### Radar sources

Albums reach the radar through two paths:
- **In-app add**: the user searches and explicitly adds an album to radar.
- **Spotify inbox**: the user opts into a Spotify playlist watched by wax. A periodic feed sync ingests each track as a radar entry. Tracks whose album is already owned or wishlisted are silently dropped ([ADR 0004](../adr/0004-spotify-radar-playlist-entry.md)).