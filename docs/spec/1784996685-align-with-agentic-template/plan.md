# Align wax with the agentic-template

Migrate wax's docs, process, context files, and skills to match the patterns encoded in
`~/workshop/projects/agentic-template`. This is a **docs/process/structure** migration — no
application (Go/templ) code changes. two-cents remains the authority where projects diverge; the
template is the current distilled form of these patterns.

## Gap summary (wax today → template target)

| # | Area | wax today | template target |
|---|---|---|---|
| 1 | Context file naming | `CLAUDE.md` canonical everywhere; root at `.claude/claude.md`; `.claude/` is the real dir | `AGENTS.md` canonical + `CLAUDE.md` symlink; `.agents/` canonical + `.claude/` symlink; root context at repo-root `/AGENTS.md`; git hook regenerates links |
| 2 | Audit skills | 3 skills: `audit` + `code-audit` + `docs-audit`, parallel-subagent dispatch | 1 `audit/` skill: `SKILL.md` + `code-audit.md` + `docs-audit.md` reference files; **single** subagent runs both checklists |
| 3 | Process skills | none for spec/implement; relies on `/build`, `/grill-with-docs` | `spec`, `implement`, `prose-compact` skills |
| 4 | Doc layers / process | "repo holds decisions, not exploration"; working docs **never committed** | two-layer model: canonical docs (live) + committed `docs/spec/<timestamp>-<name>/` frozen at merge |
| 5 | Backlog | `docs/roadmap.md` + `docs/backlog.md` | `docs/backlog/` (`features.md` + `bugs.md`); roadmap removed |
| 6 | Product | `docs/vision.md` (single file) | `docs/product/` (README index + per-feature docs; present-state, what/why) |
| 7 | Domain | **none** | `docs/domain/` (README index: entities, glossary, cross-domain ledger + per-domain docs) |
| 8 | Doc practices | no prose/structure rule; no product-docs table convention | "Structure over prose; compress what's left" (→ `/prose-compact`); module `AGENTS.md` carries a **Product docs** table (what + when) |
| 9 | Module READMEs | 12/15 have README; `auth` + `genregraph` missing (required), `server` missing (OK — singleton) | every domain module + utility has README.md + AGENTS.md; singletons AGENTS.md-only |

## Ordering principle

Mechanical low-risk renames first (context files, skills), then the process/doc-model change, then
content-bearing restructures (backlog, product, domain), finally the per-module sweep. Each task
leaves the repo consistent and mergeable on its own.

## Suggested PR batches

| PR | Tasks | Character | Reviewable in |
|---|---|---|---|
| **A — structural** | 1, 2, 3, 4, 5, 8-rules | Pure renames + skill moves + process rewrite + backlog dir + the two doc-practice rules. No new prose content. | one sitting |
| **B — product** | 6 | vision move + `library`/`radar`/`ratings` docs; other 5 registered as stubs | one sitting |
| **C — domain** | 7 | index + `library`/`review`/`genres` docs; other 6 registered as stubs | one sitting |
| **D — sweep + gate** | 8-tables, 9 | per-module Product-docs tables, backfill `auth`/`genregraph` READMEs, `/audit`, reconcile, freeze | one sitting |

PR A is the whole structural migration and unblocks the rest; B and C are independent and can land in
either order (each is a self-contained `docs/spec/` chunk if split out).

---

## Task 1 — Adopt AGENTS.md / .agents canonical + symlink convention

**Why:** tool-agnostic canonical names; Claude-specific names become generated symlinks.

- [ ] Copy `.githooks/link-agents.sh` + `post-checkout`/`post-merge`/`post-rewrite` from the template into wax `.githooks/`.
- [ ] `git config core.hooksPath .githooks` (record in README/contributing notes).
- [ ] Rename every canonical `CLAUDE.md` → `AGENTS.md`:
  - root `.claude/claude.md` → repo-root `/AGENTS.md` (root context moves to repo root).
  - each `src/internal/*/CLAUDE.md`, `docs/architecture/CLAUDE.md`, `docs/design/CLAUDE.md`, `static/CLAUDE.md`, `e2e/CLAUDE.md`, `src/internal/core/templates/CLAUDE.md` → sibling `AGENTS.md`.
- [ ] Rename `.claude/` → `.agents/` (config + skills + memory + settings), then run `link-agents.sh` to create the `.claude` symlink and all `CLAUDE.md` symlinks.
- [ ] Update `.gitignore` to ignore generated `CLAUDE.md` + `.claude` symlinks (`*_templ.go`, `node_modules/`, etc. already ignored).
- [ ] Fix all intra-doc links that referenced `.claude/CLAUDE.md` or `CLAUDE.md` paths (root context, process.md, architecture/design READMEs).
- [ ] Add the convention note (tool-agnostic entry point; `.agents/`↔`.claude/`) to the new root `AGENTS.md`.

**Verify:** `link-agents.sh` runs clean; `.claude/CLAUDE.md` resolves; no broken links (grep for `claude.md`).

## Task 2 — Merge the three audit skills into one

- [ ] Replace `.agents/skills/audit/SKILL.md` with the template's single-subagent dispatcher.
- [ ] Create `.agents/skills/audit/code-audit.md` and `docs-audit.md` from the template checklists, **adapted to wax** (wax theme name, real module names left generic, wax's `docs/` layout).
- [ ] Delete `.agents/skills/code-audit/` and `.agents/skills/docs-audit/`.
- [ ] Grep for dangling references to the removed skill dirs (root `AGENTS.md`, process.md, PATTERNS-equivalent) and fix.

**Verify:** `grep -rn "code-audit\|docs-audit"` returns only hits inside `.agents/skills/audit/`.

## Task 3 — Add spec, implement, prose-compact skills

- [ ] Copy `spec/SKILL.md`, `implement/SKILL.md`, `prose-compact/SKILL.md` from the template into wax `.agents/skills/`.
- [ ] In `implement/SKILL.md`, keep wax's real gate commands (`go build ./...`, `go test ./src/...`, `task test/e2e`) — confirm they match wax's taskfile.
- [ ] Re-run `link-agents.sh` so the new skill dirs are reachable via `.claude`.

## Task 4 — Adopt the two-layer doc model + spec convention

- [ ] Rewrite `docs/process.md` to the template's two-layer model: canonical docs (live) + `docs/spec/<timestamp>-<name>/` (committed, reconciled, frozen at merge). Phases point at `/spec`, `/implement`, `/audit`.
- [ ] Add `docs/spec/README.md` (folder convention, timestamp naming, freeze lifecycle) from the template.
- [ ] Update root `AGENTS.md`: "Starting new work" (spec folder is phase 1), doc-map row for `docs/spec/<timestamp>-<name>/`, and replace the "Working artifacts (not committed)" section with "Working docs and durable outputs" (committed spec docs + durable-output canonical-home table).
- [ ] Remove the `/grill-with-docs` reference from process.md (or keep only if that skill still exists in wax).

**Verify:** this plan's own folder (`docs/spec/1784996685-align-with-agentic-template/`) is a valid instance of the new convention.

## Task 5 — Backlog dir

- [ ] Create `docs/backlog/features.md` and `docs/backlog/bugs.md`; migrate content from `docs/backlog.md` and the backlog portion of `docs/roadmap.md`.
- [ ] Delete `docs/backlog.md` and `docs/roadmap.md`.
- [ ] Update doc-map + any references (root `AGENTS.md`, process.md) from `roadmap.md`/`backlog.md` → `docs/backlog/`.

## Task 6 — Product dir

- [ ] Create `docs/product/README.md` (index + entry-point/reference framing + boundaries) from the template.
- [ ] **Keep `docs/vision.md`** — it is genuine product vision (what/who/philosophy/aesthetic/non-goals), not a feature list, and has no home in the template's per-feature model. Move it to `docs/product/vision.md` and link it from the product README as the top-of-index overview. (The template lacks a vision doc; wax's is worth preserving — flag it as a candidate pattern to fold back into the template.)
- [ ] Author per-feature docs (present-state, what/why, link to domain, no how). Concrete set for this pass, ordered by user-centrality:

  | Feature doc | Covers | Modules it spans |
  |---|---|---|
  | `library.md` | collect: own / wishlist / removed, releases & formats | library |
  | `radar.md` | watch albums; eligibility rules; Spotify-inbox entry | library + feed + spotify *(cross-cutting)* |
  | `ratings.md` | rating lifecycle (provisional → finalized) + per-rating notes | review |
  | `tags.md` | user-defined tag groups & vocabulary | tags |
  | `notes.md` | freeform album annotations | notes |
  | `genres.md` | curated, auto-assigned genre facet | genres + genregraph |
  | `listening-history.md` | play records over time | listeninghistory |
  | `discover.md` | radar-destination discover/search (ADR 0008) | library + spotify |

- [ ] Register every doc in the product README table; the `radar`/`discover`/`genres` rows are the worked examples of the orthogonal (cross-module) decomposition.
- [ ] `auth` and `user` are **infrastructure, not product features** — no product doc; they live only as domain + module docs.

**Scope (decided — core-subset-first):** this pass ships **`vision.md` + `library` + `radar` + `ratings`** (the album-centric core). The remaining five (`tags`, `notes`, `genres`, `listening-history`, `discover`) are created as **registered stubs** — a heading + one-line scope in the doc and a row in the README table — and filled in follow-up `docs/spec/` chunks. A partial-but-registered index is valid and keeps PR B reviewable in one sitting.

## Task 7 — Domain dir (new)

Net-new modeling — wax has no domain doc today, but `data-model.md` + module READMEs already hold
most of the raw material. Synthesize, don't invent.

- [ ] Create `docs/domain/README.md` (index) from the template. Seed the three cross-cutting sections:
  - **Entities** (shared nouns): Album, Release *(format)*, Rating, Tag, TagGroup, Note, Genre, RadarEntry, InboxItem, ListeningRecord, User, Session.
  - **Glossary** (confusables — pull straight from `data-model.md` + ADRs): owned / wishlisted / removed, radar eligibility, provisional / finalized rating, release-vs-format, tag-vs-genre, *primary* genre, Spotify inbox.
  - **Cross-domain write ledger**: ownership change clears the radar entry (library); Spotify-inbox sync creates radar candidates (feed → library); genre auto-assignment derives from genregraph (genres).
- [ ] Add per-domain docs (`docs/domain/<domain>.md`) — one per domain module, in dependency-leaf-first order:

  | Domain doc | Owning module | Key content to capture |
  |---|---|---|
  | `library.md` | library | ownership state machine (owned/wishlisted/removed), radar eligibility rules |
  | `review.md` | review | rating append-only log + provisional→finalized state machine (ADR 0003) |
  | `genres.md` | genres | curated taxonomy + derived auto-assignment (the "no global taxonomy" exception) |
  | `tags.md` | tags | user tag groups + assignments |
  | `notes.md` | notes | album annotations |
  | `feed.md` | feed | activity feed + Spotify-inbox sync ownership |
  | `listeninghistory.md` | listeninghistory | play records |
  | `user.md` | user | profile/settings |
  | `auth.md` | auth | session issuance + credential store (ADR-equivalent to two-cents 0007) |

- [ ] Wire "Synchronized content" in root `AGENTS.md`: data-model decisions in `docs/architecture/data-model.md`, per-entity meaning in module READMEs, term meaning canonical in `docs/domain/README.md` — the three-way sync rule.

**Scope (decided — core-subset-first):** this pass ships the **index + `library` / `review` / `genres`** domain docs (the three with real state machines / derivation rules). The other six (`tags`, `notes`, `feed`, `listeninghistory`, `user`, `auth`) are **registered stubs** in the domain index table, filled in follow-up `docs/spec/` chunks.

## Task 8 — Doc-practice rules + per-module Product docs tables

- [ ] Add to root `AGENTS.md` Documentation practices: "Structure over prose; compress what's left" (→ `/prose-compact`).
- [ ] Add the "Product docs table in module AGENTS.md" convention block to root `AGENTS.md`.
- [ ] Sweep each `src/internal/*/AGENTS.md`: add a `## Product docs` table (what + when) linking the module's relevant `docs/product/` docs; omit where no product doc exists (external clients, `auth`, `user`, singletons).
- [ ] Backfill missing READMEs — **only the two that the template requires:**
  - `auth/README.md` (domain module — required).
  - `genregraph/README.md` (utility — required).
  - `server/` has no README by design (singleton documents rules in its `AGENTS.md`, matching the template's `core`/`server`) — leave as-is.

## Task 9 — Audit + reconcile

- [ ] Run `/audit` (the new merged skill) over the whole repo; fix findings (expect broken-link and drift hits from the moves).
- [ ] Reconcile this spec folder to what actually shipped (final edit before freeze).
- [ ] Squash-merge; the spec folder freezes.

---

## Risks / notes

- **Symlink + git:** wax IS a git repo (unlike the workshop), so the hook-generated symlinks are committed/ignored per `.gitignore` — verify the hook path is set and symlinks resolve on a fresh clone.
- **Genericization does not apply:** the template scrubbed domain terms (widgets/someapi); wax keeps its real names — only the *patterns* transfer, not the placeholder content.
- **Tasks 6 & 7 are content-heavy** (product + domain). If either balloons, spin it into its own `docs/spec/<timestamp>-<name>/` and land it separately — PR A is the pure structural migration and merges first.
- **Cross-check against two-cents** before finalizing any doc rule wording — two-cents is the authority and may already phrase a rule better than the template.
- **Feed patterns back to the template.** Two wax assets have no template equivalent and are worth folding *upstream* after this lands: (1) `docs/product/vision.md` — a product-vision doc distinct from per-feature docs; (2) the `library`/`radar` cross-module examples, which are cleaner real-world demonstrations of the orthogonal domain-vs-product split than the template's synthetic `widgets`. The template is a pattern lab — this migration is also a chance to improve it.

## Reconciliation — what shipped

All tasks landed across 4 commits on `align-agentic-template`, tracked as PR #48:

| Task | Status | Notes |
|---|---|---|
| 1 — AGENTS.md/.agents symlink convention | Shipped | 21 renames, .githooks/ added, hooksPath set |
| 2 — Merge audit skills | Shipped | Single dispatcher + 2 checklists in .agents/skills/audit/ |
| 3 — Add spec, implement, prose-compact skills | Shipped | All 3 copied from template |
| 4 — Two-layer doc model | Shipped | process.md rewritten, docs/spec/README.md added |
| 5 — Backlog dir | Shipped | docs/backlog.md + roadmap.md → docs/backlog/features.md + bugs.md |
| 6 — Product dir | Shipped | vision.md moved, library/radar/ratings authored, 5 stubs created |
| 7 — Domain dir | Shipped | index + library/review/genres authored, 6 stubs created |
| 8-rules — Doc-practice rules | Shipped | Structure-over-prose + product-docs table convention in root AGENTS.md |
| 8-tables — Per-module Product docs | Shipped | 9 module AGENTS.md files updated with Product docs tables |
| 9 — Backfill auth/genregraph READMEs | Shipped | Both READMEs written |

### Divergences from plan

- PR A, B, C, D shipped on a single branch (`align-agentic-template`) in one stacked PR rather than separate PRs — the changes are small enough to review together.
- `docs/backlog.md` and `docs/roadmap.md` were deleted and content migrated; git detected `roadmap.md → backlog/features.md` as a rename.
- `docs/specs/` .gitignore entry removed (it was a stale artefact from before the `docs/spec/` convention; the committed spec folder is unaffected).

### Post-initial-commit fixes (pre-merge audit pass)

Audit surfaced 5 issues fixed before freeze:

1. **Cross-domain write ledger used fabricated Go operation names** (`SetAlbumOwnership`, `SyncRadarInbox`, `AddRadarEntry`, `SyncSavedAlbums`, `EnrichAlbumGenres` — none exist in the codebase). Ledger rephrased to conceptual operations in line with the domain-doc rule of staying implementation-agnostic (`docs/domain/README.md`, `docs/domain/library.md`).
2. **Root `AGENTS.md` convention note** cited `.claude/CLAUDE.md` as the Claude Code entry point — that path does not resolve. Corrected to the real mechanism: the sibling `CLAUDE.md → AGENTS.md` symlink at repo root.
3. **`docs/domain/library.md` ADR links** used `../../adr/` (one level too deep from `docs/domain/`); corrected to `../adr/`.
4. **`.agents/skills/implement/SKILL.md` and `.agents/skills/spec/SKILL.md`** referenced `../../docs/` (resolves to `.agents/docs/`, nonexistent); corrected to `../../../docs/`.
