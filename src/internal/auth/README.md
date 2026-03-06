# Auth Module

## Scope

This module owns the HTTP endpoints for user authentication workflows: displaying the login page, handling OAuth callbacks, and logout. It coordinates the OAuth flow by integrating with the Spotify auth service, user service, and feed service.

This is an adapter layer - it contains no authentication business logic, session management, or OAuth primitives. Those belong in service and core modules.
