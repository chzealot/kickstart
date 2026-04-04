$ErrorActionPreference = "Stop"

$Repo = "chzealot/kickstart"
$Binary = "kickstart.exe"
$InstallDir = "$env:LOCALAPPDATA\kickstart"

function Info($msg)    { Write-Host "i $msg" -ForegroundColor Cyan }
function Success($msg) { Write-Host "√ $msg" -ForegroundColor Green }
function Warn($msg)    { Write-Host "! $msg" -ForegroundColor Yellow }
function Error($msg)   { Write-Host "x $msg" -ForegroundColor Red; exit 1 }

# Optional: use GITHUB_TOKEN for higher rate limits
$Headers = @{}
if ($env:GITHUB_TOKEN) {
    $Headers["Authorization"] = "token $env:GITHUB_TOKEN"
}

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $Arch = "arm64" }

Info "检测到系统: windows/${Arch}"

# Get latest release
Info "获取最新版本..."
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
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$AssetName"
Invoke-WebRequest -Uri $DownloadUrl -Headers $Headers -OutFile $TmpFile

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
Write-Host ""
Info "安装路��: $InstallDir\$Binary"

# Check if InstallDir is in current session PATH
if ($env:PATH -notlike "*$InstallDir*") {
    Warn "当前终端 PATH 尚未包含 $InstallDir，请重启终端或执行："
    Write-Host ""
    Write-Host "  `$env:PATH += `";$InstallDir`""
    Write-Host ""
}

Info "运行 kickstart --version 验证安装"
