#!/bin/bash
set -e

PKG_DIR="${1:-.}"

echo "Loading container images..."

LOADED=0
FAILED=0

find "$PKG_DIR" -name "*.tar" -path "*/images/*" | sort | while read tar_file; do
    name=$(basename "$tar_file")
    echo -n "  Loading: $name ... "

    if ctr -n k8s.io images import "$tar_file" 2>/dev/null; then
        echo "OK (containerd)"
        LOADED=$((LOADED+1))
    elif nerdctl load -i "$tar_file" 2>/dev/null; then
        echo "OK (nerdctl)"
        LOADED=$((LOADED+1))
    elif docker load -i "$tar_file" 2>/dev/null; then
        echo "OK (docker)"
        LOADED=$((LOADED+1))
    else
        echo "FAILED"
        FAILED=$((FAILED+1))
    fi
done

echo ""
echo "Image loading complete."
