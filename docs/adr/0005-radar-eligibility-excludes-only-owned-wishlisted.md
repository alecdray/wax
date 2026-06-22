# Radar eligibility excludes only owned and wishlisted albums

An album is eligible for the radar unless the user currently owns or wishlists it. A `removed` release no longer blocks the radar: an album the user previously discarded can be put back on the radar, and a radar entry may coexist with a `removed` release for the same album. Bringing the album into the library — owning or wishlisting it — still clears the radar entry.

The radar has always been defined as a watch state for albums *not in the library*, where the library is the owned-or-wishlisted set. The implementation was stricter than that definition: it excluded albums with *any* release record, `removed` included, so a discarded album could never return to the radar even though, by that definition, it was not in the library. Allowing albums to be enqueued for the radar from a Spotify-side inbox ([ADR 0004](0004-spotify-radar-playlist-entry.md)) forced the question, and we chose to align the rule with the original definition rather than entrench the stricter behaviour — putting a discarded album back on the radar is a legitimate "reconsider this" action, and the rule must be uniform across every entry point rather than vary by where the album was added.

The cost is the new coexistence of a radar entry with a `removed` release, which the radar surfaces like any other watched album.
