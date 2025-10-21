FROM golang:latest AS build

RUN mkdir -p /go/src
WORKDIR /go/src
COPY go.mod Makefile ./
COPY internal/ internal/
COPY cmd/ cmd/

RUN apt -y update && \
    apt -y upgrade
RUN mkdir -p /bin

RUN go mod download && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build  -o bin/bitly-ingest ./cmd/bitly-ingest/main.go


FROM debian:bullseye-slim

WORKDIR /app
RUN mkdir bin
RUN apt -y update && \
    apt -y upgrade

COPY --from=build /go/src/bin/bitly-ingest bin/
COPY data data
ENTRYPOINT ["./bin/bitly-ingest"]





