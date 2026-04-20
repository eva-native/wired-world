FROM golang:1.25-alpine

WORKDIR /src
COPY . /src

RUN go build -tags embed -o /bin/wired-world ./cmd/wired-world.go

FROM alpine:3.21

COPY --from=0 /bin/wired-world /bin/
EXPOSE 8080/tcp

ENTRYPOINT ["/bin/wired-world"]
