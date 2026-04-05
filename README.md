# kickstart

一键初始化新电脑环境的命令行工具。

A CLI tool for bootstrapping a new machine in one command.

- 管理 dotfiles / Manage dotfiles
- 安装/更新 Go 语言 / Install/update Go language
- 克隆和更新 Git 仓库 / Clone and sync Git repositories
- 安装常用工具和软件包 / Install tools and packages
- 执行配置脚本 / Run configuration scripts

## 安装 / Installation

**Mac / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/chzealot/kickstart/main/install.sh | bash
```

**Windows:**

```powershell
irm https://raw.githubusercontent.com/chzealot/kickstart/main/install.ps1 | iex
```

## 配置 / Configuration

kickstart 通过 `~/.kickstart/` 目录管理配置，主配置文件为 `~/.kickstart/config.yaml`。可通过 `-c` 参数指定其他路径。

kickstart manages configuration via the `~/.kickstart/` directory, with `~/.kickstart/config.yaml` as the main config file. Use `-c` to specify an alternative path.

### 目录结构 / Directory Structure

```
~/.kickstart/
├── config.yaml       # Main config / 主配置文件
├── tools.yaml        # Tool list / 工具列表
├── repos.yaml        # Repository list / 仓库列表
├── darwin.yaml       # macOS specific / macOS 专属配置
└── work.yaml         # Host specific / 特定主机配置
```

### 主配置文件 / Main Config

主配置文件通过 `include` 引入子配置文件，所有配置合并后生效。

The main config file uses `include` to import sub-config files. All configs are merged before taking effect.

```yaml
# ~/.kickstart/config.yaml

include:
  - tools.yaml
  - repos.yaml
  - darwin.yaml
  - work.yaml

dotfiles:
  repo: git@github.com:yourname/dotfiles.git

go: latest
```

### 子配置文件 / Sub-config Files

每个子配置文件格式与主配置文件完全一致，也支持嵌套 `include`。include 路径为相对路径时，相对于当前文件所在目录。

Sub-config files share the same format as the main config and support nested `include`. Relative paths are resolved from the directory of the current file.

`tools.yaml`:
```yaml
tools:
  - git
  - curl
  - jq
  - ripgrep
  - fzf
```

`repos.yaml`:
```yaml
repos:
  - url: git@github.com:yourname/project.git
    path: ~/workspace/project
  - url: https://github.com/yourname/notes.git
    path: ~/notes
```

### 平台特定配置 / Platform-specific Config

通过 `darwin`、`linux`、`windows` 顶级 key 为不同平台定制配置。

Use top-level keys `darwin`, `linux`, or `windows` to customize per platform.

`darwin.yaml`:
```yaml
darwin:
  tools:
    - coreutils
  repos:
    - url: git@github.com:yourname/mac-scripts.git
      path: ~/mac-scripts
```

### 主机名特定配置 / Host-specific Config

通过 `hosts` 为指定主机定制配置，key 支持 `*` 和 `?` 通配符。

Use `hosts` to customize per hostname. Keys support `*` and `?` wildcards.

`work.yaml`:
```yaml
hosts:
  my-macbook:
    tools:
      - ffmpeg
  "dev-*":
    tools:
      - docker
    repos:
      - url: git@github.com:company/infra.git
        path: ~/workspace/infra
```

### 合并规则 / Merge Rules

配置按 **include → 通用 → 平台 → 主机名** 的顺序逐层合并：

Configs are merged in order: **include → general → platform → hostname**.

- `tools`、`repos`、`scripts`: 追加合并 / appended
- `dotfiles`、`go`: 后者覆盖前者 / later values override earlier ones

### 字段说明 / Field Reference

| 字段 / Field | 说明 / Description |
|---|---|
| `include` | 子配置文件列表 / Sub-config file list (relative or absolute paths) |
| `dotfiles.repo` | Dotfiles 仓库地址，以 bare repo 方式部署到 `~/.git` / Dotfiles repo URL, deployed as bare repo to `~/.git` |
| `go` | Go 语言安装（`latest` 安装最新稳定版）/ Go installation (`latest` installs latest stable from go.dev) |
| `repos[].url` | Git 仓库地址 / Git repository URL |
| `repos[].path` | 本地目标路径（支持 `~`），不存在时 clone，已存在时 pull / Local path (`~` supported), cloned if missing, pulled if exists |
| `tools[]` | 工具名称，通过系统包管理器安装 / Tool name, installed via system package manager (brew on macOS, auto-detected on Linux) |
| `scripts[].name` | 脚本任务显示名称 / Script task display name |
| `scripts[].run` | 要执行的 shell 命令 / Shell command to execute |

### 兼容性 / Compatibility

- `-c` 指定文件时直接加载该文件 / When `-c` points to a file, it is loaded directly
- `-c` 指定目录时加载目录下的 `config.yaml` / When `-c` points to a directory, `config.yaml` inside it is loaded
- 如果 `~/.kickstart` 是文件（旧格式）而非目录，仍然能正常加载 / If `~/.kickstart` is a file (legacy format) instead of a directory, it still loads correctly

## 使用 / Usage

```bash
# 执行全部初始化流程 / Run all initialization steps
kickstart

# 安装/更新 Go 语言 / Install/update Go
kickstart go

# 仅安装工具 / Install tools only
kickstart install

# 仅同步 Git 仓库 / Sync Git repositories only
kickstart repos

# 执行配置脚本 / Run configuration scripts
kickstart scripts

# 查看环境状态 / View environment status
kickstart status

# 预览模式（不实际执行）/ Dry-run mode (no actual execution)
kickstart --dry-run
```

## 许可证 / License

[MIT](LICENSE)
