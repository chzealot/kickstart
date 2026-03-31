$ErrorActionPreference = "Stop"

$Repo = "chzealot/kickstart"
$Binary = "kickstart.exe"
$InstallDir = "$env:LOCALAPPDATA\kickstart"

function Info($msg)    { Write-Host "i $msg" -ForegroundColor Cyan }
function Success($msg) { Write-Host "√ $msg" -ForegroundColor Green }
function Warn($msg)    { Write-Host "! $msg" -ForegroundColor Yellow }
function Error($msg)   { Write-Host "x $msg" -ForegroundColor Red; exit 1 }

# Check GITHUB_TOKEN
if (-not $env:GITHUB_TOKEN) {
    Error "请设置 GITHUB_TOKEN 环境变量（需要 repo 权限的 Personal Access Token）"
}

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $Arch = "arm64" }

Info "检测到系统: windows/${Arch}"

# Get latest release
Info "获取最新版本..."
$Headers = @{ Authorization = "token $env:GITHUB_TOKEN" }
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers $Headers
$Tag = $Release.tag_name

if (-not $Tag) {
    Error "无法获取最新版本"
}

Info "最新版本: $Tag"

# Find asset
$AssetName = "kickstart_windows_${Arch}.exe"
$Asset = $Release.assets | Where-Object { $_.name -eq $AssetName }

if (-not $Asset) {
    Error "未找到对应平台的构建产物: $AssetName"
}

# Download
Info "下载 ${AssetName}..."
$TmpFile = [System.IO.Path]::GetTempFileName()
$DownloadHeaders = @{
    Authorization = "token $env:GITHUB_TOKEN"
    Accept = "application/octet-stream"
}
Invoke-WebRequest -Uri $Asset.url -Headers $DownloadHeaders -OutFile $TmpFile

# Install
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Move-Item -Force $TmpFile "$InstallDir\$Binary"

# Add to PATH if needed
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    Warn "已将 $InstallDir 添加到 PATH（重启终端生效）"
}

Success "安装成功！"
Info "运行 kickstart --version 验证安装"
