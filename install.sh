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

# Download archive
ARCHIVE_NAME="kickstart-${OS}-${ARCH}.tar.gz"
DIR_NAME="kickstart-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE_NAME}"

info "下载 ${ARCHIVE_NAME}..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

if [ -n "${GITHUB_TOKEN:-}" ]; then
    curl -fsSL \
        -H "Authorization: token $GITHUB_TOKEN" \
        -o "$TMPDIR/$ARCHIVE_NAME" \
        "$DOWNLOAD_URL"
else
    curl -fsSL -o "$TMPDIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"
fi

# Extract
info "解压..."
tar xzf "$TMPDIR/$ARCHIVE_NAME" -C "$TMPDIR"

BINARY_PATH="$TMPDIR/$DIR_NAME/$BINARY"
if [ ! -f "$BINARY_PATH" ]; then
    error "解压后未找到二进制文件: $BINARY_PATH"
fi

chmod +x "$BINARY_PATH"

# Install
info "安装到 ${INSTALL_DIR}/${BINARY}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
else
    sudo mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
fi

success "安装成功！"
echo ""
info "安装路径: ${INSTALL_DIR}/${BINARY}"

# Check if INSTALL_DIR is in PATH
case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        warn "${INSTALL_DIR} 不在 PATH 中，请添加到 shell 配置文件："
        echo ""
        echo "  # bash (~/.bashrc)"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        echo "  # zsh (~/.zshrc)"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        echo "  # fish (~/.config/fish/config.fish)"
        echo "  fish_add_path ${INSTALL_DIR}"
        echo ""
        info "添加后运行 source ~/.bashrc（或对应配置文件）使其生效"
        ;;
esac

info "运行 ${BINARY} --version 验证安装"
