# Navigation is a bottom bar on all viewports

Primary navigation is a fixed bottom bar shown on every authenticated page and at every viewport size — phone and desktop alike — replacing the former sticky top header's navigation role. It carries the top-level destinations plus an account menu, and is a domain-free primitive owned by the shared layout.

Wax is mobile-first, and a thumb-reachable bottom bar is the native pattern there. Rather than maintain a second desktop-only top bar, the same bar is kept on wider screens — its tappable row simply caps its width and centres — trading a little desktop convention for one navigation to build, test, and reason about. A responsive top-on-desktop / bottom-on-mobile split was rejected as two divergent chrome layouts for no real gain.

Identity and status do not ride on the navigation bar. A slim top header carries the wax wordmark on the left and the feed sync-status control on the right, on every authenticated library page. Because that header carries feed data it is library-owned, not a core primitive — the split keeps navigation domain-free while giving feed status one consistent, always-visible home. An earlier cut pushed feed status into individual page surfaces instead; the control then moved and disappeared between pages, which is what the header restores.
