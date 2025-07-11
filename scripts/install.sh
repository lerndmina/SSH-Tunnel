#!/bin/bash
# SSH Tunnel Manager Installation Script
# Supports Linux, macOS, and Windows (via WSL/MSYS2)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/yourusername/ssh-tunnel-manager"
BINARY_NAME="ssh-tunnel"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.ssh-tunnel-manager"

# Detect platform
detect_platform() {
    local os arch
    
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    
    case $arch in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        armv7l|armv6l) arch="arm" ;;
        i386|i686) arch="386" ;;
        *) 
            echo -e "${RED}Unsupported architecture: $arch${NC}"
            exit 1
            ;;
    esac
    
    case $os in
        linux|darwin) ;;
        mingw*|msys*|cygwin*) os="windows" ;;
        *)
            echo -e "${RED}Unsupported operating system: $os${NC}"
            exit 1
            ;;
    esac
    
    echo "$os-$arch"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Download file
download_file() {
    local url=$1
    local output=$2
    
    if command_exists curl; then
        curl -fsSL "$url" -o "$output"
    elif command_exists wget; then
        wget -q "$url" -O "$output"
    else
        echo -e "${RED}Error: curl or wget is required to download files${NC}"
        exit 1
    fi
}

# Get latest release version
get_latest_version() {
    local api_url="https://api.github.com/repos/yourusername/ssh-tunnel-manager/releases/latest"
    
    if command_exists curl; then
        curl -fsSL "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command_exists wget; then
        wget -qO- "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        echo "v1.0.0"  # fallback version
    fi
}

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}Checking prerequisites...${NC}"
    
    # Check for required tools
    if ! command_exists curl && ! command_exists wget; then
        echo -e "${RED}Error: curl or wget is required${NC}"
        exit 1
    fi
    
    if ! command_exists tar; then
        echo -e "${RED}Error: tar is required${NC}"
        exit 1
    fi
    
    # Check for SSH client
    if ! command_exists ssh; then
        echo -e "${YELLOW}Warning: OpenSSH client not found. Please install it for SSH tunnel functionality.${NC}"
    fi
    
    echo -e "${GREEN}✓ Prerequisites check passed${NC}"
}

# Install binary
install_binary() {
    local platform=$1
    local version=$2
    local temp_dir
    
    echo -e "${BLUE}Installing SSH Tunnel Manager $version for $platform...${NC}"
    
    # Create temporary directory
    temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # Determine download URL and filename
    local filename="${BINARY_NAME}-${platform}"
    if [[ $platform == *"windows"* ]]; then
        filename="${filename}.exe"
    fi
    
    local download_url="${REPO_URL}/releases/download/${version}/${filename}.tar.gz"
    local archive_path="$temp_dir/${filename}.tar.gz"
    
    echo -e "${YELLOW}Downloading from: $download_url${NC}"
    
    # Download the release
    if ! download_file "$download_url" "$archive_path"; then
        echo -e "${RED}Failed to download SSH Tunnel Manager${NC}"
        exit 1
    fi
    
    # Extract the archive
    echo -e "${BLUE}Extracting archive...${NC}"
    tar -xzf "$archive_path" -C "$temp_dir"
    
    # Find the binary
    local binary_path
    if [[ $platform == *"windows"* ]]; then
        binary_path="$temp_dir/${BINARY_NAME}.exe"
    else
        binary_path="$temp_dir/${BINARY_NAME}"
    fi
    
    if [[ ! -f "$binary_path" ]]; then
        # Try to find the binary with platform suffix
        binary_path="$temp_dir/${filename}"
        if [[ ! -f "$binary_path" ]]; then
            echo -e "${RED}Binary not found in archive${NC}"
            exit 1
        fi
    fi
    
    # Make binary executable
    chmod +x "$binary_path"
    
    # Install binary
    if [[ "$EUID" -eq 0 ]] || [[ -w "$INSTALL_DIR" ]]; then
        # Can install directly
        cp "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
        echo -e "${GREEN}✓ Installed to $INSTALL_DIR/$BINARY_NAME${NC}"
    else
        # Need sudo
        echo -e "${YELLOW}Installing to $INSTALL_DIR (requires sudo)...${NC}"
        sudo cp "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
        echo -e "${GREEN}✓ Installed to $INSTALL_DIR/$BINARY_NAME${NC}"
    fi
}

# Setup configuration directory
setup_config() {
    echo -e "${BLUE}Setting up configuration directory...${NC}"
    
    if [[ ! -d "$CONFIG_DIR" ]]; then
        mkdir -p "$CONFIG_DIR"
        chmod 755 "$CONFIG_DIR"
        echo -e "${GREEN}✓ Created configuration directory: $CONFIG_DIR${NC}"
    else
        echo -e "${YELLOW}Configuration directory already exists: $CONFIG_DIR${NC}"
    fi
    
    # Create subdirectories
    mkdir -p "$CONFIG_DIR/tunnels"
    mkdir -p "$CONFIG_DIR/backups"
    mkdir -p "$CONFIG_DIR/templates"
    
    echo -e "${GREEN}✓ Configuration setup complete${NC}"
}

# Verify installation
verify_installation() {
    echo -e "${BLUE}Verifying installation...${NC}"
    
    if command_exists "$BINARY_NAME"; then
        local installed_version
        installed_version=$("$BINARY_NAME" --version 2>/dev/null | head -n1 || echo "unknown")
        echo -e "${GREEN}✓ SSH Tunnel Manager installed successfully${NC}"
        echo -e "${CYAN}Version: $installed_version${NC}"
        echo -e "${CYAN}Location: $(which $BINARY_NAME)${NC}"
        return 0
    else
        echo -e "${RED}✗ Installation verification failed${NC}"
        echo -e "${YELLOW}You may need to add $INSTALL_DIR to your PATH${NC}"
        return 1
    fi
}

# Show completion message
show_completion() {
    echo
    echo -e "${GREEN}🎉 SSH Tunnel Manager installation completed successfully!${NC}"
    echo
    echo -e "${BLUE}Next steps:${NC}"
    echo -e "  1. Run '${CYAN}ssh-tunnel setup${NC}' to create your first tunnel"
    echo -e "  2. Use '${CYAN}ssh-tunnel --help${NC}' to see all available commands"
    echo -e "  3. Visit the documentation: ${CYAN}${REPO_URL}${NC}"
    echo
    echo -e "${BLUE}Configuration directory:${NC} $CONFIG_DIR"
    echo -e "${BLUE}Binary location:${NC} $INSTALL_DIR/$BINARY_NAME"
    echo
}

# Show usage
show_usage() {
    echo "SSH Tunnel Manager Installation Script"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  --version VERSION    Install specific version (default: latest)"
    echo "  --install-dir DIR    Install directory (default: $INSTALL_DIR)"
    echo "  --config-dir DIR     Config directory (default: $CONFIG_DIR)"
    echo "  --help              Show this help message"
    echo
    echo "Examples:"
    echo "  $0                           # Install latest version"
    echo "  $0 --version v1.2.3          # Install specific version"
    echo "  $0 --install-dir ~/bin       # Install to custom directory"
    echo
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --config-dir)
                CONFIG_DIR="$2"
                shift 2
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Main installation function
main() {
    local platform version
    
    echo -e "${BLUE}=== SSH Tunnel Manager Installation ===${NC}"
    echo
    
    # Parse arguments
    parse_args "$@"
    
    # Detect platform
    platform=$(detect_platform)
    echo -e "${CYAN}Detected platform: $platform${NC}"
    
    # Get version to install
    if [[ -z "$VERSION" ]]; then
        echo -e "${BLUE}Getting latest version...${NC}"
        VERSION=$(get_latest_version)
    fi
    echo -e "${CYAN}Installing version: $VERSION${NC}"
    echo
    
    # Run installation steps
    check_prerequisites
    install_binary "$platform" "$VERSION"
    setup_config
    
    if verify_installation; then
        show_completion
    else
        echo -e "${RED}Installation completed but verification failed${NC}"
        echo -e "${YELLOW}Please check your PATH and try running: $INSTALL_DIR/$BINARY_NAME --help${NC}"
        exit 1
    fi
}

# Run main function with all arguments
main "$@"
