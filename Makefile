BINARY    := liift
BIN_DIR   := ./bin
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS   := -s -w \
  -X liift/api/handlers.Version=$(VERSION) \
  -X liift/api/handlers.Commit=$(COMMIT) \
  -X 'liift/api/handlers.BuildTime=$(BUILD_TIME)'

.PHONY: build dev test lint clean docker-build docker-up docker-down

## Build the production binary (frontend first, then Go)
build:
	cd web && npm run build
	CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -buildvcs=false -o $(BIN_DIR)/$(BINARY) .

## Run backend (air) and frontend (vite) concurrently in dev mode
dev:
	@trap 'kill 0' EXIT; \
	cd web && npm run dev & \
	air

## Run Go tests
test:
	go test ./...

## Run golangci-lint (install: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

## Remove build artifacts
clean:
	rm -rf $(BIN_DIR) web/dist tmp/

## Build the Docker image
docker-build:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILD_TIME="$(BUILD_TIME)" \
	  -t $(BINARY):$(VERSION) \
	  -t $(BINARY):latest \
	  .

## Start with docker compose (SQLite, simplest)
docker-up:
	docker compose up --build

## Stop docker compose
docker-down:
	docker compose down
