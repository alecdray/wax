# Radar

**What it is:** a watch list for albums you're interested in but haven't acted on yet — a holding pen between discovery and your library.

## Why

You hear about an album you want to check out, but you're not ready to own or wishlist it. Streaming platforms give you a like button; radar gives you a deliberate, scannable list that clears itself when you decide. It's the inbox for your music curiosity.

## What a user can do

| Action | Result |
|---|---|
| Search and add an album to radar | It appears on your radar page. |
| Own or wishlist a radar album | It moves to your library; radar entry is cleared. |
| Remove an album from radar | It disappears from your radar list. |
| Opt into Spotify radar inbox | A Spotify playlist becomes your radar feed — saved tracks appear as radar entries. |
| Browse radar by recently added or artist | Radar reorders to match. |

## Headline rules

- An album is **radar-eligible** unless it is currently owned or wishlisted. A removed (formerly owned) album can go back on the radar.
- Owning or wishlisting an album **clears its radar entry** automatically — no manual cleanup needed.
- The **Spotify radar inbox** is opt-in. The user creates a playlist on Spotify, wax watches it, and each track saved there becomes a radar entry. An inbox track whose album is already owned or wishlisted is silently dropped.
- Radar sync runs periodically with exponential backoff on failure — a transient Spotify outage won't hammer the API, and a revoked token won't retry forever.
- Radar lives on its own page under the library, not as a separate top-level section.