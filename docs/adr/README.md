# Architecture Decision Records

Short entries that capture **why** a decision was made, when the rationale would otherwise be lost once the old approach is gone.

## Format

```
# NNNN — Short title

**Date:** YYYY-MM-DD

**Was:** What it used to be. One or two sentences.

**Why:** The reason for the change. One or two sentences.
```

The **current** state is the codebase — don't restate it. No implementation details: no file names, class names, function names, or exact UI strings. If a sentence would need to change after a routine refactor, it doesn't belong here.

## Naming

`NNNN-short-slug.md` — four-digit zero-padded prefix, never renumbered. A decision that replaces an earlier one is a new ADR; reference the prior number in the body. Don't edit the old one.

## When to write one

Write an ADR when a change replaces a meaningful prior approach and a future reader would otherwise wonder *"why is it like this?"*. Routine changes don't qualify.
