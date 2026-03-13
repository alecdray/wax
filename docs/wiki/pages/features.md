---
description: >
  What Wax can do today — shipped, production features and their scope. Belongs here: feature
  descriptions, user-facing behaviour, and constraints of live functionality. Does not belong here:
  planned or in-queue work (→ roadmap), implementation internals (→ architecture), or data
  structure definitions (→ data-model).
links:
  - "[roadmap](roadmap.md)"
  - "[vision](vision.md)"
  - "[data-model](data-model.md)"
  - "[frontend](frontend.md)"
  - "[integrations](integrations.md)"
---

[wiki](../wiki.md)

# Features

Shipped features in production today.

## My Library

The core of the app. A user's library is their collection of music — albums, artists, tracks, and releases (format variants: digital, vinyl, CD, cassette).

- Digital media is automatically synced from Spotify on a recurring schedule
- Users can browse and sort their library by title, artist, rating, date added, and last played
- Albums open in Spotify directly from the library
- All albums load at once on the dashboard — known to cause poor performance on mobile

A carousel above the library table offers two togglable views for surfacing albums worth acting on:

| View | What it shows |
|---|---|
| **Recently Spun** | Albums from [listening history](#listening-history), in reverse-chronological order; default view |
| **Unrated** | Albums in the library with no [rating](#rankings--reviews) yet — a prompt to rate what you've been playing |

Each carousel item shows album art, title, and artist, and links directly to the album in Spotify.

---

## Listening History

Wax records what you've been playing by polling Spotify's recently played tracks every hour in the background.

- Each track play is stored with a timestamp and linked to its album and artist
- Last played time per album is derived from play history and surfaces as a sort option in the library
- Sync runs hourly for all users with an active Spotify connection; token failures are handled gracefully (that user is skipped, the job continues)
- Because Spotify's API returns only the last 50 recently played tracks, syncing frequently is important — gaps can occur during long sessions or if the sync falls behind (see [integrations](./integrations.md) for the full constraint)

---

## Rankings & Reviews

Users can rate and review albums in their library.

- 0–10 score per album, manually set or guided by a 3-question questionnaire (scoring approach inspired by [Pitchfork](https://pitchfork.com/news/how-to-rate-albums-using-pitchfork-scores/))
- Free-text review notes attached to the rating
- Rating and review are independently updatable

### Rating Questionnaire

The questionnaire produces a recommended score from three questions, each capturing a distinct dimension of the listening experience:

| Question | What it measures |
|---|---|
| **Track consistency** — "How would you describe the album track by track?" | Whether the album holds up across its runtime; skips and weak tracks pull this down |
| **Emotional impact** — "How did this album make you feel while listening?" | The primary differentiator between good and great; measures absorption, not just enjoyment |
| **Gut check** — "When the album ended, what was your immediate reaction?" | The overall impression beyond individual tracks — relief, satisfaction, or the urge to restart |

Emotional impact carries the most weight because an album full of consistent, pleasant tracks can still leave you cold — while a flawed record with genuine emotional pull tends to be the one you remember.

### Rating Scale

Every score maps to a label. The labels are intentionally opinionated — they describe a relationship with the album, not a letter grade.

| Score | Label | What it means |
|---|---|---|
| 0–2.9 | **DOA** | Active rejection; nearly impossible via the questionnaire (floor is 2.0), so usually a manual override |
| 3.0–3.9 | **Nope** | Genuinely disliked — not just indifferent, a real negative reaction |
| 4.0–5.9 | **Not For Me** | The widest band; the album may be objectively fine but doesn't resonate — could be genre, mood, or just not clicking |
| 6.0–6.4 | **Lukewarm** | Faint praise — you finished it, maybe found some moments, but it didn't stick |
| 6.5–6.9 | **Solid** | A clear positive without being a standout; good enough to recommend with a caveat |
| 7.0–7.9 | **Staff Pick** | Actively good; worth telling someone about |
| 8.0–8.9 | **Heavy Rotation** | Something you return to — essential listening |
| 9.0–9.9 | **Instant Classic** | Near-perfect; exceptional across the board |
| 10.0 | **Masterpiece** | The absolute ceiling; reserved for records that are genuinely transformative |

The narrow bands at 6.x (Lukewarm / Solid) reflect the reality that many albums cluster in the "decent but not memorable" zone — splitting that range gives more resolution where most scores land.

---

## Tagging

Users can apply custom tags to albums for flexible organization and discovery.

- Tags belong to tag groups or stand alone
- No limit on tags per album or tags per group
- Two built-in tag group concepts: **Sound** (genre, style, influences) and **Mood** (context, feeling, occasion)
- Users define their own tags within these groups

