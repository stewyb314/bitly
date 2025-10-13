test:
	go test -race -v ./...
build:
	go build -o bin/bitly-ingest ./cmd/bitly-ingest/main.go
run: build
	bin/bitly-ingest