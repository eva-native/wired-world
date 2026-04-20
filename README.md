# wired-world

Simple guest book/board on Go, backed by Redis.

![CI](https://github.com/eva-native/wired-world/actions/workflows/ci.yml/badge.svg)

## Requirements

- Redis 7+

## Deployment

**1) Docker Compose (recommended)**
```
docker compose up -d
```
The app listens on `http://localhost:8000`.

**2) Pre-built Docker image**
```
docker run -d -p 8080:8080 ghcr.io/eva-native/wired-world -redis=<host>:6379
```

**3) Build and run with Docker**
```
docker build -t wired-world:latest .
docker run -d -p 8080:8080 wired-world:latest -redis=<host>:6379
```

**4) Run directly**
```
go run ./cmd/wired-world.go -redis=localhost:6379
```

## Program options

| Flag | Default | Description |
|---|---|---|
| `-addr` | `:8080` | HTTP server listen address |
| `-redis` | `localhost:6379` | Redis address `host:port` |
| `-metrics-addr` | `:9090` | Internal address for `/metrics` endpoint |

## Observability

Prometheus metrics are exposed at `/metrics` on the metrics port (`:9090` by default). The metrics port is separate from the main app port and should not be publicly exposed.

Available metrics:

- `http_requests_total` — request count by method, path, and status code
- `http_request_duration_seconds` — request latency histogram
- `http_requests_in_flight` — current number of in-flight requests

To scrape with Prometheus, add to `prometheus.yml`:
```yaml
scrape_configs:
  - job_name: wired-world
    static_configs:
      - targets: ["localhost:9090"]
```

All application logs are output as JSON to stdout.

## Upcoming

- Use WebSockets instead of polling
