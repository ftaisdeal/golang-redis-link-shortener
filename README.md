# Go + Redis Link Shortener

Minimal production-oriented starter for a URL shortener using Go and Redis.

## Features

- `POST /api/shorten` to create short links
- `GET /{code}` to redirect
- Optional custom aliases
- Optional TTL expiration
- Click counter per short code
- Redis-backed IP rate limiting
- URL validation with localhost/private-network blocking
- Health endpoint (`/health`)

## API

### Create short URL

`POST /api/shorten`

Request body:

```json
{
  "url": "https://example.com/very/long/path",
  "custom_code": "optionalAlias",
  "expires_in_seconds": 86400
}
```

Response:

```json
{
  "code": "optionalAlias",
  "short_url": "http://localhost:8080/optionalAlias",
  "expires_at": "2026-03-08T12:00:00Z"
}
```

### Redirect

`GET /{code}` -> returns `307 Temporary Redirect`

### Health

`GET /health`

## Run locally

1. Start Redis

```bash
docker run --name shortener-redis -p 6379:6379 -d redis:7-alpine
```

2. Configure environment

```bash
cp .env.example .env
```

3. Run app

```bash
go run ./cmd/server
```

## Run with Docker Compose

```bash
docker compose up --build
```

## Config

- `PORT` default `8080`
- `BASE_URL` default `http://localhost:8080`
- `REDIS_ADDR` default `localhost:6379`
- `REDIS_PASSWORD` default empty
- `REDIS_DB` default `0`
- `CODE_LENGTH` default `7`
- `RATE_LIMIT_PER_MINUTE` default `60`

## Test

```bash
go test ./...
```
