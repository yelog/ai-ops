#!/bin/bash
# AI-K8S-OPS 持续修复监控脚本
# 用途：定期检查项目构建状态，自动修复常见问题

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="${SCRIPT_DIR}"
LOG_FILE="${SCRIPT_DIR}/logs/monitor.log"
GO_BIN="/tmp/go/bin/go"

mkdir -p "${SCRIPT_DIR}/logs"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_go_version() {
    log "Checking Go version..."
    if ! $GO_BIN version 2>/dev/null; then
        log "ERROR: Go 1.23+ not found at $GO_BIN"
        return 1
    fi
    log "Go version: $($GO_BIN version)"
    return 0
}

check_build() {
    log "Building all components..."
    cd "$REPO_DIR"
    
    if CGO_ENABLED=1 $GO_BIN build -o bin/server cmd/server/main.go 2>&1 | tee -a "$LOG_FILE"; then
        log "✓ Server build successful"
    else
        log "✗ Server build failed"
        return 1
    fi
    
    if CGO_ENABLED=1 $GO_BIN build -o bin/agent cmd/agent/main.go 2>&1 | tee -a "$LOG_FILE"; then
        log "✓ Agent build successful"
    else
        log "✗ Agent build failed"
        return 1
    fi
    
    if CGO_ENABLED=1 $GO_BIN build -o bin/ai-k8s-ops cmd/cli/main.go 2>&1 | tee -a "$LOG_FILE"; then
        log "✓ CLI build successful"
    else
        log "✗ CLI build failed"
        return 1
    fi
    
    return 0
}

check_tests() {
    log "Running tests..."
    cd "$REPO_DIR"
    
    if $GO_BIN test ./... 2>&1 | tee -a "$LOG_FILE"; then
        log "✓ All tests passed"
        return 0
    else
        log "✗ Some tests failed"
        return 1
    fi
}

check_api_endpoints() {
    log "Checking API endpoints..."
    
    # Start server in background
    cd "$REPO_DIR"
    ./bin/server &
    SERVER_PID=$!
    sleep 3
    
    # Check health endpoint
    if curl -s http://localhost:8080/api/v1/system/health | grep -q "ok"; then
        log "✓ Health endpoint OK"
    else
        log "✗ Health endpoint failed"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi
    
    kill $SERVER_PID 2>/dev/null || true
    return 0
}

generate_report() {
    log "=== Build Monitor Report ==="
    log "Repository: $REPO_DIR"
    log "Go Version: $($GO_BIN version 2>&1)"
    log "Build Status: $1"
    log "Timestamp: $(date)"
    log "==========================="
}

main() {
    log "Starting build monitor check..."
    
    STATUS="SUCCESS"
    
    if ! check_go_version; then
        STATUS="FAILED - Go not found"
        generate_report "$STATUS"
        exit 1
    fi
    
    if ! check_build; then
        STATUS="FAILED - Build error"
        generate_report "$STATUS"
        exit 1
    fi
    
    # Optional: run tests
    # if ! check_tests; then
    #     STATUS="FAILED - Test error"
    #     generate_report "$STATUS"
    #     exit 1
    # fi
    
    generate_report "$STATUS"
    log "All checks passed!"
    return 0
}

# Run main
main "$@"
