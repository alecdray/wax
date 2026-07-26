---
name: app-audit
description: Pre-merge audit gate — checks both code and docs in a single pass. ALWAYS prompt the user to run this before merging, pushing a PR, or marking a branch as done. Keywords: audit, pre-merge, before merge, before push, before merging, ready to merge, ready to ship, final check, merge check, pre-PR, code audit, docs audit.
argument-hint: "[optional: 'diff' to scope to changed files vs main, or a path]"
---

Run the project's full pre-merge audit — code and docs in one pass — by dispatching a **single subagent**.

## Step

Dispatch one `Agent` (subagent_type `Explore`) with this prompt:

> You are running the pre-merge audit for wax. Read `.agents/skills/app-audit/code-audit.md` and `.agents/skills/app-audit/docs-audit.md` — those files contain the full checklists for code and docs respectively. Run both checklists in a single pass against this repo. Argument (scope): `<arg-or-none>`. Return only the Pre-Merge Audit block defined in `.agents/skills/app-audit/SKILL.md`'s Output section.

Wait for the subagent to return, then relay its output verbatim.

## Output

---

## Pre-Merge Audit

**Verdict:** <one line — ✅ clean, or ⚠️ N findings, worst category named>

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

### Docs — Rot-prone content
For each finding: `path:line` — what's wrong, which rule, recommended cut/soften.

### Docs — Drift against the codebase
For each finding: `path:line` — what the doc says, what the code does, how to reconcile.

### Docs — Unregistered duplication
For each finding: the fact, the 2+ `path:line` locations, recommended single home + link (or register).

### Docs — Spec-doc / durable-output issues
For each finding: `path` — an unreconciled spec folder, throwaway scratch committed outside `docs/spec/`, or a durable output stranded in a spec folder + its canonical home.

### Docs — Structural violations
For each finding: `path` — what's wrong, which rule, recommended fix.

### Clean
Modules, templs, and docs that passed without findings.

### Judgement calls
Anything ambiguous where you chose not to flag, or where the right call depends on intent the audit can't infer.

---