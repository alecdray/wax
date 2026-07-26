# Docs Audit Checklist

Audit project documentation for rot-prone content, drift against the codebase, and structural violations. Read-only — report findings, do not edit.

## Scope

Default: the root `README.md`, root `AGENTS.md`, everything under `docs/` (excluding gitignored paths), and every `README.md` / `AGENTS.md` under `src/internal/<module>/`.

When an argument is supplied:

- File path or directory — limit to that scope.
- `diff` — limit to files changed vs `main` (`git diff --name-only main...HEAD`), filtered to docs.

## Steps

1. **Read the rule docs.** The root `AGENTS.md`, `docs/architecture/AGENTS.md`, and `docs/design/AGENTS.md` are the spec — especially the **Documentation practices**, **Synchronized content**, and **Working docs and durable outputs** sections of the root `AGENTS.md`, plus [`docs/spec/README.md`](../../../docs/spec/README.md) for the spec-folder convention. Re-read the relevant rule before flagging anything ambiguous.
2. **Establish sources of truth.** The rule docs name what's canonical for each kind of claim: domain language → `docs/domain/README.md`, decisions → `docs/adr/`, schema → `db/migrations/`, module membership → the `src/internal/` listing. Gather what you need.
3. **Audit each in-scope doc** for rot, drift, and misplacement. An `AGENTS.md` asserting current state must not carry historical, forward-looking, or comparative content.
4. **Check duplication against the registry.** Any fact stated in 2+ docs must be registered under **Synchronized content** in the root `AGENTS.md`. Unregistered duplication is a violation — recommend cutting from one location and linking, or registering it.
5. **Check spec docs and durable outputs.** Working docs for the change belong under `docs/spec/<timestamp>-<name>/`, committed and — for the current change — reconciled to what shipped; flag a spec folder that still describes pre-implementation intent. Flag genuinely throwaway scratch committed *outside* `docs/spec/`. Flag any durable output stranded in a spec folder that should be codified into its canonical home. Do not flag a *frozen* spec folder (one whose work already merged) for drifting from current code — by design it is a point-in-time record.

## Output sections

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