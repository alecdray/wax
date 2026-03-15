---
description: >
  The domain — what entities exist, what they represent, and how they relate to each other.
  Belongs here: entity definitions, relationships, and key data design decisions. Does not belong
  here: how entities are queried or stored (→ architecture), what users can do with them
  (→ features), or where data comes from (→ integrations).
links:
  - features
  - architecture
  - integrations
---

[Parent: wiki](../wiki.md)

# Data Model

The core domain entities and how they relate.

## Entities

### Music Catalog

| Entity | Description |
|---|---|
| **Album** | The primary unit. Holds metadata (title, art, release date) sourced from Spotify |
| **Artist** | A music artist, linked to one or many albums |
| **Track** | An individual track, belonging to an album |
| **Release** | A format variant of an album (digital, vinyl, CD, cassette) |

Albums → Tracks, Albums → Artists, Albums → Releases are all many-to-many or one-to-many relationships depending on context.

### User Library

A user's library is their personal collection of the above entities. The library represents *what they own or have saved*, not the global catalog.

- **User Releases** — releases a user owns
- **User Tracks** — tracks a user has saved
- **User Artists** — artists a user follows

### Annotations

User-generated content attached to library entities:

| Entity | Description |
|---|---|
| **Album Rating Log** | An append-only log of 0–10 rating entries for an album; each entry optionally includes a note and carries its own timestamp |
| **Tag Group** | A named category for organizing tags (e.g. Sound, Mood) |
| **Tag** | A user-defined label applied to albums, optionally grouped |
| **Album Tag** | Join between an album and a tag |

### Activity

| Entity | Description |
|---|---|
| **Track Play** | A record of when a user played a track (from Spotify history) |

### System

| Entity | Description |
|---|---|
| **User** | An account, authenticated via Spotify. Stores an encrypted Spotify refresh token |
| **Feed** | Tracks sync state for external data sources (e.g. Spotify library sync) |

## Relationships

```
User
 ├── User Releases → Release → Album
 ├── Album Rating Log → Album
 ├── Tag Groups → Tags → Album Tags → Album
 └── Track Plays → Track → Album

Album
 ├── Artists (many-to-many)
 ├── Tracks (one-to-many)
 └── Releases (one-to-many)
```

## Key Design Decisions

- **Album is the anchor** — almost every user interaction is scoped to an album
- **Releases model format** — the same record can exist as vinyl, digital, etc. under the same album
- **Tags are user-defined** — no global taxonomy imposed; users build their own vocabulary
- **Ratings are optional** — the library is valuable without ratings; annotations layer on top

