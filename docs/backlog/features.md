# Features

Candidate features and improvements. Priority column: **high** / **mid** / **low** / **—** (unprioritized). Each ships through the [development process](../process.md).

## In queue

| Priority | Feature | Summary |
|---|---|---|
| high | **Crates** *(shipped)* | Named, unordered collections — "jazz," "road trip music," "gifts for friends." Membership is the point, not position. Distinct from tags: crates are first-class objects with their own pages; tags are lightweight inline labels. Spec: `docs/spec/1785103459-album-groups/`. |
| high | **Wishlist surfaces** | Physical media wishlist — mark an album as "I want to buy this on vinyl/CD" and view those albums in a dedicated list. |
| mid | **Ranked Lists** | User-curated ordered lists where position is the point — "my top 10 of the year," "albums to hear before you die." Order is explicit and meaningful. |
| mid | **Hidden Albums** | Soft-remove albums from the main library view without deleting them (e.g., podcasts or junk synced from Spotify) |
| mid | **Scroll position restore on back navigation** | Navigating from the library to an artist detail page and pressing back returns to the top of the library instead of the previous scroll position. |
| mid | **Infinite scroll: earlier load trigger** | The library only loads more albums once the user reaches the very bottom. With large libraries this creates a visible gap between scroll-stop and new content — the load should trigger earlier (e.g., within one viewport of the bottom). |
| low | **Stats & Insights** | Analytics across library, listening history, ratings, ranked lists, and album groups |
| low | **Notifications: in-app event feed** | In-app notifications for background events (sync completions, activity) |
| low | **Notifications: rate-limit toast** | When the shared Spotify guard is paused, user-triggered actions fail silently (HTMX doesn't swap 4xx). Needs an app-wide transient-error display mechanism and a "Spotify is rate-limiting us, try again shortly" toast. |
| low | **Linked Albums** | Connect albums to each other, building a personal music graph |
| low | **Album Detail — Non-Library Albums** | A read-only detail view for albums not in the user's library; fixes carousel items and discover results that currently fall back to Spotify outlinks |
| low | **Saved Tracks Sync** | Sync the user's saved tracks from Spotify so loved tracks can be highlighted within album views |
| low | **Auth-aware feed dormancy** | When a Spotify token is revoked (user disconnects Wax in Spotify settings), feeds back off silently with no UI feedback. Give feeds a dormant/needs-reconnect state for auth failures — stop auto-polling and surface a "reconnect Spotify" prompt. |

## Ideas

Directionally clear, not yet queued.

- **Comparative ranking** — derive a rating by pitting an album against others already rated; a series of "is this better than that?" questions produces a score grounded in relative preference.
- **Tag management** — dedicated interface for managing tags at the user level (create, rename, merge, delete) without going through individual albums.
- **Album Detail — external sources** — links to Pitchfork, Wikipedia, NPR, YouTube per album; users should be able to attach their own resource links (live performances, Tiny Desk concerts, interviews, articles).
- **Influences** — surface what influenced an album and what it influenced.
- **Shared by / shared with** — optional field to track who introduced you to a record.
- **Multiple view modes** — grid/cover wall, compact text-only, table — mode switcher in the library header, persisted per user.

## Open questions

- **Progressive Web App (PWA)** — whether to convert Wax to a PWA for offline support and installability; deferred until the mobile experience is more fully developed.
- **Social features** — Goodreads-style network is a natural long-term direction, but it's unclear whether this belongs in the core product or as a separate surface; secondary to personal library depth.
