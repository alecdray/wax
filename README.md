# Shmoopicks

Listen with wax.

A music library management application with album ratings, reviews, and Spotify integration.

## Tech Stack

- **Language**: [Go](https://go.dev/) 1.25.5
- **Database**: [SQLite](https://www.sqlite.org/) with [SQLC](https://sqlc.dev/) for type-safe queries
- **Migrations**: [Goose](https://github.com/pressly/goose)
- **Templates**: [templ](https://templ.guide/) (Go templating language that compiles to Go)
- **Frontend**: [HTMX](https://htmx.org/) for dynamic interactions, [Tailwind CSS](https://tailwindcss.com/) with [DaisyUI](https://daisyui.com/) for styling
- **Authentication**: [JWT](https://jwt.io/) tokens
- **External APIs**: [Spotify API](https://developer.spotify.com/documentation/web-api), [MusicBrainz API](https://musicbrainz.org/doc/MusicBrainz_API)
- **Task Runner**: [Task](https://taskfile.dev/)

## Architecture

### Project Structure

```
src/internal/
├── core/           # Shared utilities (contextx, httpx, templates, db, task manager)
├── auth/           # Authentication and authorization
├── user/           # User management
├── library/        # Music library (albums, tracks)
├── review/         # Rating and review system
├── feed/           # Data feeds from external sources (e.g. Spotify library sync)
├── spotify/        # Spotify integration
├── musicbrainz/    # MusicBrainz metadata
└── server/         # HTTP server setup and routing
```

### Key Patterns

1. **Service Layer**: Each module has a `Service` struct that contains business logic
2. **Adapters**: HTTP handlers and template components live in `adapters/` subdirectories
3. **Context Enhancement**: Custom `contextx.ContextX` wraps `context.Context` for user authentication
4. **Type-Safe SQL**: SQLC generates Go code from SQL queries
5. **Template Generation**: `.templ` files compile to `_templ.go` files via `templ generate`

## Modules

Each module in `src/internal/` follows consistent patterns with service layers, adapters for HTTP/templates, and domain models. See individual module READMEs for detailed documentation.

## Development

### Getting Started

1. Copy `.env.template` to `.env` and configure required variables
2. Run database migrations: `task db/up`
3. Start development server: `task dev` or `task run`

### Task Runner

This project uses [Task](https://taskfile.dev/) for build automation. Run `task` without arguments to list all available commands, or see `taskfile.yml` for task definitions.

### Environment Variables

See `.env.template` for required configuration and detailed documentation of all environment variables.

## Deployment

### Docker

Build the image (also compiles Tailwind CSS):

```sh
task docker/build
```

Run the container:

```sh
task docker/run
```

Secrets are read from your `.env` file. Set `DOCKER_DATA_DIR` in `.env` to control where the SQLite database is stored on the host (defaults to `/data`).

### Docker Compose

A `docker-compose.example.yml` is provided as a starting point:

```sh
cp docker-compose.example.yml docker-compose.yml
docker compose up -d
```

Docker Compose reads secrets from your `.env` file automatically.

The SQLite database is persisted in a named Docker volume (`app-data`). To use a host directory instead, replace the `app-data` volume mount with a bind mount (e.g. `./data:/data`).

Generate secrets with:

```sh
openssl rand -hex 32  # JWT_SECRET
openssl rand -hex 32  # SPOTIFY_TOKEN_SECRET
```

## Roadmap

See [docs/roadmap.md](./docs/roadmap.md) for planned features and development progress.

## Design Philosophy

- Progressive enhancement, server-rendered HTML
- Deeper, more intimate relationship with music
