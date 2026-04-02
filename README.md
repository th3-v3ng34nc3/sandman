<p align="center">
  <img src="sandman.png" alt="Sandman" width="950"/>
</p>

# <img src="sandman-ico.png" width="30" style="vertical-align:middle"/> Sandman

> **One tool. Every scan. Linux and Windows.**

Sandman is a unified security scanning CLI that wraps [Trivy](https://github.com/aquasecurity/trivy), [Opengrep](https://github.com/opengrep/opengrep), [OWASP ZAP](https://github.com/zaproxy/zaproxy), and [ClamAV](https://www.clamav.net) behind a single, consistent interface — with automatic JSON report generation for every scan.

---

## What Sandman can scan

| Command | What it finds | Engine | Report |
|---|---|---|---|
| `scan image` | CVEs in container images | Trivy | `sandman-image-{ts}.json` |
| `scan secrets` | Hardcoded API keys, tokens, passwords | Trivy | `sandman-secrets-{ts}.json` |
| `scan code` | Security flaws in source code (SAST) | Opengrep | `sandman-code-{ts}.json` |
| `scan iac` | Misconfigurations in Terraform, K8s, Helm, Dockerfiles | Trivy | `sandman-iac-{ts}.json` |
| `scan vuln` | OS and package CVEs on live Linux / Windows systems | Trivy | `sandman-vuln-{ts}.json` |
| `scan malware` | Malware, viruses, and trojans | ClamAV | `sandman-malware-{ts}.log` |
| `scan dast` | Vulnerabilities in running web applications | OWASP ZAP | `sandman-dast-{ts}.json` |
| `scan all` | All of the above in one shot | All engines | Per-scan files, shared `{ts}` |

> Reports are **always saved automatically**. Every scan generates a timestamped file (e.g. `sandman-image-20250401-150405.json`) so nothing is lost. Use `--output` to override the file path.

---

## Output example

Every scan prints a clear header before running and a footer confirming where the report was saved:

```
──────────────────────────────────────────────────────────
🌙 Sandman — Image Scan
   Target   : nginx:latest
   Severity : HIGH,CRITICAL
   Report   : sandman-image-20250401-150405.json
──────────────────────────────────────────────────────────

[trivy output...]

──────────────────────────────────────────────────────────
✅ Sandman image scan complete
   Report saved → sandman-image-20250401-150405.json
──────────────────────────────────────────────────────────
```

`scan all` runs every scan in sequence and prints a combined summary:

```
══════════════════════════════════════════════════════════
🌙 Sandman — Scan Summary
──────────────────────────────────────────────────────────
   image      ✅  sandman-image-20250401-150405.json
   secrets    ✅  sandman-secrets-20250401-150405.json
   code       ✅  sandman-code-20250401-150405.json
   iac        ✅  sandman-iac-20250401-150405.json
   vuln       ✅  sandman-vuln-20250401-150405.json
   malware    ⚠️   sandman-malware-20250401-150405.log
   dast       ✅  sandman-dast-20250401-150405.json
══════════════════════════════════════════════════════════
```

---

## Installation

### Option 1 — Docker (recommended, zero dependencies)

All scan engines are pre-installed in the image.

```bash
docker pull rajvanshi/sandman:latest
# or pin to a specific version
docker pull rajvanshi/sandman:v0.0.1
```

### Option 2 — Pre-built binary

Download the binary for your platform from the [Releases](../../releases) page.

**Linux**
```bash
# amd64
curl -LO https://github.com/th3-v3ng34nc3/sandman/releases/latest/download/sandman_linux_amd64.zip
unzip sandman_linux_amd64.zip && sudo mv sandman /usr/local/bin/

# .deb (Debian / Ubuntu)
curl -LO https://github.com/th3-v3ng34nc3/sandman/releases/latest/download/sandman_linux_amd64.deb
sudo dpkg -i sandman_linux_amd64.deb

# .rpm (RHEL / Fedora)
sudo rpm -i sandman_linux_amd64.rpm
```

**Windows** (PowerShell — run as Administrator)
```powershell
# Automated — installs Sandman + all dependencies
.\windows-installer.ps1

# Manual download
Invoke-WebRequest -Uri "https://github.com/th3-v3ng34nc3/sandman/releases/latest/download/sandman_windows_amd64.zip" -OutFile sandman.zip
Expand-Archive sandman.zip -DestinationPath C:\tools\sandman
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";C:\tools\sandman", "User")
```

### Option 3 — Build from source

Requires Go 1.22+.

```bash
git clone https://github.com/th3-v3ng34nc3/sandman.git
cd sandman
go mod tidy
go build -o sandman .
```

---

## Installing scan engines

> **Skip this section if using Docker** — all engines are bundled in `rajvanshi/sandman`.

Use the included installers to set up all dependencies automatically.

**Linux**
```bash
sudo bash install.sh
```

**Windows** (run as Administrator)
```powershell
.\windows-installer.ps1
```

Both scripts present an interactive menu — select your scan use-case and only the required engines are installed.

### Manual installation

#### Trivy — image, secrets, IaC, and vulnerability scanning

**Linux**
```bash
# Debian / Ubuntu
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | gpg --dearmor | sudo tee /usr/share/keyrings/trivy.gpg > /dev/null
echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb generic main" | sudo tee /etc/apt/sources.list.d/trivy.list
sudo apt-get update && sudo apt-get install -y trivy
```

**Windows**
```powershell
winget install Aquasecurity.Trivy
```

#### ClamAV — malware and virus scanning

**Linux**
```bash
sudo apt-get install -y clamav && sudo freshclam
```

**Windows**
```powershell
winget install ClamAV.ClamAV
freshclam   # update virus definitions
```

#### Opengrep — SAST

**Linux**
```bash
curl -sSL https://github.com/opengrep/opengrep/releases/latest/download/opengrep-linux-amd64 \
  -o /usr/local/bin/opengrep && chmod +x /usr/local/bin/opengrep
```

**Windows**
```powershell
Invoke-WebRequest -Uri "https://github.com/opengrep/opengrep/releases/latest/download/opengrep-windows-amd64.exe" -OutFile C:\tools\sandman\opengrep.exe
```

#### OWASP ZAP — DAST

**Linux**
```bash
sudo snap install zaproxy --classic
```

**Windows**
```powershell
winget install ZAProxy.ZAP
```

---

## Usage

### scan image

```bash
sandman scan image nginx:latest
sandman scan image myapp:1.2.3 --severity CRITICAL
sandman scan image alpine:3.18 --output my-report.json
```

### scan secrets

```bash
sandman scan secrets ./
sandman scan secrets ./src --severity HIGH,CRITICAL,MEDIUM
```

### scan code — SAST

```bash
sandman scan code ./src
sandman scan code ./src --format sarif --output results.sarif
```

### scan iac

Supports Terraform, CloudFormation, Kubernetes manifests, Helm charts, and Dockerfiles.

```bash
sandman scan iac ./infra
sandman scan iac ./k8s --severity CRITICAL
```

### scan vuln — OS / package CVEs

Works on live Linux and Windows filesystems.

```bash
# Linux root
sandman scan vuln /

# Windows system drive
sandman scan vuln C:\

# Specific directory
sandman scan vuln /opt/myapp --severity HIGH,CRITICAL
```

### scan malware

```bash
sandman scan malware /home
sandman scan malware C:\Users --output malware-report.txt
sandman scan malware /uploads
```

> ClamAV exits with code `1` when threats are found and `2` on error. Sandman handles these correctly — findings produce a `⚠️` result in `scan all` rather than a false failure.

### scan dast

```bash
# Passive baseline scan (safe, no active attacks)
sandman scan dast https://staging.myapp.com

# Full active scan — only use against targets you own
sandman scan dast https://staging.myapp.com --full

# API scan with OpenAPI spec
sandman scan dast https://api.myapp.com --api-spec ./openapi.yaml

# Save as JSON
sandman scan dast https://staging.myapp.com --format json --output dast.json
```

### scan all — full pipeline

All reports share a single timestamp so they're easy to correlate.

```bash
# Full pipeline
sandman scan all \
  --image myapp:latest \
  --path ./src \
  --target https://staging.myapp.com

# Filesystem only
sandman scan all --path ./src

# With active DAST
sandman scan all \
  --image myapp:latest \
  --path . \
  --target https://staging.myapp.com \
  --full
```

---

## Flags reference

### Persistent flags (all scan subcommands)

| Flag | Default | Description |
|---|---|---|
| `--severity` | `HIGH,CRITICAL` | Severity filter: `LOW`, `MEDIUM`, `HIGH`, `CRITICAL` |
| `--format` | `json` | Output format: `json`, `table`, `sarif`. DAST also accepts `html`, `xml` |
| `--output` | _(auto-generated)_ | Override the report file path |

### scan dast flags

| Flag | Default | Description |
|---|---|---|
| `--full` | `false` | Full active scan (`zap-full-scan.py`) instead of passive baseline |
| `--api-spec` | — | OpenAPI/Swagger spec path or URL — triggers `zap-api-scan.py` |

### scan all flags

| Flag | Description |
|---|---|
| `--image` | Container image to scan |
| `--path` | Filesystem path — runs secrets, code, IaC, vuln, and malware scans |
| `--target` | Live URL — runs DAST scan |
| `--full` | Use full active ZAP scan instead of baseline |

---

## Running in Docker

```bash
# Container image scan
docker run --rm rajvanshi/sandman:v0.0.1 scan image nginx:latest

# Scan local source code — mount the directory
docker run --rm -v $(pwd):/src rajvanshi/sandman:v0.0.1 scan code /src

# IaC scan
docker run --rm -v $(pwd)/infra:/infra rajvanshi/sandman:v0.0.1 scan iac /infra

# OS vulnerability scan (mount root read-only)
docker run --rm -v /:/host:ro rajvanshi/sandman:v0.0.1 scan vuln /host

# Malware scan with persistent ClamAV definitions
docker run --rm \
  -v /var/lib/clamav:/var/lib/clamav \
  -v $(pwd):/scan \
  rajvanshi/sandman:v0.0.1 scan malware /scan

# DAST
docker run --rm rajvanshi/sandman:v0.0.1 scan dast https://example.com

# Full pipeline — save reports to host
docker run --rm \
  -v $(pwd):/src \
  -v $(pwd)/reports:/reports \
  rajvanshi/sandman:v0.0.1 scan all \
    --image nginx:latest \
    --path /src \
    --target https://example.com
```

---

## Kubernetes

A sample Job manifest is included at [`k8s/job.yaml`](k8s/job.yaml).

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: sandman-scan-nginx
spec:
  template:
    spec:
      containers:
      - name: sandman
        image: rajvanshi/sandman:v0.0.1
        args: ["scan", "image", "nginx:latest"]
      restartPolicy: Never
```

```bash
kubectl apply -f k8s/job.yaml
kubectl logs job/sandman-scan-nginx
```

---

## CI/CD integration

Sandman exits non-zero when any scan finds issues, making it a natural pipeline gate.

### GitHub Actions

```yaml
- name: Sandman security scan
  run: |
    docker run --rm \
      -v ${{ github.workspace }}:/src \
      rajvanshi/sandman:v0.0.1 scan all \
        --image ${{ env.IMAGE }}:${{ github.sha }} \
        --path /src \
        --severity HIGH,CRITICAL
```

### GitLab CI

```yaml
security-scan:
  image: rajvanshi/sandman:v0.0.1
  script:
    - scan image $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - scan code /src --format sarif --output gl-sast-report.json
  artifacts:
    reports:
      sast: gl-sast-report.json
```

---

## How it works

Sandman translates a unified flag set into the correct invocation per engine, streams output to the terminal, saves a JSON report, and exits non-zero when issues are found.

```
sandman scan image nginx:latest
  └─▶  trivy image --severity HIGH,CRITICAL --format json
                   --output sandman-image-20250401-150405.json nginx:latest

sandman scan secrets ./src
  └─▶  trivy fs --scanners secret --severity HIGH,CRITICAL --format json
                --output sandman-secrets-20250401-150405.json ./src

sandman scan code ./src
  └─▶  opengrep scan --config p/default --json
                     --output sandman-code-20250401-150405.json ./src

sandman scan iac ./infra
  └─▶  trivy config --severity HIGH,CRITICAL --format json
                    --output sandman-iac-20250401-150405.json ./infra

sandman scan vuln /
  └─▶  trivy fs --scanners vuln --severity HIGH,CRITICAL --format json
                --output sandman-vuln-20250401-150405.json /

sandman scan malware /app
  └─▶  clamscan --recursive --infected
                --log=sandman-malware-20250401-150405.log /app

sandman scan dast https://example.com
  └─▶  zap-baseline.py -t https://example.com
                       -J sandman-dast-20250401-150405.json
```

---

## Built with

| Tool | Purpose |
|---|---|
| [Trivy](https://github.com/aquasecurity/trivy) | Container image, secrets, IaC, and vulnerability scanning |
| [Opengrep](https://github.com/opengrep/opengrep) | SAST — static source code analysis |
| [OWASP ZAP](https://github.com/zaproxy/zaproxy) | DAST — dynamic application security testing |
| [ClamAV](https://www.clamav.net) | Malware and virus detection |
| [Cobra](https://github.com/spf13/cobra) | CLI framework |
| [GoReleaser](https://goreleaser.com) | Cross-platform builds and packaging |

---

## License

MIT © [th3-v3ng34nc3](https://github.com/th3-v3ng34nc3)
