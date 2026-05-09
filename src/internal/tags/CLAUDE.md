# tags — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- Tag normalization (lowercase, trim, strip non-letter/digit/`-&`) is a private helper in `service.go` — domain logic, not a utility.
