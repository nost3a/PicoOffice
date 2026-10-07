#requires -Version 5.1
<#
.SYNOPSIS
  PicoOffice installer (Windows / PowerShell)

.DESCRIPTION
  Downloads the Windows zip from GitHub Release and extracts it locally;
  optionally creates a desktop shortcut or uninstalls.

  Release asset naming (kept in sync with install.sh):
    picooffice_<tag>_windows_amd64.zip   contains picooffice.exe

.EXAMPLE
  # install latest
  powershell -ExecutionPolicy Bypass -File install.ps1

.EXAMPLE
  # install a specific version and create a desktop shortcut
  powershell -ExecutionPolicy Bypass -File install.ps1 -Version v0.3.0 -Shortcut

.EXAMPLE
  # dry-run only, no side effects
  powershell -ExecutionPolicy Bypass -File install.ps1 -DryRun

.EXAMPLE
  # uninstall
  powershell -ExecutionPolicy Bypass -File install.ps1 -Uninstall
#>
param(
  [string]$Repo      = 'nost3a/PicoOffice',
  [string]$Version   = 'latest',
  [string]$InstallDir = "$env:LOCALAPPDATA\PicoOffice",
  [switch]$Shortcut,
  [switch]$Uninstall,
  [switch]$DryRun
)

$ErrorActionPreference = 'Stop'

function Write-Step($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Write-Plan($msg) { if ($DryRun) { Write-Host "  [dry-run] $msg" -ForegroundColor DarkGray } }

# ---------------- uninstall ----------------
if ($Uninstall) {
  Write-Step "卸载 PicoOffice"

  if (Test-Path $InstallDir) {
    Write-Plan "Remove-Item -Recurse -Force '$InstallDir'"
    if (-not $DryRun) { Remove-Item -Recurse -Force $InstallDir }
  } else {
    Write-Host "  安装目录不存在，跳过: $InstallDir"
  }

  $lnk = Join-Path ([Environment]::GetFolderPath('Desktop')) 'PicoOffice.lnk'
  if (Test-Path $lnk) {
    Write-Plan "Remove-Item '$lnk'"
    if (-not $DryRun) { Remove-Item $lnk }
  }

  Write-Host "[OK] 卸载完成" -ForegroundColor Green
  exit 0
}

# ---------------- resolve version ----------------
$tag = $Version
if ($Version -eq 'latest') {
  if ($DryRun) {
    Write-Host "  [dry-run] 跳过 GitHub API 查询；实际运行时请求 https://api.github.com/repos/$Repo/releases/latest" -ForegroundColor DarkGray
    $tag = 'latest'
  } else {
    try {
      Write-Step "查询最新 Release ..."
      $rel = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" `
              -Headers @{ 'User-Agent' = 'picooffice-installer' }
      $tag = $rel.tag_name
    } catch {
      Write-Error "无法解析 latest 版本: $($_.Exception.Message)。请用 -Version 指定 tag（如 v0.3.0），" +
                  "或确认 -Repo $Repo 是否正确。"
      exit 1
    }
  }
}

$asset = "picooffice_${tag}_windows_amd64.zip"
$url   = "https://github.com/$Repo/releases/download/$tag/$asset"

Write-Step "仓库: $Repo | 版本: $tag"
Write-Step "安装目录: $InstallDir"

# ---------------- dry-run: print only ----------------
if ($DryRun) {
  $tmpZip = Join-Path $env:TEMP $asset
  Write-Plan "Invoke-WebRequest -Uri '$url' -OutFile '$tmpZip'"
  Write-Plan "Expand-Archive -Path '$tmpZip' -DestinationPath '$InstallDir' -Force"
  if ($Shortcut) {
    Write-Plan "创建桌面快捷方式 -> '$InstallDir\picooffice.exe'"
  }
  Write-Host "[OK] dry-run 完成，未做任何实际改动" -ForegroundColor Green
  exit 0
}

# ---------------- download ----------------
$tmpZip = Join-Path $env:TEMP $asset
try {
  Write-Step "下载 $asset"
  Invoke-WebRequest -Uri $url -OutFile $tmpZip -UseBasicParsing
} catch {
  Write-Warning "下载失败: $($_.Exception.Message)"
  Write-Host "  若 Windows 暂未发布预编译二进制，可尝试源码安装:" -ForegroundColor Yellow
  Write-Host "    go install github.com/$Repo/backend/cmd/picooffice@$tag" -ForegroundColor Yellow
  exit 1
}

# ---------------- extract ----------------
if (-not (Test-Path $InstallDir)) {
  New-Item -ItemType Directory -Path $InstallDir | Out-Null
}
Write-Step "解压到 $InstallDir"
Expand-Archive -Path $tmpZip -DestinationPath $InstallDir -Force
Remove-Item $tmpZip -Force

# ---------------- desktop shortcut ----------------
if ($Shortcut) {
  $exe = Join-Path $InstallDir 'picooffice.exe'
  if (Test-Path $exe) {
    $lnkPath = Join-Path ([Environment]::GetFolderPath('Desktop')) 'PicoOffice.lnk'
    $ws = New-Object -ComObject WScript.Shell
    $sc = $ws.CreateShortcut($lnkPath)
    $sc.TargetPath       = $exe
    $sc.WorkingDirectory = $InstallDir
    $sc.Description      = 'PicoOffice'
    $sc.Save()
    Write-Step "已创建桌面快捷方式: $lnkPath"
  } else {
    Write-Warning "未找到 $exe，跳过创建快捷方式"
  }
}

# ---------------- done ----------------
Write-Host ""
Write-Host "[OK] PicoOffice 安装成功！" -ForegroundColor Green
Write-Host "  启动: $InstallDir\picooffice.exe"
Write-Host "  访问: http://localhost:8080"
