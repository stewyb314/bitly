test:
	go test -race -v ./...
build:
	go mod download && \
	go mod tidy && \
	go build -o bin/bitly-ingest ./cmd/bitly-ingest/main.go
run: build
	bin/bitly-ingest

docker-run: image
	docker stop bitly-ingest || true && \
	docker rm bitly-ingest || true && \
	docker run -d  --name bitly-ingest stewyb/bitly-ingest:latest 
image:
	docker build -t stewyb/bitly-ingest:latest  . && \
	docker push stewyb/bitly-ingest:latest
