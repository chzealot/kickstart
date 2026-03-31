# kickstart

一键初始化新电脑环境的命令行工具。

- 管理 dotfiles
- 安装常用工具和软件包
- 配置软件和系统偏好设置

## 安装

> 需要先设置 `GITHUB_TOKEN` 环境变量（需有 repo 权限的 [Personal Access Token](https://github.com/settings/tokens)）。

**Mac / Linux：**

```bash
curl -fsSL -H "Authorization: token $GITHUB_TOKEN" https://raw.githubusercontent.com/chzealot/kickstart/main/install.sh | bash
```

**Windows：**

```powershell
irm -Headers @{Authorization="token $env:GITHUB_TOKEN"} https://raw.githubusercontent.com/chzealot/kickstart/main/install.ps1 | iex
```
