# Architecture docs

These docs describe the **conceptual** architecture: archetypes, foundational singletons, the rules each archetype follows. They are loaded by agents working anywhere in the project, so they must stay durable.

## Rule: no exhaustive lists

Do **not** enumerate every existing module, every external client, every utility, every `core/*` sub-package, or any other set whose membership changes when code is added or renamed. Lists like that rot the same way the wiki did — they go out of date silently as soon as the codebase moves.

### The test

Before listing concrete names, ask: *if a new instance is added next month, will this list be wrong?* If yes, it's an exhaustive list — drop it.

### What's OK

- **Single illustrative examples** that ground a rule (e.g. *"a topic file named after the package, like `library/library.go`"*). One concrete anchor per rule, not a roster.
- **Counts that match the architecture itself** (e.g. *"three archetypes"*, *"two singletons"*) — these only change when the architecture changes, in which case this doc is the obvious edit.
- **Conceptual descriptions** of what an archetype owns, allows, or forbids — phrased without naming current instances.

### What's not OK

- *"Domain modules (`library`, `review`, `tags`, `notes`, ...)"* — a list of all current instances.
- *"Allowed: `core/contextx`, `core/httpx`, `core/db`, `core/task`, ..."* — a list of all current sub-packages.
- *"Vendor SDKs like `github.com/zmb3/spotify/v2`, `github.com/lithammer/fuzzysearch/fuzzy`"* — a list of currently-used libraries.
- *"Canonical examples: `feed/task.go`, `listeninghistory/task.go`"* — a partial enumeration that pretends to be a single example.

For these, write conceptually (*"any domain module"*, *"any `core/*` sub-package"*, *"vendor SDKs for the wrapped API"*, *"existing `task.go` files in `src/internal/`"*) and let the live codebase be the source of truth for the membership. The README's `grep -h "^# " src/internal/*/CLAUDE.md` one-liner gives the at-a-glance view.

## Rule: conceptual over implementation

Architecture docs describe the **why** and the **what**: what an archetype is, what it owns, what its boundaries are. Specific file names, function signatures, helper functions, build commands, and import paths are implementation details — concrete enough that they belong in per-package `CLAUDE.md` files or in code, not as the load-bearing content of an archetype doc.

That said, a small amount of concrete grounding (single example file names, sample function signatures) is fine where it makes a rule unambiguous. The line is: don't make agents reach for these docs to look up *names*, make them reach to understand *intent*.

## Rule: framing notes that name current state will rot

The italicized framing line at the top of each archetype doc says *"existing modules may not yet conform — see `docs/architecture/refactor-backlog.md`"*. Do not replace that with anything specific to current divergent modules (e.g. *"`spotify` diverges because of X"*) — that becomes false the moment that module is fixed. The backlog is the right place for current-state details; the archetype doc stays neutral.
