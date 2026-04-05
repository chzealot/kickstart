# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

kickstart 是一个用 Go 编写的开源命令行工具，用于一键初始化新电脑环境（dotfiles、工具安装、软件配置）。仓库地址: github.com/chzealot/kickstart，MIT 协议。用户界面语言为中文。

## 常用命令

```bash
# 构建
go build -o kickstart .

# 带版本信息构建
go build -ldflags "-X github.com/chzealot/kickstart/internal/version.Version=v0.1.0" -o kickstart .

# 运行全部单元测试（排除 integration）
./scripts/run-unit-tests.sh

# 运行集成测试
./scripts/run-integration-tests.sh

# 运行单个包的测试
go test -v -race ./internal/version/...
go test -v -race ./cmd/...

# 运行单个测试
go test -v -race -run TestCLI_Help ./tests/integration/...
```

## 架构

**CLI 框架**: Cobra。所有子命令在 `cmd/` 下各自文件中定义，通过 `init()` 注册到 `rootCmd`。`run` 是默认命令（无子命令时执行）。

**TUI 输出**: `internal/ui/` 封装了 lipgloss 样式和 bubbletea spinner。所有命令通过 `ui.Title()`, `ui.Success()`, `ui.Error()` 等函数输出，不要直接 `fmt.Println`。

**版本检测**: `internal/version/` 提供 `AsyncChecker`，在 `rootCmd.PersistentPreRun` 中异步启动，`PersistentPostRun` 中检查结果并提示升级。`Version/Commit/Date` 通过 ldflags 在构建时注入。

**配置管理**: `internal/config/` 解析 `~/.kickstart/config.yaml`（YAML 格式），支持 include 子配置文件、按平台（darwin/linux/windows）和主机名（含通配符）定制。通过 `-c` flag 可指定其他路径，兼容旧的单文件格式。

**Go 安装**: `internal/goinstall/` 提供 Go 语言的安装和更新逻辑。从 `go.dev/dl` 获取最新版本（降级到 `golang.google.cn`），下载 tar.gz 并安装到 `/usr/local/go`，包含 SHA256 校验。通过配置文件 `go: latest` 启用。

**Python 安装**: `internal/pyinstall/` 提供 Python 的安装和更新逻辑。macOS 从 python.org 下载 .pkg 安装器，Linux 使用系统包管理器。安装后自动创建 `python` → `python3` 符号链接。通过配置文件 `python: latest` 启用。

**工具安装**: `internal/installer/` 提供通用的工具安装逻辑，根据配置文件中的 tools 列表，通过系统包管理器（brew/apt-get/dnf 等）安装。

**自更新**: `cmd/upgrade.go` 优先使用 `gh` CLI 下载 release，降级为 `curl` + `GITHUB_TOKEN`。下载文件缓存到 `~/.cache/kickstart/{version}/`，支持 checksum 校验和自动清理历史版本。

## 测试结构

- **单元测试**: `cmd/cmd_test.go`, `internal/ui/ui_test.go`, `internal/version/*_test.go` — 测试命令注册、flag、输出函数、版本逻辑
- **集成测试**: `tests/integration/cli_test.go` — 编译二进制后执行 CLI 端到端验证
- 所有测试启用 `-race` 检测。注意全局变量 `version.Version` 在测试中不要直接修改，避免 data race。

## 发布流程

在 GitHub 上创建 Release → 触发 `.github/workflows/release.yml` → 先跑测试 → 6 平台矩阵并行构建 → 上传二进制到 Release。

**Release Note 规范**: Release Note 必须使用中英双语编写，中文在前、英文在后，用 `---` 分隔。

## 安装方式

```bash
curl -fsSL https://raw.githubusercontent.com/chzealot/kickstart/main/install.sh | bash
```
