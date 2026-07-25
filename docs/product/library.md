# Library

**What it is:** your album collection — own albums across formats, track wishlist items, browse and filter your library. The central surface of wax.

## Why

Streaming platforms flatten your music into a single undifferentiated list. The library gives you ownership over your collection: you decide what belongs, in what formats, and how it's organized. It's the home you return to — not a feed you scroll past.

## What a user can do

| Action | Result |
|---|---|
| Own an album in a format (vinyl, CD, digital, cassette) | It appears in your library with that format badge. |
| Wishlist an album | It appears in your library, marked wishlist — an intent to acquire. |
| Remove an album | It disappears from your library. You can re-add it later. |
| Browse by format filter | Library shows only albums owned in the selected format. |
| Sort by artist, title, date added, rating | Library reorders to match. |
| Open an album detail | Full view with all formats, ratings, tags, notes, and tracklist. |
| Sync Spotify saved albums | Albums saved on Spotify appear as owned in your library. |

## Headline rules

- An album appears in your library when at least one of its formats is owned or wishlisted.
- Ownership is per-format — you can own the vinyl and wishlist the CD of the same album.
- Removing an album clears all format ownership for it. Re-adding starts fresh.
- Synced albums from Spotify are owned in a "streaming" format with no physical equivalent — they behave like any other owned album in the library.
- **Radar clears on own/wishlist.** Owning or wishlisting an album removes it from your radar — you've acted on it, so it leaves the watch list.
- **Auth and user management** (login, sessions, Spotify OAuth) are infrastructure, not product features — see the [auth](../../src/internal/auth/AGENTS.md) and [user](../../src/internal/user/AGENTS.md) module docs.