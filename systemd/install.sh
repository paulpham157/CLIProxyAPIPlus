#!/bin/bash
# Installation script for CLIProxyAPI Plus systemd service
# This script automates the installation process described in INSTALL.md

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="cli-proxy-api-plus"
SERVICE_USER="cliproxy"
INSTALL_DIR="/opt/${SERVICE_NAME}"
DATA_DIR="/var/lib/${SERVICE_NAME}"
LOG_DIR="/var/log/${SERVICE_NAME}"
CONFIG_DIR="/etc/${SERVICE_NAME}"
BINARY_NAME="${SERVICE_NAME}"

# Functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

check_binary() {
    if [ ! -f "${BINARY_NAME}" ]; then
        print_error "Binary '${BINARY_NAME}' not found in current directory"
        print_info "Please build the application first: go build -o ${BINARY_NAME} ./cmd/server"
        exit 1
    fi
}

create_user() {
    if id "${SERVICE_USER}" &>/dev/null; then
        print_info "User '${SERVICE_USER}' already exists"
    else
        print_info "Creating system user '${SERVICE_USER}'..."
        useradd --system --no-create-home --shell /bin/false "${SERVICE_USER}"
    fi
}

create_directories() {
    print_info "Creating directory structure..."
    mkdir -p "${INSTALL_DIR}"
    mkdir -p "${DATA_DIR}"
    mkdir -p "${LOG_DIR}"
    mkdir -p "${CONFIG_DIR}"
}

install_files() {
    print_info "Installing application files..."
    
    # Copy binary
    cp "${BINARY_NAME}" "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    
    # Copy config if not exists
    if [ ! -f "${INSTALL_DIR}/config.yaml" ]; then
        if [ -f "config.yaml" ]; then
            print_info "Copying existing config.yaml..."
            cp config.yaml "${INSTALL_DIR}/"
        elif [ -f "config.example.yaml" ]; then
            print_info "Copying config.example.yaml as config.yaml..."
            cp config.example.yaml "${INSTALL_DIR}/config.yaml"
        else
            print_warn "No config file found. You'll need to create ${INSTALL_DIR}/config.yaml manually"
        fi
    else
        print_info "Existing config.yaml found, keeping it"
    fi
    
    # Copy environment example if not exists
    if [ ! -f "${CONFIG_DIR}/environment" ]; then
        if [ -f "systemd/environment.example" ]; then
            print_info "Copying environment.example (you can edit ${CONFIG_DIR}/environment)..."
            cp systemd/environment.example "${CONFIG_DIR}/environment.example"
        fi
    fi
}

set_permissions() {
    print_info "Setting permissions..."
    chown -R "${SERVICE_USER}:${SERVICE_USER}" "${INSTALL_DIR}"
    chown -R "${SERVICE_USER}:${SERVICE_USER}" "${DATA_DIR}"
    chown -R "${SERVICE_USER}:${SERVICE_USER}" "${LOG_DIR}"
    
    # Config directory should be readable by service but owned by root
    chown -R root:root "${CONFIG_DIR}"
    if [ -f "${CONFIG_DIR}/environment" ]; then
        chmod 600 "${CONFIG_DIR}/environment"
    fi
}

install_service() {
    print_info "Installing systemd service..."
    
    if [ ! -f "${SERVICE_NAME}.service" ]; then
        print_error "Service file '${SERVICE_NAME}.service' not found"
        exit 1
    fi
    
    cp "${SERVICE_NAME}.service" "/etc/systemd/system/"
    systemctl daemon-reload
}

enable_service() {
    print_info "Enabling service..."
    systemctl enable "${SERVICE_NAME}"
}

start_service() {
    print_info "Starting service..."
    systemctl start "${SERVICE_NAME}"
    sleep 2
    
    if systemctl is-active --quiet "${SERVICE_NAME}"; then
        print_info "Service started successfully!"
    else
        print_error "Service failed to start. Check logs with: journalctl -u ${SERVICE_NAME} -n 50"
        exit 1
    fi
}

show_status() {
    echo ""
    print_info "Service Status:"
    systemctl status "${SERVICE_NAME}" --no-pager || true
    echo ""
}

print_next_steps() {
    echo ""
    print_info "Installation completed successfully!"
    echo ""
    echo "Next steps:"
    echo "  1. Edit configuration: ${INSTALL_DIR}/config.yaml"
    echo "  2. Set environment variables (if needed): ${CONFIG_DIR}/environment"
    echo "  3. Restart service: sudo systemctl restart ${SERVICE_NAME}"
    echo ""
    echo "Useful commands:"
    echo "  - View logs: sudo journalctl -u ${SERVICE_NAME} -f"
    echo "  - Check status: sudo systemctl status ${SERVICE_NAME}"
    echo "  - Restart: sudo systemctl restart ${SERVICE_NAME}"
    echo "  - Stop: sudo systemctl stop ${SERVICE_NAME}"
    echo ""
}

# Main installation flow
main() {
    print_info "Starting installation of ${SERVICE_NAME} systemd service..."
    echo ""
    
    check_root
    check_binary
    create_user
    create_directories
    install_files
    set_permissions
    install_service
    enable_service
    start_service
    show_status
    print_next_steps
}

# Run main function
main
