#!/bin/bash

set -e

REPO="IniZio/vendetta"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $OS in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            echo "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    BINARY_NAME="vendetta-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi

    echo "$BINARY_NAME"
}

# Get latest release from GitHub API
get_latest_release() {
    curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
install_binary() {
    BINARY_NAME=$1
    TAG=$2

    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$BINARY_NAME"

    echo "Downloading $BINARY_NAME from $DOWNLOAD_URL"

    if command -v curl >/dev/null 2>&1; then
        curl -L -o "/tmp/vendetta" "$DOWNLOAD_URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "/tmp/vendetta" "$DOWNLOAD_URL"
    else
        echo "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    chmod +x "/tmp/vendetta"

    mkdir -p "$INSTALL_DIR"
    mv "/tmp/vendetta" "$INSTALL_DIR/vendetta"

    echo "Vendetta installed successfully to $INSTALL_DIR/vendetta"
    echo "Run 'vendetta --help' to get started"
}

main() {
    echo "Installing vendetta..."

    BINARY_NAME=$(detect_platform)
    TAG=$(get_latest_release)

    if [ -z "$TAG" ]; then
        echo "Failed to get latest release"
        exit 1
    fi

    echo "Latest release: $TAG"
    echo "Platform: $BINARY_NAME"

    install_binary "$BINARY_NAME" "$TAG"
}

main "$@"
