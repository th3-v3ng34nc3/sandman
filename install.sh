#!/usr/bin/env bash
# Sandman dependency installer — Linux
# https://github.com/th3-v3ng34nc3/sandman
set -euo pipefail

# ── Colours ────────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

ok()     { echo -e "  ${GREEN}✔${NC}  $*"; }
warn()   { echo -e "  ${YELLOW}⚠${NC}  $*"; }
info()   { echo -e "  ${CYAN}→${NC}  $*"; }
err()    { echo -e "  ${RED}✖${NC}  $*"; }
section(){ echo -e "\n${BOLD}${CYAN}── $* ──${NC}"; }

# ── Privilege check ────────────────────────────────────────────────────────────
if [[ $EUID -ne 0 ]]; then
  err "Please run this script as root: sudo bash install.sh"
  exit 1
fi

# ── Distro detection ───────────────────────────────────────────────────────────
if   command -v apt-get &>/dev/null; then PKG_MGR="apt"
elif command -v dnf     &>/dev/null; then PKG_MGR="dnf"
elif command -v yum     &>/dev/null; then PKG_MGR="yum"
else
  err "No supported package manager found (apt / dnf / yum)."
  exit 1
fi

ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

# ── Dependency check ───────────────────────────────────────────────────────────
need() {
  # Returns 1 (needs install) or 0 (already present)
  if command -v "$1" &>/dev/null; then
    ok "$2 already installed — $(command -v "$1")"
    return 0
  fi
  warn "$2 not found"
  return 1
}

# ── Installers ─────────────────────────────────────────────────────────────────
install_trivy() {
  section "Trivy"
  need trivy "Trivy" && return

  info "Installing Trivy…"
  if [[ $PKG_MGR == "apt" ]]; then
    apt-get install -y wget gnupg apt-transport-https &>/dev/null
    wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key \
      | gpg --dearmor -o /usr/share/keyrings/trivy.gpg
    echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] \
https://aquasecurity.github.io/trivy-repo/deb generic main" \
      | tee /etc/apt/sources.list.d/trivy.list > /dev/null
    apt-get update -qq && apt-get install -y trivy
  else
    local url="https://github.com/aquasecurity/trivy/releases/latest/download/trivy_linux_${ARCH}.rpm"
    info "Downloading $url"
    $PKG_MGR install -y "$url"
  fi
  ok "Trivy installed — $(trivy --version 2>&1 | head -1)"
}

install_opengrep() {
  section "Opengrep"
  need opengrep "Opengrep" && return

  local url="https://github.com/opengrep/opengrep/releases/latest/download/opengrep-linux-${ARCH}"
  info "Downloading $url"
  curl -sSL "$url" -o /usr/local/bin/opengrep
  chmod +x /usr/local/bin/opengrep
  ok "Opengrep installed — $(opengrep --version 2>&1 | head -1)"
}

install_clamav() {
  section "ClamAV"
  if need clamscan "ClamAV"; then
    info "Updating virus definitions…"
    freshclam --quiet 2>/dev/null || warn "freshclam update failed — run manually"
    return
  fi

  info "Installing ClamAV…"
  if [[ $PKG_MGR == "apt" ]]; then
    apt-get install -y clamav clamav-freshclam &>/dev/null
    systemctl stop clamav-freshclam 2>/dev/null || true
  else
    $PKG_MGR install -y clamav clamav-update &>/dev/null
  fi

  info "Updating virus definitions (may take a minute)…"
  freshclam --quiet 2>/dev/null || warn "freshclam update failed — run manually before scanning"
  ok "ClamAV installed — $(clamscan --version 2>&1 | head -1)"
}

install_zap() {
  section "OWASP ZAP"
  need zap-baseline.py "ZAP (zap-baseline.py)" && return

  # Ensure Java and Python are present (ZAP requirement)
  info "Installing Java and Python runtime…"
  if [[ $PKG_MGR == "apt" ]]; then
    apt-get install -y default-jre python3 curl &>/dev/null
  else
    $PKG_MGR install -y java-17-openjdk python3 curl &>/dev/null
  fi

  # Prefer snap; fall back to tarball
  if command -v snap &>/dev/null; then
    info "Installing ZAP via snap…"
    snap install zaproxy --classic
    ok "ZAP installed via snap"
    return
  fi

  info "snap not available — downloading ZAP release tarball…"
  local ver
  ver=$(curl -sSL "https://api.github.com/repos/zaproxy/zaproxy/releases/latest" \
        | grep '"tag_name"' | sed 's/.*"v\([^"]*\)".*/\1/')

  curl -sSL \
    "https://github.com/zaproxy/zaproxy/releases/latest/download/ZAP_${ver}_Linux.tar.gz" \
    -o /tmp/zap.tar.gz
  tar -xzf /tmp/zap.tar.gz -C /opt
  local zap_dir
  zap_dir=$(ls -d /opt/ZAP_* 2>/dev/null | sort | tail -1)
  chmod +x "$zap_dir"/zap-*.py
  ln -sf "$zap_dir/zap-baseline.py"  /usr/local/bin/zap-baseline.py
  ln -sf "$zap_dir/zap-full-scan.py" /usr/local/bin/zap-full-scan.py
  ln -sf "$zap_dir/zap-api-scan.py"  /usr/local/bin/zap-api-scan.py
  rm -f /tmp/zap.tar.gz
  ok "ZAP installed — $zap_dir"
}

# ── Menu ───────────────────────────────────────────────────────────────────────
show_menu() {
  echo -e "${BOLD}${CYAN}"
  echo "  ┌──────────────────────────────────────────────────────────┐"
  echo "  │            🌙  Sandman — Dependency Installer            │"
  echo "  │         https://github.com/th3-v3ng34nc3/sandman         │"
  echo "  └──────────────────────────────────────────────────────────┘"
  echo -e "${NC}"
  echo "  Select the scan use-case(s) you need dependencies for:"
  echo ""
  echo "  1)  Container image vulnerabilities     →  trivy"
  echo "  2)  Hardcoded secrets in source code    →  trivy"
  echo "  3)  Source code analysis (SAST)         →  opengrep"
  echo "  4)  Infrastructure as Code (IaC)        →  trivy"
  echo "  5)  OS / package CVE scanning           →  trivy"
  echo "  6)  Malware and virus detection         →  clamav"
  echo "  7)  Live web app scanning (DAST)        →  zap + java + python"
  echo "  8)  All of the above                    →  all engines"
  echo ""
  echo "  0)  Exit"
  echo ""
  printf "  ${BOLD}Your choice [0-8]:${NC} "
  read -r CHOICE
  echo ""
}

# ── Main ───────────────────────────────────────────────────────────────────────
show_menu

NEED_TRIVY=false
NEED_OPENGREP=false
NEED_CLAMAV=false
NEED_ZAP=false

case "$CHOICE" in
  1|2|4|5) NEED_TRIVY=true ;;
  3)        NEED_OPENGREP=true ;;
  6)        NEED_CLAMAV=true ;;
  7)        NEED_ZAP=true ;;
  8)        NEED_TRIVY=true; NEED_OPENGREP=true; NEED_CLAMAV=true; NEED_ZAP=true ;;
  0)        echo "  Bye."; exit 0 ;;
  *)        err "Invalid choice: $CHOICE"; exit 1 ;;
esac

$NEED_TRIVY    && install_trivy
$NEED_OPENGREP && install_opengrep
$NEED_CLAMAV   && install_clamav
$NEED_ZAP      && install_zap

echo ""
echo -e "${BOLD}${GREEN}  ✔  Setup complete. Run 'sandman --help' to get started.${NC}"
echo ""
