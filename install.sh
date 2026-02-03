#!/bin/bash
# oview installation script
# Usage: curl -fsSL https://raw.githubusercontent.com/.../install.sh | bash
#    or: ./install.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="oview"
REPO_URL="https://github.com/yourusername/oview"  # TODO: Update with real repo
VERSION="latest"

# Functions
print_header() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                  oview Installation                       ║"
    echo "║        Local Software Factory Environment Manager         ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_step() {
    echo -e "${GREEN}==>${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC}  $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC}  $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

check_os() {
    print_step "Checking operating system..."

    OS="$(uname -s)"
    case "$OS" in
        Linux*)     OS_TYPE="linux";;
        Darwin*)    OS_TYPE="darwin";;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64)     ARCH_TYPE="amd64";;
        aarch64)    ARCH_TYPE="arm64";;
        arm64)      ARCH_TYPE="arm64";;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    print_success "OS: $OS_TYPE, Architecture: $ARCH_TYPE"
}

check_docker() {
    print_step "Checking Docker..."

    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        echo ""
        echo "Please install Docker first:"
        case "$OS_TYPE" in
            linux)
                echo "  Ubuntu/Debian: sudo apt-get install docker.io"
                echo "  Fedora:        sudo dnf install docker"
                echo "  Arch:          sudo pacman -S docker"
                echo ""
                echo "Or visit: https://docs.docker.com/engine/install/"
                ;;
            darwin)
                echo "  brew install --cask docker"
                echo ""
                echo "Or download Docker Desktop: https://www.docker.com/products/docker-desktop"
                ;;
        esac
        exit 1
    fi

    if ! docker ps &> /dev/null; then
        print_warning "Docker is installed but not running"
        echo ""
        case "$OS_TYPE" in
            linux)
                echo "Start Docker with:"
                echo "  sudo systemctl start docker"
                echo "  sudo systemctl enable docker"
                echo ""
                echo "Add your user to docker group:"
                echo "  sudo usermod -aG docker \$USER"
                echo "  newgrp docker"
                ;;
            darwin)
                echo "Start Docker Desktop from Applications"
                ;;
        esac

        read -p "Do you want to continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        print_success "Docker is running"
    fi
}

check_go() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go $GO_VERSION is installed"
        return 0
    else
        print_info "Go is not installed (needed only for building from source)"
        return 1
    fi
}

prompt_install_method() {
    echo ""
    print_step "Choose installation method:"
    echo "  1. Download prebuilt binary (recommended, faster)"
    echo "  2. Build from source (requires Go 1.23+)"
    echo ""

    while true; do
        read -p "Select method [1-2] (default: 1): " method
        method=${method:-1}

        case $method in
            1)
                INSTALL_METHOD="binary"
                break
                ;;
            2)
                INSTALL_METHOD="source"
                if ! check_go; then
                    print_error "Go is required to build from source"
                    echo "Install Go from: https://go.dev/dl/"
                    exit 1
                fi
                break
                ;;
            *)
                print_warning "Invalid choice, please enter 1 or 2"
                ;;
        esac
    done
}

install_from_binary() {
    print_step "Downloading prebuilt binary..."

    # TODO: Update with real release URL when repo is public
    BINARY_URL="$REPO_URL/releases/download/$VERSION/oview-$OS_TYPE-$ARCH_TYPE"

    TEMP_FILE=$(mktemp)

    if command -v curl &> /dev/null; then
        curl -fsSL -o "$TEMP_FILE" "$BINARY_URL" || {
            print_error "Failed to download binary"
            print_info "Falling back to building from source..."
            install_from_source
            return
        }
    elif command -v wget &> /dev/null; then
        wget -q -O "$TEMP_FILE" "$BINARY_URL" || {
            print_error "Failed to download binary"
            print_info "Falling back to building from source..."
            install_from_source
            return
        }
    else
        print_error "Neither curl nor wget found"
        exit 1
    fi

    chmod +x "$TEMP_FILE"

    print_step "Installing binary to $INSTALL_DIR..."

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi

    print_success "Binary installed"
}

install_from_source() {
    print_step "Building from source..."

    BUILD_DIR=$(mktemp -d)
    cd "$BUILD_DIR"

    print_info "Cloning repository..."
    git clone "$REPO_URL" . || {
        print_error "Failed to clone repository"
        print_info "Trying to build from current directory..."

        if [ -f "$(dirname "$0")/go.mod" ]; then
            cd "$(dirname "$0")"
        else
            print_error "Not in oview directory"
            exit 1
        fi
    }

    print_info "Building binary..."
    go build -o "$BINARY_NAME" . || {
        print_error "Build failed"
        exit 1
    }

    print_step "Installing binary to $INSTALL_DIR..."

    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    fi

    cd - > /dev/null
    rm -rf "$BUILD_DIR"

    print_success "Binary built and installed"
}

verify_installation() {
    print_step "Verifying installation..."

    if ! command -v "$BINARY_NAME" &> /dev/null; then
        print_error "Installation failed: $BINARY_NAME not found in PATH"
        exit 1
    fi

    VERSION_OUTPUT=$("$BINARY_NAME" version 2>&1) || {
        print_error "Binary installed but not working correctly"
        exit 1
    }

    print_success "Installation verified: $VERSION_OUTPUT"
}

setup_infrastructure() {
    echo ""
    print_step "Setting up oview infrastructure..."
    echo ""

    read -p "Do you want to run 'oview install' now? (Y/n) " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        "$BINARY_NAME" install || {
            print_warning "Infrastructure setup failed"
            print_info "You can run 'oview install' manually later"
        }
    else
        print_info "Skipped infrastructure setup"
        print_info "Run 'oview install' when you're ready"
    fi
}

print_next_steps() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║              Installation Complete! 🎉                    ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Navigate to your project:"
    echo "     cd /path/to/your/project"
    echo ""
    echo "  2. Initialize oview (interactive):"
    echo "     oview init"
    echo ""
    echo "  3. Set up project runtime:"
    echo "     oview up"
    echo ""
    echo "  4. Index your codebase:"
    echo "     oview index"
    echo ""
    echo "Documentation:"
    echo "  - README:        https://github.com/yourusername/oview/blob/main/README.md"
    echo "  - Interactive:   https://github.com/yourusername/oview/blob/main/INTERACTIVE-INIT.md"
    echo "  - Embeddings:    https://github.com/yourusername/oview/blob/main/EMBEDDINGS-SIMPLE.md"
    echo ""
    echo "Commands:"
    echo "  oview --help     Show all commands"
    echo "  oview version    Show version"
    echo "  oview uninstall  Remove infrastructure"
    echo ""
}

uninstall() {
    print_header
    print_step "Uninstalling oview..."
    echo ""

    # Check if oview is installed
    if command -v oview &> /dev/null; then
        # Try to run oview uninstall
        read -p "Remove Docker infrastructure? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            oview uninstall --keep-config || print_warning "Failed to uninstall infrastructure"
        fi
    fi

    # Remove binary
    print_step "Removing binary..."
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        if [ -w "$INSTALL_DIR" ]; then
            rm "$INSTALL_DIR/$BINARY_NAME"
        else
            sudo rm "$INSTALL_DIR/$BINARY_NAME"
        fi
        print_success "Binary removed"
    else
        print_info "Binary not found"
    fi

    # Remove config (optional)
    read -p "Remove configuration (~/.oview)? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf ~/.oview
        print_success "Configuration removed"
    fi

    echo ""
    print_success "Uninstallation complete"
}

# Main script
main() {
    # Check for uninstall flag
    if [ "$1" = "uninstall" ] || [ "$1" = "--uninstall" ]; then
        uninstall
        exit 0
    fi

    print_header

    # Preflight checks
    check_os
    check_docker

    # If we're in the oview source directory and go.mod exists, offer to build
    if [ -f "$(dirname "$0")/go.mod" ] && [ -f "$(dirname "$0")/main.go" ]; then
        print_info "Detected oview source directory"
        INSTALL_METHOD="source"

        read -p "Build and install from current directory? (Y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Nn]$ ]]; then
            prompt_install_method
        fi
    else
        prompt_install_method
    fi

    # Install
    if [ "$INSTALL_METHOD" = "binary" ]; then
        install_from_binary
    else
        install_from_source
    fi

    # Verify
    verify_installation

    # Setup infrastructure
    setup_infrastructure

    # Next steps
    print_next_steps
}

# Run main function
main "$@"
