---
name: docs-audit
description: Pre-merge audit of project documentation for rot-prone content, drift against the codebase, and structural violations. Read-only — reports findings; does not edit. Keywords: docs audit, audit docs, doc rot, doc drift, pre-merge check, documentation audit.
context: fork
agent: Explore
argument-hint: "[optional: file path, directory, or 'diff' to scope to changed files]"
---

Audit project documentation for compliance with the rules in `.claude/claude.md`, `docs/architecture/CLAUDE.md`, and `docs/design/CLAUDE.md`. Read-only — report findings, do not edit.

## Scope

Default: the root `README.md`, `.claude/claude.md`, everything under `docs/` (excluding gitignored paths), and every `README.md` / `CLAUDE.md` under `src/internal/<module>/`.

When an argument is supplied:

- File path or directory — limit to that scope.
- `diff` — limit to files changed vs `main` (`git diff --name-only main...HEAD`), filtered to docs.

## Steps

1. **Read the rule docs.** Those listed above are the spec. They define what counts as rot, what counts as drift, what belongs where, and the allowed exceptions. Re-read the relevant rule before flagging anything ambiguous.
2. **Establish sources of truth.** The rule docs tell you what's canonical for each kind of factual claim (schema for entities, code for constants, `src/internal/` listing for modules, etc.). Gather what you need.
3. **Audit each in-scope doc.** Flag content that violates the rules, contradicts the code, or sits in the wrong location.
4. **Check duplication.** Any content that exists in two places must be registered in the **Synchronized content** section of `.claude/claude.md`. Unregistered duplication is a violation — either pick one location or register it.
5. **Report.** Group findings, sort by path, include the rule violated and a recommended action.

## Output

---

## Docs Audit Summary

### Rot-prone content
For each finding: `path:line` — what's wrong, which rule, recommended cut/soften.

### Drift against the codebase
For each finding: `path:line` — what the doc says, what the code does, how to reconcile.

### Structural violations
For each finding: `path` — what's wrong, which rule, recommended fix.

### Clean
Documents that passed without findings.

### Judgement calls
Anything ambiguous where you chose not to flag, or where the right call depends on intent the audit can't infer.

---
