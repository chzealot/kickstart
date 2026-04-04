#!/bin/bash
set -euo pipefail

REPO="chzealot/kickstart"
INSTALL_DIR="/usr/local/bin"
BINARY="kickstart"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()    { echo -e "${CYAN}ℹ${NC} $*"; }
success() { echo -e "${GREEN}✔${NC} $*"; }
warn()    { echo -e "${YELLOW}⚠${NC} $*"; }
error()   { echo -e "${RED}✘${NC} $*"; exit 1; }

# Optional: use GITHUB_TOKEN for higher rate limits
AUTH_HEADER=""
if [ -n "${GITHUB_TOKEN:-}" ]; then
    AUTH_HEADER="-H \"Authorization: token $GITHUB_TOKEN\""
fi

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "不支持的架构: $ARCH" ;;
esac

info "检测到系统: ${OS}/${ARCH}"

# Get latest release tag
info "获取最新版本..."
if [ -n "${GITHUB_TOKEN:-}" ]; then
    LATEST=$(curl -fsSL \
        -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | cut -d'"' -f4)
else
    LATEST=$(curl -fsSL \
        "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | cut -d'"' -f4)
fi

if [ -z "$LATEST" ]; then
    error "无法获取最新版本"
fi

info "最新版本: ${LATEST}"

# Download binary
ASSET="${BINARY}_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"

info "下载 ${ASSET}..."

TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT

if [ -n "${GITHUB_TOKEN:-}" ]; then
    curl -fsSL \
        -H "Authorization: token $GITHUB_TOKEN" \
        -o "$TMPFILE" \
        "$DOWNLOAD_URL"
else
    curl -fsSL -o "$TMPFILE" "$DOWNLOAD_URL"
fi

chmod +x "$TMPFILE"

# Install
info "安装到 ${INSTALL_DIR}/${BINARY}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMPFILE" "${INSTALL_DIR}/${BINARY}"
else
    sudo mv "$TMPFILE" "${INSTALL_DIR}/${BINARY}"
fi

success "安装成功！"
echo ""
info "运行 ${BINARY} --version 验证安装"
