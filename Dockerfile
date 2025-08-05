FROM golang:latest
WORKDIR /src
COPY . /src
RUN go build -o /bin/wired-world ./cmd/wired-world.go

FROM ubuntu:latest
COPY --from=0 /bin/wired-world /bin/
EXPOSE 8080/tcp
CMD ["/bin/wired-world"]
