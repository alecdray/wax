# Wax

Listen with wax.

A music library management application with album ratings, reviews, and Spotify integration.

For product, architecture, and roadmap, browse [`docs/`](./docs/).

## Tech Stack

- **Language**: [Go](https://go.dev/) 1.25.5
- **Database**: [SQLite](https://www.sqlite.org/) with [SQLC](https://sqlc.dev/) for type-safe queries
- **Migrations**: [Goose](https://github.com/pressly/goose)
- **Templates**: [templ](https://templ.guide/) (Go templating language that compiles to Go)
- **Frontend**: [HTMX](https://htmx.org/) for dynamic interactions, [Tailwind CSS](https://tailwindcss.com/) with [DaisyUI](https://daisyui.com/) for styling
- **Authentication**: [JWT](https://jwt.io/) tokens
- **External APIs**: [Spotify API](https://developer.spotify.com/documentation/web-api), [MusicBrainz API](https://musicbrainz.org/doc/MusicBrainz_API)
- **Task Runner**: [Task](https://taskfile.dev/)

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
