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

- `-addr`: listen address (`:8080` by default)
- `-redis`: Redis address `host:port` (`localhost:6379` by default)

## In feature

- use Websockets instead of poll
