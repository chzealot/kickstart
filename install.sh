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

# Check GITHUB_TOKEN
if [ -z "${GITHUB_TOKEN:-}" ]; then
    error "请设置 GITHUB_TOKEN 环境变量（需要 repo 权限的 Personal Access Token）"
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
LATEST=$(curl -fsSL \
    -H "Authorization: token $GITHUB_TOKEN" \
    "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
    error "无法获取最新版本"
fi

info "最新版本: ${LATEST}"

# Download binary
ASSET="${BINARY}_${OS}_${ARCH}"
DOWNLOAD_URL="https://api.github.com/repos/${REPO}/releases/latest/assets"

info "下载 ${ASSET}..."

# Get asset download URL
ASSET_URL=$(curl -fsSL \
    -H "Authorization: token $GITHUB_TOKEN" \
    "$DOWNLOAD_URL" \
    | grep -B3 "\"name\": \"${ASSET}\"" | grep '"url"' | cut -d'"' -f4)

if [ -z "$ASSET_URL" ]; then
    error "未找到对应平台的构建产物: ${ASSET}"
fi

TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT

curl -fsSL \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/octet-stream" \
    -o "$TMPFILE" \
    "$ASSET_URL"

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
