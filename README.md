# kickstart

一键初始化新电脑环境的命令行工具。

- 管理 dotfiles
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

配置文件包含三个部分，均为可选：

```yaml
# ~/.kickstart

# Dotfiles 管理
# 将 dotfiles 仓库以 bare repo 方式部署到 ~/.git
dotfiles:
  repo: git@github.com:yourname/dotfiles.git

# 工具安装
# 列出需要安装的命令行工具
# macOS 使用 brew，Linux 自动检测 apt-get/dnf/pacman 等
tools:
  - rsync
  - jq
  - ripgrep
  - fzf
  - fd

# 软件配置
# 安装完成后执行的 shell 命令
configs:
  - name: zsh 默认 shell
    run: chsh -s $(which zsh)
```

### dotfiles

指定一个 Git 仓库地址，kickstart 会将其以 bare repo 方式 clone 到 `~/.git`，使 `$HOME` 目录成为工作区，直接管理 dotfiles。

### tools

列出需要安装的命令行工具名称。kickstart 会自动检测已安装的工具并跳过，未安装的通过系统包管理器安装：

- **macOS**：使用 Homebrew（需预先安装）
- **Linux**：自动检测 apt-get、dnf、yum、pacman、zypper、apk

### configs

定义安装完成后需要执行的配置命令。每项包含 `name`（显示名称）和 `run`（shell 命令）。

## 使用

```bash
# 执行全部初始化流程
kickstart

# 仅安装工具
kickstart install

# 查看环境状态
kickstart status

# 预览模式（不实际执行）
kickstart --dry-run
```

## 许可证

[MIT](LICENSE)
