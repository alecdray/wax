# Code Audit Checklist

Audit project code for compliance with the rules in `docs/architecture/` and `docs/design/`. Read-only — report findings, do not edit.

## Scope

Default: every directory under `src/internal/`, every `.templ` file, and `static/src/main.css`.

When an argument is supplied:

- Module name (e.g. `library`) — limit to `src/internal/<module>/`.
- File path or directory — limit to that scope.
- `diff` — limit to files changed vs `main` (`git diff --name-only main...HEAD`), filtered to code paths.

## Steps

1. **Read the rule docs** in `docs/architecture/` (README, archetypes, known-gaps) and `docs/design/` (README, archetypes, principles, plus the convention files). Also read each singleton's `AGENTS.md` under `src/internal/`. These are the spec.
2. **Classify each module.** Read its `AGENTS.md` to determine archetype (or singleton). A missing or undeclared archetype is itself a violation.
3. **Audit each module against its archetype's rules** — file layout, imports, service contracts, persistence isolation (only `repo.go` touches generated query code), the peer-adapter rule, provider isolation (no external-client import outside the module that owns it). The archetype docs list the rules.
4. **Audit each `.templ` against its archetype** (determined by location + suffix) and against the cross-cutting design rules in `principles.md` (theme tokens, testids, OOB single-sourcing, HTMX-first, inline errors).
5. **Reconcile with known gaps.** Cross-check every potential violation against `docs/architecture/known-gaps.md`. Matches are tracked, not new. Also report any known-gap entry that no longer matches reality — it should be removed.

## Output sections

### Code — Archetype violations
For each finding: `path:line` *(archetype)* — what's wrong, which rule, recommended fix.

### Code — Design rule violations
For each finding: `path:line` — what's wrong, which rule, recommended fix.

### Code — Convention violations
For each finding: `path` — what's wrong, recommended fix.

### Code — Known gaps confirmed
For each: gap title — still present, as documented.

### Code — Known gaps no longer present
For each: gap title — no matching violation found; candidate for removal from `known-gaps.md`.