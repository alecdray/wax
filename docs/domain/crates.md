# Crates — domain

A **crate** is a named, unordered collection of albums owned by a user.

## Key rules

- A crate belongs to one user; users cannot share crates.
- An album may belong to multiple crates.
- Crates carry no position metadata — membership is the point, not order.
- Crate names are not unique — two crates named "Jazz" are allowed, though discouraged by the UI.
- Deleting a crate removes all its memberships (cascade); the albums themselves are unaffected.

## Relationships

- `crates` — owned by `users`
- `crate_albums` — junction between `crates` and `albums`; album data is owned by the `library` module and read through `library.Service`
