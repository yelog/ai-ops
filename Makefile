.PHONY: all build clean test run install help

all: build

build:
	@echo "Building backend..."
	go build -o bin/server cmd/server/main.go
	go build -o bin/agent cmd/agent/main.go
	@echo "Building frontend..."
	cd frontend && npm run build

clean:
	rm -rf bin/
	rm -rf data/*.db
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/

test:
	go test -v -race -coverprofile=coverage.out ./...
	cd frontend && npm test

run:
	go run cmd/server/main.go

install:
	go mod download
	cd frontend && npm install

help:
	@echo "Available targets:"
	@echo "  build    - Build all binaries"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run all tests"
	@echo "  run      - Run development server"
	@echo "  install  - Install dependencies"