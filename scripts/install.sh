#!/usr/bin/env bash
#
# PicoOffice installer (Linux / macOS)
#
# Usage:
#   scripts/install.sh [--port 8080] [--version v0.3.0] [--systemd]
#                      [--uninstall] [--dry-run] [--help]
#
# Release asset naming (kept in sync with install.ps1):
#   picooffice_<tag>_<os>_<arch>.tar.gz   contains binary picooffice
#   e.g. picooffice_v0.3.0_linux_amd64.tar.gz
#
set -euo pipefail

# ---------------- defaults ----------------
GITHUB_REPO="${GITHUB_REPO:-nost3a/PicoOffice}"   # override via GITHUB_REPO env var, e.g. nost3a/PicoOffice
INSTALL_DIR="/usr/local/bin"            # root target; non-root falls back to $HOME/.local/bin
VERSION="latest"                        # "latest" or a tag like v0.3.0
PORT="8080"
DO_SYSTEMD=0
DO_UNINSTALL=0
DRY_RUN=0

# ---------------- colored output ----------------
if [ -t 1 ]; then
  C_RED=$'\033[31m'; C_GREEN=$'\033[32m'; C_YELLOW=$'\033[33m'
  C_BLUE=$'\033[34m'; C_DIM=$'\033[2m'; C_OFF=$'\033[0m'
else
  C_RED=""; C_GREEN=""; C_YELLOW=""; C_BLUE=""; C_DIM=""; C_OFF=""
fi
info() { echo "${C_BLUE}==>${C_OFF} $*"; }
ok()   { echo "${C_GREEN}[OK]${C_OFF} $*"; }
warn() { echo "${C_YELLOW}[!]${C_OFF} $*" >&2; }
die()  { echo "${C_RED}[X] $*${C_OFF}" >&2; exit 1; }

usage() {
  cat <<'EOF'
PicoOffice 安装脚本

用法: install.sh [选项]

选项:
  --port <端口>    服务监听端口（默认 8080）
  --version <tag>  指定版本，如 v0.3.0（默认 latest，自动解析最新 Release）
  --systemd        安装并注册 systemd 服务（需 sudo）
  --uninstall      卸载：停止服务、删除二进制与 unit 文件
  --dry-run        只打印将要执行的步骤，不实际下载/修改任何文件
  -h, --help       显示本帮助

环境变量:
  GITHUB_REPO      Override default repo, format owner/name, e.g. nost3a/PicoOffice

示例:
  curl -fsSL https://raw.githubusercontent.com/nost3a/PicoOffice/main/scripts/install.sh | bash
  sudo ./install.sh --port 9000 --systemd
  ./install.sh --version v0.3.0 --dry-run
EOF
}

# ---------------- arg parsing ----------------
while [ $# -gt 0 ]; do
  case "$1" in
    --port)      PORT="${2:?--port requires an argument}"; shift 2 ;;
    --version)   VERSION="${2:?--version requires an argument}"; shift 2 ;;
    --systemd)   DO_SYSTEMD=1; shift ;;
    --uninstall) DO_UNINSTALL=1; shift ;;
    --dry-run)   DRY_RUN=1; shift ;;
    -h|--help)   usage; exit 0 ;;
    *) die "未知参数: $1（用 --help 查看用法）" ;;
  esac
done

run() {
  # dry-run prints only, no side effects
  if [ "$DRY_RUN" = 1 ]; then
    echo "  ${C_DIM}[dry-run]${C_OFF} $*"
  else
    eval "$@"
  fi
}

# ---------------- env detection ----------------
detect_os() {
  local u; u="$(uname -s)"
  case "$u" in
    Linux)  echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)      die "不支持的系统: $u" ;;
  esac
}

detect_arch() {
  local m; m="$(uname -m)"
  case "$m" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) die "不支持的架构: $m" ;;
  esac
}

OS_NAME="$(detect_os)"
ARCH="$(detect_arch)"
IS_ROOT=0
[ "$(id -u)" -eq 0 ] && IS_ROOT=1

# resolve effective install dir: root writes to /usr/local/bin; otherwise prefer it, fall back to ~/.local/bin
resolve_install_dir() {
  if [ "$IS_ROOT" = 1 ]; then
    echo "$INSTALL_DIR"
  elif [ -w "$INSTALL_DIR" ]; then
    echo "$INSTALL_DIR"
  else
    echo "$HOME/.local/bin"
  fi
}
EFFECTIVE_INSTALL_DIR="$(resolve_install_dir)"

need_sudo() {
  # wrap in sudo when non-root and touching system dirs
  if [ "$IS_ROOT" = 1 ]; then
    eval "$@"
  else
    eval "sudo $*"
  fi
}

# ---------------- uninstall ----------------
do_uninstall() {
  info "开始卸载 PicoOffice"
  run "sudo systemctl stop picooffice"
  run "sudo systemctl disable picooffice"
  run "sudo rm -f /etc/systemd/system/picooffice.service"
  run "sudo systemctl daemon-reload"
  local candidate
  for candidate in "/usr/local/bin/picooffice" "$HOME/.local/bin/picooffice"; do
    run "rm -f '$candidate'"
  done
  ok "卸载完成"
  exit 0
}

# ---------------- main flow ----------------
[ "$DO_UNINSTALL" = 1 ] && do_uninstall

info "仓库: $GITHUB_REPO | 系统: $OS_NAME | 架构: $ARCH | 端口: $PORT"
info "安装目录: $EFFECTIVE_INSTALL_DIR"

# 1) resolve version tag
TAG="$VERSION"
if [ "$VERSION" = "latest" ]; then
  if [ "$DRY_RUN" = 1 ]; then
    TAG="latest"
    warn "[dry-run] 跳过 GitHub API 查询；实际运行时会请求 https://api.github.com/repos/$GITHUB_REPO/releases/latest"
  else
    info "查询最新 Release ..."
    TAG="$(curl -fsSL "https://api.github.com/repos/$GITHUB_REPO/releases/latest" \
            | grep -m1 '"tag_name"' \
            | sed -E 's/.*"tag_name" *: *"([^"]+)".*/\1/')" || TAG=""
    [ -z "$TAG" ] && die "无法解析 latest 版本。请用 --version 指定 tag（如 v0.3.0），" \
                        "或确认 GITHUB_REPO=$GITHUB_REPO 是否正确。"
  fi
fi
ok "目标版本: $TAG"

# 2) build asset name and download
ASSET="picooffice_${TAG}_${OS_NAME}_${ARCH}.tar.gz"
URL="https://github.com/$GITHUB_REPO/releases/download/$TAG/$ASSET"
info "下载地址: $URL"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

if [ "$DRY_RUN" = 1 ]; then
  run "curl -fsSL -o '$TMPDIR/$ASSET' '$URL'"
  run "tar -xzf '$TMPDIR/$ASSET' -C '$TMPDIR'"
  run "install -m 0755 '$TMPDIR/picooffice' '$EFFECTIVE_INSTALL_DIR/picooffice'"
else
  info "下载中 ..."
  if ! curl -fsSL -o "$TMPDIR/$ASSET" "$URL"; then
    warn "下载失败：该平台暂未发布预编译二进制（$URL）"
    if command -v go >/dev/null 2>&1; then
      info "检测到本地 Go 环境，回退源码安装: go install github.com/$GITHUB_REPO/backend/cmd/picooffice@$TAG"
      exec go install "github.com/$GITHUB_REPO/backend/cmd/picooffice@$TAG"
    else
      die "未检测到 go 命令。请先安装 Go，或手动从源码构建后将二进制放到 $EFFECTIVE_INSTALL_DIR/picooffice"
    fi
  fi

  info "解压 ..."
  tar -xzf "$TMPDIR/$ASSET" -C "$TMPDIR"

  # 3) install binary
  if [ ! -d "$EFFECTIVE_INSTALL_DIR" ]; then
    run "mkdir -p '$EFFECTIVE_INSTALL_DIR'"
  fi
  if [ -w "$EFFECTIVE_INSTALL_DIR" ] || [ "$IS_ROOT" = 1 ]; then
    install -m 0755 "$TMPDIR/picooffice" "$EFFECTIVE_INSTALL_DIR/picooffice"
  else
    sudo install -m 0755 "$TMPDIR/picooffice" "$EFFECTIVE_INSTALL_DIR/picooffice"
  fi
  ok "已安装到 $EFFECTIVE_INSTALL_DIR/picooffice"
fi

# 4) optional systemd registration
if [ "$DO_SYSTEMD" = 1 ]; then
  UNIT="/etc/systemd/system/picooffice.service"
  SVC_USER="${SUDO_USER:-$(id -un)}"
  info "生成 systemd unit: $UNIT (User=$SVC_USER, PICO_PORT=$PORT)"

  if [ "$DRY_RUN" = 1 ]; then
    echo "  ${C_DIM}[dry-run]${C_OFF} sudo tee $UNIT <<'UNIT'"
    cat <<EOF
[Unit]
Description=PicoOffice
After=network.target

[Service]
Type=simple
User=$SVC_USER
Environment=PICO_PORT=$PORT
ExecStart=$EFFECTIVE_INSTALL_DIR/picooffice
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF
    echo "  ${C_DIM}[dry-run]${C_OFF} UNIT"
    echo "  ${C_DIM}[dry-run]${C_OFF} sudo systemctl daemon-reload && sudo systemctl enable --now picooffice${C_OFF}"
  else
    need_sudo tee "$UNIT" >/dev/null <<UNIT
[Unit]
Description=PicoOffice
After=network.target

[Service]
Type=simple
User=$SVC_USER
Environment=PICO_PORT=$PORT
ExecStart=$EFFECTIVE_INSTALL_DIR/picooffice
Restart=on-failure

[Install]
WantedBy=multi-user.target
UNIT
    need_sudo systemctl daemon-reload
    need_sudo systemctl enable --now picooffice
    ok "systemd 服务已启动并设置开机自启"
  fi
fi

# ---------------- done ----------------
echo
if [ "$DRY_RUN" = 1 ]; then
  ok "dry-run 完成，未做任何实际改动"
  echo
  echo "${C_BLUE}将执行的下一步:${C_OFF}"
  echo "  启动: PICO_PORT=$PORT $EFFECTIVE_INSTALL_DIR/picooffice"
  if [ "$DO_SYSTEMD" = 1 ]; then
    echo "  服务: sudo systemctl status picooffice"
  fi
else
  ok "PicoOffice 安装成功！"
  echo
  echo "${C_BLUE}下一步:${C_OFF}"
  if [ "$DO_SYSTEMD" = 1 ]; then
    echo "  服务状态: systemctl status picooffice"
    echo "  查看日志: journalctl -u picooffice -f"
  else
    echo "  手动启动: PICO_PORT=$PORT $EFFECTIVE_INSTALL_DIR/picooffice"
  fi
  echo "  浏览器访问: http://localhost:$PORT"
fi
