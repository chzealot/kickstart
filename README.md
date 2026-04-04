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

kickstart 通过 `~/.kickstart/` 目录管理配置，主配置文件为 `~/.kickstart/config.yaml`。可通过 `-c` 参数指定其他路径。

### 目录结构

```
~/.kickstart/
├── config.yaml       # 主配置文件
├── tools.yaml        # 工具列表
├── repos.yaml        # 仓库列表
├── darwin.yaml       # macOS 专属配置
└── work.yaml         # 特定主机配置
```

### 主配置文件

主配置文件通过 `include` 引入子配置文件，所有配置合并后生效：

```yaml
# ~/.kickstart/config.yaml

include:
  - tools.yaml
  - repos.yaml
  - darwin.yaml
  - work.yaml

dotfiles:
  repo: git@github.com:yourname/dotfiles.git
```

### 子配置文件

每个子配置文件格式与主配置文件完全一致，也支持嵌套 `include`。include 路径为相对路径时，相对于当前文件所在目录。

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

### 平台特定配置

通过 `darwin`、`linux`、`windows` 顶级 key 为不同平台定制配置：

`darwin.yaml`:
```yaml
darwin:
  tools:
    - coreutils
  repos:
    - url: git@github.com:yourname/mac-scripts.git
      path: ~/mac-scripts
```

### 主机名特定配置

通过 `hosts` 为指定主机定制配置，key 支持 `*` 和 `?` 通配符：

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

### 合并规则

配置按 **include → 通用 → 平台 → 主机名** 的顺序逐层合并：

- `tools`、`repos`、`configs`：追加合并
- `dotfiles`：后者覆盖前者

### 各字段说明

| 字段 | 说明 |
|------|------|
| `include` | 子配置文件列表（相对路径或绝对路径） |
| `dotfiles.repo` | Dotfiles 仓库地址，以 bare repo 方式部署到 `~/.git` |
| `repos[].url` | Git 仓库地址 |
| `repos[].path` | 本地目标路径（支持 `~`），不存在时 clone，已存在时 pull |
| `tools[]` | 工具名称，通过系统包管理器安装（macOS 用 brew，Linux 自动检测） |
| `configs[].name` | 配置任务显示名称 |
| `configs[].run` | 要执行的 shell 命令 |

### 兼容性

- `-c` 指定文件时直接加载该文件
- `-c` 指定目录时加载目录下的 `config.yaml`
- 如果 `~/.kickstart` 是文件（旧格式）而非目录，仍然能正常加载

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
