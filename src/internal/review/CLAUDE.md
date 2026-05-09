# review — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Pure-logic types split across `rating.go` (rating values, scoring questions, labels) and `state.go` (rating-state machine — snoozing, rerate timing). Two topic files because they're distinct concepts with no shared types and no methods crossing them.
