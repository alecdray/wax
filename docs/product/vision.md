# Vision

**Wax** is a personal music library application built around a deeper, more intentional relationship with albums.

## What it is

A place to own your music identity — not just stream it. Wax lets you collect, annotate, rate, and reflect on your albums in ways that streaming platforms don't support. It's album-centric by design, treating records as objects worth thinking about rather than content to consume.

## Who it's for

Music listeners who want more than play counts and algorithmic recommendations. People who remember where they were when they first heard an album. Collectors — of vinyl, CDs, or digital libraries — who care about the *relationship* with music, not just the catalog.

## Tagline

*Listen with Wax.*

## Design philosophy

- **Album-first** — albums are the primary unit, not tracks or playlists
- **Personal over social** — the core value is your own library and annotations, not what others think
- **Mobile-first** — the primary surface is a phone in hand; desktop is the responsive variant, not the inverse
- **Progressive enhancement** — server-rendered HTML with targeted interactivity, not a heavy SPA
- **Depth over breadth** — fewer features done meaningfully, not a feature parity race with streaming platforms

## Aesthetic direction

The visual and interaction design draws from analog music equipment — turntables, tape machines, record sleeves. The feel should be warm and inviting: wood, warm light, the glow of a candle or fire. Vintage and analog in spirit, but modern and digital-first in execution. The goal is an interface that feels like a place you want to spend time, not a utility.

## Deployment context

Wax is a **private, self-hosted application** — not a public internet service. It runs on a private network (e.g. Tailscale) for a small number of trusted users: the owner and a handful of invited friends, or self-deployers running their own instance.

**What this means for implementation decisions:**

- **Threat model is a small trusted network, not the public internet.** SQL injection, auth bugs, and session security still matter — those are correctness issues regardless of who can reach the server. But public-internet hardening (CAPTCHAs, IP blocking, aggressive rate limiting for anonymous traffic, WAF rules) is out of scope.
- **Scale is a handful of concurrent users, not thousands.** Correctness and readability should be prioritised over premature optimisation. N+1 queries and unbounded result sets are still worth flagging where they could cause real latency or resource problems, but micro-optimisation for throughput is not a goal.
- **No anonymous public traffic.** Every active session belongs to a known, authenticated user. Features like account lockout, registration spam prevention, and public-facing abuse controls are not needed.
- **Self-deployment is a supported path.** Deployment instructions, environment variable documentation, and a reasonably straightforward setup matter. Managed-infrastructure complexity (autoscaling, multi-region, zero-downtime deploys at scale) does not.

## What it is not

- A streaming platform
- A social network (though light social features may come later)
- A replacement for Spotify — it *integrates* with Spotify as a data source
- A public internet service or SaaS product
