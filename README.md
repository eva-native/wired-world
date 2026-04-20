# wired-world

Simple guest book/board on Go, backed by Redis.

## Requirements

- Redis 7+

## Deployment

1) Docker Compose (recommended)
```
docker compose up -d
```

2) Docker container
```
docker build -t wired-world:latest .
docker run -d -p 8080:8080 wired-world:latest -redis=<host>:6379
```

3) Run directly
```
go run ./cmd/wired-world.go -redis=localhost:6379
```

## Program options

| Flag | Default | Description |
|---|---|---|
| `-addr` | `:8080` | HTTP server listen address |
| `-redis` | `localhost:6379` | Redis address `host:port` |
| `-behind-proxy` | `false` | Trust `X-Real-IP` / `X-Forwarded-For` headers for rate limiting |

## Rate limiting

`POST /post` is rate limited to **1 request per 5 seconds per IP** with a burst of 2. Exceeding the limit returns `HTTP 429` with a `Retry-After: 5` header.

When running behind a reverse proxy (e.g. nginx), set `-behind-proxy=true` so the real client IP is read from `X-Real-IP` / `X-Forwarded-For` headers. Only enable this when a trusted proxy sets these headers.

For nginx, ensure this is set:
```nginx
proxy_set_header X-Real-IP $remote_addr;
```

## In feature

- use Websockets instead of poll
