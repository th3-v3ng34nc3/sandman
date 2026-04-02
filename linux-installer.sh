#!/usr/bin/env bash
# Sandman installer — Linux
set -euo pipefail

# ── Colours ────────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

ok()   { echo -e "  ${GREEN}✔${NC}  $*"; }
warn() { echo -e "  ${YELLOW}⚠${NC}  $*"; }
info() { echo -e "  ${CYAN}→${NC}  $*"; }
err()  { echo -e "  ${RED}✖${NC}  $*"; }
header() { echo -e "\n${BOLD}${CYAN}$*${NC}\n"; }

# ── Root check ─────────────────────────────────────────────────────────────────
require_root() {
  if [[ $EUID -ne 0 ]]; then
    err "This script must be run as root (use sudo)."
    exit 1
  fi
}

# ── Distro detection ───────────────────────────────────────────────────────────
detect_distro() {
  if command -v apt-get &>/dev/null; then PKG_MGR="apt"
  elif command -v dnf &>/dev/null;     then PKG_MGR="dnf"
  elif command -v yum &>/dev/null;     then PKG_MGR="yum"
  else
    err "Unsupported package manager. Install dependencies manually."
    exit 1
  fi
}

# ── Dependency check ───────────────────────────────────────────────────────────
is_installed() { command -v "$1" &>/dev/null; }

check_dep() {
  local bin=$1 label=$2
  if is_installed "$bin"; then
    ok "$label already installed ($(command -v "$bin"))"
    return 0
  else
    warn "$label not found"
    return 1
  fi
}

require_tool() {
  if ! is_installed "$1"; then
    info "Installing $1..."
    if [[ $PKG_MGR == "apt" ]]; then apt-get install -y "$1" &>/dev/null; else $PKG_MGR install -y "$1" &>/dev/null; fi
  fi
}

# ── Installers ─────────────────────────────────────────────────────────────────
install_trivy() {
  header "Installing Trivy"
  if check_dep trivy "Trivy"; then return; fi

  info "Adding Trivy repository…"
  if [[ $PKG_MGR == "apt" ]]; then
    apt-get install -y wget apt-transport-https gnupg &>/dev/null
    wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key \
      | gpg --dearmor --yes -o /usr/share/keyrings/trivy.gpg
    echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb generic main" \
      > /etc/apt/sources.list.d/trivy.list
    apt-get update -qq && apt-get install -y trivy
  else
    # RPM-based — use official Trivy repository
    info "Adding Trivy repository…"
    cat << 'EOF' > /etc/yum.repos.d/trivy.repo
[trivy]
name=Trivy repository
baseurl=https://aquasecurity.github.io/trivy-repo/rpm/releases/$releasever/$basearch/
gpgcheck=1
enabled=1
gpgkey=https://aquasecurity.github.io/trivy-repo/rpm/public.key
EOF
    $PKG_MGR install -y trivy
  fi

  ok "Trivy installed → $(trivy --version 2>&1 | head -1)"
}

install_opengrep() {
  header "Installing Opengrep"
  if check_dep opengrep "Opengrep"; then return; fi

  require_tool curl
  ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
  URL="https://github.com/opengrep/opengrep/releases/latest/download/opengrep-linux-${ARCH}"
  info "Downloading $URL"
  curl -sSL "$URL" -o /usr/local/bin/opengrep
  chmod +x /usr/local/bin/opengrep

  ok "Opengrep installed → $(opengrep --version 2>&1 | head -1)"
}

install_clamav() {
  header "Installing ClamAV"
  if check_dep clamscan "ClamAV"; then
    info "Updating virus definitions…"
    freshclam --quiet || true
    return
  fi

  if [[ $PKG_MGR == "apt" ]]; then
    apt-get install -y clamav clamav-freshclam &>/dev/null
    # Stop the daemon so freshclam can run standalone
    systemctl stop clamav-freshclam 2>/dev/null || true
  else
    $PKG_MGR install -y clamav clamav-update &>/dev/null
  fi

  info "Updating virus definitions (this may take a minute)…"
  chown -R clamav:clamav /var/lib/clamav 2>/dev/null || true
  freshclam --quiet || warn "freshclam update failed — run manually before scanning"

  ok "ClamAV installed → $(clamscan --version 2>&1 | head -1)"
}

install_zap() {
  header "Installing OWASP ZAP"
  if check_dep zap-baseline.py "ZAP (zap-baseline.py)"; then return; fi

  require_tool curl

  # Try snap first
  if command -v snap &>/dev/null; then
    info "Installing via snap…"
    snap install zaproxy --classic
    ok "ZAP installed via snap"
    return
  fi

  # Fallback — download the ZAP release tarball
  info "snap not available — downloading ZAP release…"
  ZAP_VERSION=$(curl -sSIL -o /dev/null -w "%{url_effective}" https://github.com/zaproxy/zaproxy/releases/latest | grep -oE "[0-9]+\.[0-9]+\.[0-9]+")

  curl -sSL "https://github.com/zaproxy/zaproxy/releases/latest/download/ZAP_${ZAP_VERSION}_Linux.tar.gz" \
    -o /tmp/zap.tar.gz
  tar -xzf /tmp/zap.tar.gz -C /opt
  ZAP_DIR="/opt/ZAP_${ZAP_VERSION}"
  ln -sf "$ZAP_DIR/zap-baseline.py"  /usr/local/bin/zap-baseline.py
  ln -sf "$ZAP_DIR/zap-full-scan.py" /usr/local/bin/zap-full-scan.py
  ln -sf "$ZAP_DIR/zap-api-scan.py"  /usr/local/bin/zap-api-scan.py
  chmod +x "$ZAP_DIR"/zap-*.py
  rm -f /tmp/zap.tar.gz

  # ZAP scan scripts need Python + java
  info "Installing Java and Python (ZAP runtime requirements)…"
  if [[ $PKG_MGR == "apt" ]]; then
    apt-get install -y default-jre python3 &>/dev/null
  else
    $PKG_MGR install -y java-17-openjdk python3 &>/dev/null
  fi

  ok "ZAP installed → $ZAP_DIR"
}

install_sandman() {
  header "Installing Sandman CLI"
  if check_dep sandman "Sandman"; then return; fi

  require_tool curl
  require_tool unzip
  ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
  URL="https://github.com/th3-v3ng34nc3/sandman/releases/download/v0.0.1/sandman_linux_${ARCH}.zip"
  info "Downloading $URL"
  curl -sSL "$URL" -o /tmp/sandman.zip
  unzip -q /tmp/sandman.zip -d /tmp/sandman_ext
  mv /tmp/sandman_ext/sandman /usr/local/bin/sandman
  chmod +x /usr/local/bin/sandman
  rm -rf /tmp/sandman.zip /tmp/sandman_ext

  ok "Sandman installed → $(sandman version 2>&1 | head -1)"
}

# ── Menu ───────────────────────────────────────────────────────────────────────
show_menu() {
  echo -e "${BOLD}${CYAN}"
  echo "  ┌────────────────────────────────────────────────────┐"
  echo "  │           🌙  Sandman — Dependency Setup           │"
  echo "  └────────────────────────────────────────────────────┘"
  echo -e "${NC}"
  echo "  What do you want to scan?"
  echo ""
  echo "  1) Container images                →  trivy"
  echo "  2) Source code for secrets         →  trivy"
  echo "  3) Source code (SAST)              →  opengrep"
  echo "  4) Infrastructure as Code (IaC)    →  trivy"
  echo "  5) OS / package vulnerabilities    →  trivy"
  echo "  6) Files for malware / viruses     →  clamav"
  echo "  7) Live web applications (DAST)    →  zap + python + java"
  echo "  8) All underlying scan engines     →  all engines"
  echo "  9) Sandman CLI                     →  sandman binary"
  echo "  10) Everything                     →  sandman + all engines"
  echo ""
  echo "  0) Exit"
  echo ""
  printf "  ${BOLD}Your choice [0-10]:${NC} "
  read -r choice
  echo ""
}

# ── Main ───────────────────────────────────────────────────────────────────────
main() {
  require_root
  detect_distro
  show_menu

  case "$choice" in
    1) install_trivy ;;
    2) install_trivy ;;
    3) install_opengrep ;;
    4) install_trivy ;;
    5) install_trivy ;;
    6) install_clamav ;;
    7) install_zap ;;
    8)
      install_trivy
      install_opengrep
      install_clamav
      install_zap
      ;;
    9) install_sandman ;;
    10)
      install_sandman
      install_trivy
      install_opengrep
      install_clamav
      install_zap
      ;;
    0) echo "  Bye."; exit 0 ;;
    *)
      err "Invalid choice: $choice"
      exit 1
      ;;
  esac

  echo ""
  echo -e "${BOLD}${GREEN}  ✔  Setup complete. Run 'sandman --help' to get started.${NC}"
  echo ""
}

main "$@"
