# kickstart

一键初始化新电脑环境的命令行工具。

- 管理 dotfiles
- 克隆和更新 Git 仓库
- 安装常用工具和软件包
- 配置软件和系统偏好设置

## 安装

**Mac / Linux：**

```bash
curl -fsSL https://raw.githubusercontent.com/chzealot/kickstart/main/install.sh | bash
```

**Windows：**

```powershell
irm https://raw.githubusercontent.com/chzealot/kickstart/main/install.ps1 | iex
```

## 配置

kickstart 通过 `~/.kickstart` 配置文件驱动，使用 YAML 格式。可通过 `-c` 参数指定其他路径。

### 基础配置

配置文件包含四个部分，均为可选：

```yaml
# ~/.kickstart

# Dotfiles 管理
# 将 dotfiles 仓库以 bare repo 方式部署到 ~/.git
dotfiles:
  repo: git@github.com:yourname/dotfiles.git

# Git 仓库列表
# 自动 clone 或更新指定的 Git 仓库
repos:
  - url: git@github.com:yourname/project.git
    path: ~/workspace/project
  - url: https://github.com/yourname/notes.git
    path: ~/notes

# 工具安装
# 列出需要安装的命令行工具
# macOS 使用 brew，Linux 自动检测 apt-get/dnf/pacman 等
tools:
  - rsync
  - jq
  - ripgrep
  - fzf

# 软件配置
# 安装完成后执行的 shell 命令
configs:
  - name: zsh 默认 shell
    run: chsh -s $(which zsh)
```

### 平台特定配置

通过 `darwin`、`linux`、`windows` 顶级 key 为不同平台定制配置，会与通用配置合并：

```yaml
tools:
  - git
  - curl

darwin:
  tools:
    - coreutils
  repos:
    - url: git@github.com:yourname/mac-scripts.git
      path: ~/mac-scripts

linux:
  tools:
    - build-essential
```

### 主机名特定配置

通过 `hosts` 为指定主机定制配置，key 支持 `*` 和 `?` 通配符：

```yaml
tools:
  - git

hosts:
  my-macbook:
    tools:
      - ffmpeg
    repos:
      - url: git@github.com:yourname/private.git
        path: ~/private
  "dev-*":
    tools:
      - docker
```

### 合并规则

配置按 **通用 → 平台 → 主机名** 的顺序逐层合并：

- `tools`、`repos`、`configs`：追加合并
- `dotfiles`：后者覆盖前者

### 各字段说明

| 字段 | 说明 |
|------|------|
| `dotfiles.repo` | Dotfiles 仓库地址，以 bare repo 方式部署到 `~/.git` |
| `repos[].url` | Git 仓库地址 |
| `repos[].path` | 本地目标路径（支持 `~`），不存在时 clone，已存在时 pull |
| `tools[]` | 工具名称，通过系统包管理器安装（macOS 用 brew，Linux 自动检测） |
| `configs[].name` | 配置任务显示名称 |
| `configs[].run` | 要执行的 shell 命令 |

## 使用

```bash
# 执行全部初始化流程
kickstart

# 仅安装工具
kickstart install

# 仅同步 Git 仓库
kickstart repos

# 查看环境状态
kickstart status

# 预览模式（不实际执行）
kickstart --dry-run
```

## 许可证

[MIT](LICENSE)
