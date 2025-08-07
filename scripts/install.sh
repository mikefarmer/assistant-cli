#!/bin/bash

set -e

# Assistant-CLI Installation Script
# This script detects the platform and architecture, then downloads and installs
# the appropriate binary for Assistant-CLI.

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="mikefarmer/assistant-cli"
BINARY_NAME="assistant-cli"
INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/.local/bin"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Function to detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    # Detect OS - Only macOS supported
    case "$(uname -s)" in
        Darwin*)
            os="darwin"
            ;;
        *)
            print_error "Unsupported operating system: $(uname -s)"
            print_error "Assistant-CLI currently only supports macOS"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Function to get latest release version
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        print_error "Failed to get latest version"
        exit 1
    fi
    echo "$version"
}

# Function to download binary
download_binary() {
    local platform=$1
    local version=$2
    local binary_name="${BINARY_NAME}-${platform}"
    local url="https://github.com/${REPO}/releases/download/${version}/${binary_name}"
    
    if [ "$platform" = "windows-amd64" ]; then
        binary_name="${BINARY_NAME}-${platform}.exe"
        url="https://github.com/${REPO}/releases/download/${version}/${binary_name}"
    fi
    
    print_step "Downloading ${binary_name} from ${url}"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "/tmp/${BINARY_NAME}" "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "/tmp/${BINARY_NAME}" "$url"
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    # Make executable
    chmod +x "/tmp/${BINARY_NAME}"
}

# Function to install binary
install_binary() {
    local install_path=""
    local needs_sudo=false
    
    # Try system-wide installation first
    if [ -w "$INSTALL_DIR" ] || [ "$(id -u)" = "0" ]; then
        install_path="$INSTALL_DIR"
        needs_sudo=false
    elif sudo -n true 2>/dev/null; then
        install_path="$INSTALL_DIR"
        needs_sudo=true
    else
        # Fall back to user installation
        print_warning "Cannot install system-wide. Installing to user directory: $USER_INSTALL_DIR"
        install_path="$USER_INSTALL_DIR"
        needs_sudo=false
        
        # Create user bin directory if it doesn't exist
        mkdir -p "$USER_INSTALL_DIR"
        
        # Check if user bin is in PATH
        if [[ ":$PATH:" != *":$USER_INSTALL_DIR:"* ]]; then
            print_warning "$USER_INSTALL_DIR is not in your PATH."
            print_status "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
            echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        fi
    fi
    
    print_step "Installing to $install_path"
    
    if [ "$needs_sudo" = true ]; then
        sudo mv "/tmp/${BINARY_NAME}" "${install_path}/${BINARY_NAME}"
    else
        mv "/tmp/${BINARY_NAME}" "${install_path}/${BINARY_NAME}"
    fi
    
    print_status "Assistant-CLI installed successfully to ${install_path}/${BINARY_NAME}"
}

# Function to verify installation
verify_installation() {
    local install_path=""
    
    # Find the installed binary
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        install_path="${INSTALL_DIR}/${BINARY_NAME}"
    elif [ -f "${USER_INSTALL_DIR}/${BINARY_NAME}" ]; then
        install_path="${USER_INSTALL_DIR}/${BINARY_NAME}"
    else
        print_error "Installation verification failed: binary not found"
        exit 1
    fi
    
    print_step "Verifying installation"
    
    # Test binary execution
    if "$install_path" --version >/dev/null 2>&1; then
        local version
        version=$("$install_path" --version | cut -d' ' -f3)
        print_status "Verification successful! Assistant-CLI $version is ready to use."
    else
        print_error "Installation verification failed: binary cannot execute"
        exit 1
    fi
}

# Function to show next steps
show_next_steps() {
    echo ""
    print_status "Installation complete! Next steps:"
    echo ""
    echo "1. Set up authentication:"
    echo "   export ASSISTANT_CLI_API_KEY=\"your-google-cloud-api-key\""
    echo ""
    echo "2. Test the installation:"
    echo "   ${BINARY_NAME} --version"
    echo "   ${BINARY_NAME} --help"
    echo ""
    echo "3. Try text-to-speech:"
    echo "   echo \"Hello, World!\" | ${BINARY_NAME} synthesize -o hello.mp3"
    echo ""
    echo "4. For help and documentation:"
    echo "   https://github.com/${REPO}#readme"
    echo ""
}

# Main installation function
main() {
    print_status "Starting Assistant-CLI installation"
    
    # Check for required commands
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        print_error "This script requires either curl or wget to be installed"
        exit 1
    fi
    
    # Detect platform
    print_step "Detecting platform"
    local platform
    platform=$(detect_platform)
    print_status "Detected platform: $platform"
    
    # Get latest version
    print_step "Getting latest version"
    local version
    version=$(get_latest_version)
    print_status "Latest version: $version"
    
    # Download binary
    download_binary "$platform" "$version"
    
    # Install binary
    install_binary
    
    # Verify installation
    verify_installation
    
    # Show next steps
    show_next_steps
    
    print_status "Installation completed successfully!"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Assistant-CLI Installation Script"
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --version, -v  Show version and exit"
        echo ""
        echo "Environment Variables:"
        echo "  INSTALL_DIR    Override installation directory (default: /usr/local/bin)"
        echo ""
        echo "Example:"
        echo "  curl -s https://raw.githubusercontent.com/mikefarmer/assistant-cli/main/scripts/install.sh | bash"
        echo ""
        exit 0
        ;;
    --version|-v)
        echo "Assistant-CLI Installation Script v1.0.0"
        exit 0
        ;;
esac

# Override install directory if provided
if [ -n "${INSTALL_DIR_OVERRIDE:-}" ]; then
    INSTALL_DIR="$INSTALL_DIR_OVERRIDE"
fi

# Run main installation
main