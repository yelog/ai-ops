#!/bin/bash
set -e

PKG_DIR="${1:-.}"
OS="${2:-ubuntu}"

echo "Setting up system dependencies for $OS..."

case "$OS" in
    ubuntu|debian)
        if [ -d "$PKG_DIR/core/packages/ubuntu" ]; then
            echo "Installing Ubuntu/Debian packages..."
            dpkg -i "$PKG_DIR/core/packages/ubuntu/"*.deb 2>/dev/null || true
            echo "Ubuntu packages installed."
        else
            echo "No Ubuntu packages found in $PKG_DIR/core/packages/ubuntu"
        fi
        ;;
    centos|rhel|rocky|alma)
        if [ -d "$PKG_DIR/core/packages/centos" ]; then
            echo "Installing CentOS/RHEL packages..."
            rpm -ivh "$PKG_DIR/core/packages/centos/"*.rpm 2>/dev/null || true
            echo "CentOS packages installed."
        else
            echo "No CentOS packages found in $PKG_DIR/core/packages/centos"
        fi
        ;;
    *)
        echo "WARNING: Unsupported OS: $OS"
        echo "Supported: ubuntu, debian, centos, rhel, rocky, alma"
        ;;
esac

# Disable swap
echo "Disabling swap..."
swapoff -a || true
sed -i '/swap/d' /etc/fstab || true

# Load kernel modules
echo "Loading kernel modules..."
modprobe br_netfilter || true
modprobe overlay || true

cat > /etc/modules-load.d/k8s.conf <<EOF
br_netfilter
overlay
EOF

# Set sysctl parameters
echo "Configuring sysctl..."
cat > /etc/sysctl.d/k8s.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sysctl --system > /dev/null 2>&1

echo "System dependencies setup complete."
