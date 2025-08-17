FROM golang:alpine

WORKDIR /src
COPY . /src
ENV CGO_ENABLED=1

RUN apk add --no-cache alpine-sdk &&\
    go build -tags embed -o /bin/wired-world ./cmd/wired-world.go

FROM alpine:latest

COPY --from=0 /bin/wired-world /bin/
EXPOSE 8080/tcp

ENTRYPOINT ["/bin/wired-world"]
