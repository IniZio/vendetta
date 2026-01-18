#!/bin/bash
set -euo pipefail

VERSION="${1:-latest}"
REPO_URL="https://github.com/IniZio/nexus"
RELEASES_URL="${REPO_URL}/releases"
API_URL="https://api.github.com/repos/IniZio/nexus"

NEXUS_BIN_DIR="${HOME}/.local/bin"
NEXUS_CONFIG_DIR="${HOME}/.nexus"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

error() { echo -e "${RED}❌ $*${NC}" >&2; exit 1; }
success() { echo -e "${GREEN}✅ $*${NC}"; }
warn() { echo -e "${YELLOW}⚠️  $*${NC}"; }
info() { echo -e "ℹ️  $*"; }

detect_os() {
	case "$(uname -s)" in
		Darwin) echo "darwin" ;;
		Linux) echo "linux" ;;
		*) error "Unsupported OS: $(uname -s)" ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64) echo "amd64" ;;
		aarch64) echo "arm64" ;;
		arm64) echo "arm64" ;;
		*) error "Unsupported architecture: $(uname -m)" ;;
	esac
}

check_command() {
	command -v "$1" > /dev/null 2>&1 || return 1
}

check_dependencies() {
	local missing=()
	
	for cmd in git ssh curl; do
		if ! check_command "$cmd"; then
			missing+=("$cmd")
		fi
	done
	
	if ! check_command gh; then
		warn "GitHub CLI not installed. Will prompt to install later."
	fi
	
	if [[ ${#missing[@]} -gt 0 ]]; then
		error "Missing required commands: ${missing[*]}"
	fi
	
	success "All dependencies found"
}

download_binary() {
	local os="$1" arch="$2" version="$3"
	local filename="nexus-${os}-${arch}"
	local download_url="${RELEASES_URL}/download/${version}/${filename}"
	local temp_file="/tmp/${filename}"
	
	info "Downloading nexus from ${download_url}..." >&2
	
	if ! curl -L --progress-bar -o "$temp_file" "$download_url"; then
		error "Failed to download nexus binary"
	fi
	
	chmod +x "$temp_file"
	echo "$temp_file"
}

setup_config_dir() {
	mkdir -p "$NEXUS_CONFIG_DIR"
	
	if [[ ! -f "$NEXUS_CONFIG_DIR/config.yaml" ]]; then
		cat > "$NEXUS_CONFIG_DIR/config.yaml" << EOF
github:
  username: ""
  user_id: 0
  avatar_url: ""

ssh:
  key_path: ~/.ssh/id_ed25519
  public_key: ""

editor: cursor

server:
  host: localhost
  port: 3001

workspaces: []
EOF
		success "Created ~/.nexus/config.yaml"
	else
		info "~/.nexus/config.yaml already exists"
	fi
}

setup_ssh_dir() {
	mkdir -p ~/.ssh
	chmod 700 ~/.ssh
	success "SSH directory ready (~/. ssh)"
}

install_binary() {
	local temp_file="$1"
	
	mkdir -p "$NEXUS_BIN_DIR"
	mv "$temp_file" "$NEXUS_BIN_DIR/nexus"
	chmod +x "$NEXUS_BIN_DIR/nexus"
	
	if [[ ":$PATH:" != *":$NEXUS_BIN_DIR:"* ]]; then
		warn "~/.local/bin not in PATH"
		info "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc):"
		echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
	fi
	
	success "Installed nexus to $NEXUS_BIN_DIR/nexus"
}

main() {
	echo "=== Nexus Installer ==="
	echo ""
	
	local os version arch
	os=$(detect_os)
	arch=$(detect_arch)
	
	info "Detected: ${os} (${arch})"
	
	check_dependencies
	
	version="${VERSION}"
	if [[ "$version" == "latest" ]]; then
		info "Fetching latest release..."
		# Use POSIX-compatible grep (works on both macOS/BSD and Linux/GNU)
		version=$(curl -s "${API_URL}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/' || echo "latest")
	fi
	
	info "Installing version: $version"
	
	local temp_file
	temp_file=$(download_binary "$os" "$arch" "$version")
	
	setup_ssh_dir
	setup_config_dir
	install_binary "$temp_file"
	
	echo ""
	success "Nexus installed successfully!"
	echo ""
	echo "Next steps:"
	echo "  1. Add to PATH: export PATH=\"\$HOME/.local/bin:\$PATH\""
	echo "  2. Setup GitHub: nexus auth github"
	echo "  3. Setup SSH: nexus ssh setup"
	echo "  4. Create workspace: nexus workspace create owner/repo"
	echo ""
}

main "$@"
