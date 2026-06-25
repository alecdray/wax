# Primary genres are an app-curated facet derived from the Wikidata genre graph

Wax assigns every album zero or more **primary genres** — a small, fixed, app-curated set of broad buckets (rock, pop, electronic, jazz, metal, …) — for coarse library filtering. Genres are a distinct annotation from tags: a shared global taxonomy, auto-derived rather than authored, additive to the free-form user tag vocabulary rather than replacing it.

Genres are sourced by resolving an album's Discogs genre/style terms to nodes in a Wikidata-derived genre graph, then crawling each node's ancestors. An album's primaries are the **most-specific** allowlisted nodes-or-ancestors on each path: where matched genres nest — metal is a descendant of rock in the graph — the deeper bucket wins, so a metal album lands in *metal*, not *rock*; an album spanning unrelated branches (hyperpop → pop + electronic) keeps one primary per branch.

Curated rather than read off the graph's own shape: the graph's top level is abstract structural categories (popular, vocal, experimental), and the recognizable buckets sit at varying depths, so no depth cut yields {rock, pop, electronic, …}. The allowlist is the source of truth; the graph supplies only the ancestor relationships that map sub-genres up. Primaries are computed from stored leaf genres at read time, so changing the allowlist never re-derives stored data.

Enrichment runs as a background task, giving the whole library coverage without blocking sync or page loads. Each album records that it was processed, so an unresolved album (queried, nothing matched) is distinguishable from an unprocessed one and is surfaced under an uncategorized filter bucket.

Rejected: deriving the bucket set from the graph's top level (yields abstract non-genres); keeping every matched ancestor (broad buckets swallow their sub-buckets); replacing the user tag system (the two are different kinds of annotation, so genres are additive).
