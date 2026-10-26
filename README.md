# DealStalker

Self-hosted price tracker that automatically checks product prices on comparison sites and notifies you as soon as a price reaches your target.

> **Status:** Work in progress – not ready for use yet.

## Planned features

- Track products from price comparison sites (starting with Geizhals)
- Set a target price per product and user
- Each product is fetched only once per interval, no matter how many users watch it
- Notifications via ntfy and Discord
- Multi-user support with a web interface
- Easy self-hosting with Docker Compose

## Tech stack

- **Backend:** Go
- **Frontend:** Next.js
- **Deployment:** Docker / Docker Compose

## Documentation

See [Domain model](docs/domain-model.md).

## Development

Requirements: Go and Docker.

Start the local database:

```bash
docker compose up -d
```

Apply the database migrations:

```bash
go tool goose -dir migrations postgres "postgres://dealstalker:dealstalker@localhost:5432/dealstalker?sslmode=disable" up
```

## License

See [LICENSE](LICENSE).