.PHONY: build run test-local test-docker

# Build locally (requires local opencv)
build:
	go build -tags gocv -o bin/server ./cmd/server

# Run locally
run:
	go run -tags gocv ./cmd/server/main.go

# Test locally (Mock mode by default, use TAGS=gocv to enable)
test-local:
	go test -v ./...

# Test using Docker (Recommended if local setup is broken)
test-docker:
	docker compose run --rm app go test -tags gocv -v ./...

# Run app using Docker
run-docker:
	docker compose up --build
