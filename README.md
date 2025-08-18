# wired-world

Simple guest book/board on go

## Deployment

1) Docker compose
  ```
  docker compose up -d
  ```

2) Docker container
```
docker build -t wired-world:latest .
docker run -ti -p 8080:8080 -v $(pwd)/db:/db wired-world:latest -db=file://db/w.db?cache=share&mode=rwc
```

## Program options

- -addr: provide listen address (:8080 by default)
- -db: provide dsn for sqlite (recomended use with query = cache=share&mode=rwc for concurency) (:memory: by default)

## In feature

- use Websockets instead poll
- use concurency storage
