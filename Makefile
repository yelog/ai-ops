.PHONY: all build clean test run install help release-build

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w \
	-X github.com/your-org/ai-k8s-ops/pkg/version.Version=$(VERSION) \
	-X github.com/your-org/ai-k8s-ops/pkg/version.BuildTime=$(BUILD_TIME) \
	-X github.com/your-org/ai-k8s-ops/pkg/version.GitCommit=$(GIT_COMMIT)

all: build

build:
	@echo "Building backend (version=$(VERSION) commit=$(GIT_COMMIT))..."
	go build -ldflags "$(LDFLAGS)" -o bin/server cmd/server/main.go
	go build -ldflags "$(LDFLAGS)" -o bin/agent cmd/agent/main.go
	go build -ldflags "$(LDFLAGS)" -o bin/ai-k8s-ops cmd/cli/main.go
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

# Cross-compile for a specific platform: make release-build GOOS=linux GOARCH=amd64
release-build:
	@echo "Cross-compiling for $(GOOS)/$(GOARCH)..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o bin/server cmd/server/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o bin/ai-k8s-ops cmd/cli/main.go

help:
	@echo "Available targets:"
	@echo "  build          - Build all binaries and frontend"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run all tests"
	@echo "  run            - Run development server"
	@echo "  install        - Install dependencies"
	@echo "  release-build  - Cross-compile for release (set GOOS and GOARCH)"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION    - Override version (default: git describe)"
	@echo "  GOOS       - Target OS for release-build"
	@echo "  GOARCH     - Target arch for release-build"
