# Architecture Decision Records

Short entries that capture **why** a decision was made, when the rationale would otherwise be lost once the old approach is gone.

## Decision log

| # | Decision | Summary |
|---|---|---|
| [0001](0001-library-visual-list.md) | Library shifts from table view to visual list | The library dashboard is a cover-art-first visual list with chip-bar filtering, replacing a sortable table; dashboard Spotify outlinks are dropped to keep navigation rooted in Wax. |
| [0002](0002-loading-feedback-for-network-actions.md) | Loading feedback for network actions | User-triggered network actions gain layered feedback — an app-wide indeterminate progress bar on every request, a busy/non-resubmittable state on discrete actions, and dim-and-overlay on in-place data reloads (a trailing spinner for append-style loads). |
| [0003](0003-rating-lifecycle-determined-by-action.md) | Rating lifecycle state is determined by the save action | Saving always lands provisional, finalizing always lands finalized, both from any prior state; saving a finalized album demotes it — the only un-finalize path. |

## Format

Lead with the decision — 1–2 sentences naming what was decided — under an `# h1` title. Then add only the **minimal context needed to understand that decision and its implications**: the constraint or trade-off that forced it. Rejected alternatives and consequences earn space only when they carry real weight, and as a brief clause or short list — never an obligatory section. An ADR that runs past a few short paragraphs is usually restating things that belong elsewhere.

The **current** state is the codebase — don't restate it. Keep implementation details out: no file names, class names, function names, or exact UI strings that a routine refactor would invalidate. If a sentence would need to change after such a refactor, it doesn't belong here.

## Naming

`NNNN-short-slug.md` — four-digit zero-padded prefix, never renumbered. A decision that replaces an earlier one is a new ADR; reference the prior number in the body. Don't edit the old one.

## When to write one

Write an ADR when a change replaces a meaningful prior approach, or locks in a foundational choice, that a future reader would otherwise wonder *"why is it like this?"* about. Routine changes don't qualify.
