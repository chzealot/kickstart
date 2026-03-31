# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

kickstart 是一个用 Go 编写的命令行工具，用于一键初始化新电脑环境（dotfiles、工具安装、软件配置）。仓库为私有仓库 (github.com/chzealot/kickstart)。用户界面语言为中文。

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

**自更新**: `cmd/upgrade.go` 优先使用 `gh` CLI 下载 release，降级为 `curl` + `GITHUB_TOKEN`。

## 测试结构

- **单元测试**: `cmd/cmd_test.go`, `internal/ui/ui_test.go`, `internal/version/*_test.go` — 测试命令注册、flag、输出函数、版本逻辑
- **集成测试**: `tests/integration/cli_test.go` — 编译二进制后执行 CLI 端到端验证
- 所有测试启用 `-race` 检测。注意全局变量 `version.Version` 在测试中不要直接修改，避免 data race。

## 发布流程

在 GitHub 上创建 Release → 触发 `.github/workflows/release.yml` → 先跑测试 → 6 平台矩阵并行构建 → 上传二进制到 Release。

## 安装方式

因为是私有仓库，安装需要 `GITHUB_TOKEN`：
```bash
curl -fsSL -H "Authorization: token $GITHUB_TOKEN" https://raw.githubusercontent.com/chzealot/kickstart/main/install.sh | bash
```
