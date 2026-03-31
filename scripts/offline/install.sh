#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PKG_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== AI-K8S-OPS Offline Installer ==="
echo "Package directory: $PKG_DIR"

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo "ERROR: Cannot detect OS"
    exit 1
fi
echo "Detected OS: $OS"

# Preflight checks
echo ""
echo "=== Preflight Checks ==="

# Check disk space (need at least 5GB free)
FREE_SPACE=$(df -BG "$PKG_DIR" | tail -1 | awk '{print $4}' | tr -d 'G')
if [ "$FREE_SPACE" -lt 5 ]; then
    echo "ERROR: Need at least 5GB free disk space, only ${FREE_SPACE}GB available"
    exit 1
fi
echo "Disk space: ${FREE_SPACE}GB available"

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "ERROR: Must run as root"
    exit 1
fi
echo "Running as root: OK"

# Step 1: Install system dependencies
echo ""
echo "=== Step 1: Installing system dependencies ==="
bash "$SCRIPT_DIR/setup-deps.sh" "$PKG_DIR" "$OS"

# Step 2: Load container images
echo ""
echo "=== Step 2: Loading container images ==="
bash "$SCRIPT_DIR/load-images.sh" "$PKG_DIR"

# Step 3: Install binaries
echo ""
echo "=== Step 3: Installing K8S binaries ==="
if [ -d "$PKG_DIR/core/binaries" ]; then
    cp "$PKG_DIR/core/binaries/kubeadm" /usr/local/bin/
    cp "$PKG_DIR/core/binaries/kubelet" /usr/local/bin/
    cp "$PKG_DIR/core/binaries/kubectl" /usr/local/bin/
    chmod +x /usr/local/bin/{kubeadm,kubelet,kubectl}
    echo "Binaries installed to /usr/local/bin/"

    kubeadm version --short 2>/dev/null && echo "kubeadm: OK"
    kubectl version --client --short 2>/dev/null && echo "kubectl: OK"
fi

echo ""
echo "=== Offline installation complete ==="
echo ""
echo "Next steps:"
echo "  1. Initialize cluster:  kubeadm init --kubernetes-version=v1.28.0"
echo "  2. Set up kubeconfig:   mkdir -p ~/.kube && cp /etc/kubernetes/admin.conf ~/.kube/config"
echo "  3. Install network:     kubectl apply -f calico.yaml"
echo "  4. Join worker nodes:   kubeadm join ..."
