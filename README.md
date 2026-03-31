# 🌙 Sandman

> **One tool. Every scan. Linux and Windows.**

Sandman is a unified security scanning CLI that wraps best-in-class open-source engines — [Trivy](https://github.com/aquasecurity/trivy), [Opengrep](https://github.com/opengrep/opengrep), [OWASP ZAP](https://github.com/zaproxy/zaproxy), and [ClamAV](https://www.clamav.net) — behind a single, consistent interface.

Instead of remembering the flags and quirks of four different tools, you run one command. Sandman selects the right engine, builds the correct arguments, streams output directly to your terminal, and exits non-zero when issues are found — making it trivially easy to drop into any CI/CD pipeline.

---

## What Sandman can scan

| Command | What it finds | Engine |
|---|---|---|
| `scan image` | CVEs in container images (OS packages, app deps) | Trivy |
| `scan secrets` | Hardcoded API keys, tokens, passwords in source code | Trivy |
| `scan code` | Security flaws in source code (SAST) | Opengrep |
| `scan iac` | Misconfigurations in Terraform, K8s, Helm, Dockerfiles | Trivy |
| `scan vuln` | OS and package CVEs on live Linux / Windows systems | Trivy |
| `scan malware` | Malware, viruses, and trojans in files and directories | ClamAV |
| `scan dast` | Vulnerabilities in running web applications (DAST) | OWASP ZAP |
| `scan all` | All of the above in one shot | All engines |

---

## Installation

### Option 1 — Docker (recommended, zero dependencies)

The Docker image ships with all scan engines pre-installed. No local setup required.

```bash
docker pull rajvanshi/sandman:latest
```

### Option 2 — Pre-built binary

Download the binary for your platform from the [Releases](../../releases) page.

**Linux**
```bash
# amd64
curl -LO https://github.com/rajvanshi/sandman/releases/latest/download/sandman_linux_amd64.zip
unzip sandman_linux_amd64.zip && sudo mv sandman /usr/local/bin/

# .deb (Debian / Ubuntu)
curl -LO https://github.com/rajvanshi/sandman/releases/latest/download/sandman_linux_amd64.deb
sudo dpkg -i sandman_linux_amd64.deb

# .rpm (RHEL / Fedora)
sudo rpm -i sandman_linux_amd64.rpm
```

**Windows** (PowerShell)
```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/rajvanshi/sandman/releases/latest/download/sandman_windows_amd64.zip" -OutFile sandman.zip
Expand-Archive sandman.zip -DestinationPath C:\tools\sandman

# Add to PATH
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";C:\tools\sandman", "User")
```

### Option 3 — Build from source

Requires Go 1.22+.

```bash
git clone https://github.com/rajvanshi/sandman.git
cd sandman
go mod tidy
go build -o sandman .
```

---

## Installing scan engines

Sandman is a wrapper — the underlying scan engines must be in your `PATH`. If a required engine is missing, Sandman will tell you clearly before attempting the scan.

> **Skip this section if you are using Docker** — all engines are pre-installed in `rajvanshi/sandman:latest`.

### Trivy — container, secrets, IaC, and vulnerability scanning

Used by: `scan image`, `scan secrets`, `scan iac`, `scan vuln`

**Linux**
```bash
# Debian / Ubuntu
sudo apt-get install wget apt-transport-https gnupg
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
echo "deb https://aquasecurity.github.io/trivy-repo/deb generic main" | sudo tee /etc/apt/sources.list.d/trivy.list
sudo apt-get update && sudo apt-get install trivy

# RPM-based (RHEL / Fedora / CentOS)
sudo rpm -ivh https://github.com/aquasecurity/trivy/releases/latest/download/trivy_linux_amd64.rpm
```

**Windows**
```powershell
# Scoop
scoop install trivy

# Winget
winget install aquasecurity.trivy

# Manual: download trivy.exe from https://github.com/aquasecurity/trivy/releases
```

### ClamAV — malware and virus scanning

Used by: `scan malware`

**Linux**
```bash
# Debian / Ubuntu
sudo apt-get install clamav
sudo freshclam          # update virus definitions

# RHEL / Fedora
sudo dnf install clamav clamav-update
sudo freshclam
```

**Windows**

1. Download the official ClamAV installer from [clamav.net/downloads](https://www.clamav.net/downloads)
2. Run the installer — this installs `clamscan.exe` and adds it to `PATH`
3. Update virus definitions:
```cmd
freshclam
```

Or via Winget:
```powershell
winget install ClamAV.ClamAV
freshclam
```

> **Keep definitions fresh.** ClamAV is only effective with up-to-date virus signatures. Run `freshclam` regularly or set up a scheduled task / cron job.

### Opengrep — SAST / source code analysis

Used by: `scan code`

**Linux**
```bash
# Download binary from GitHub releases
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -sSL "https://github.com/opengrep/opengrep/releases/latest/download/opengrep-linux-${ARCH}" \
  -o /usr/local/bin/opengrep && chmod +x /usr/local/bin/opengrep
```

**Windows**
```powershell
# Download opengrep.exe from GitHub releases
Invoke-WebRequest -Uri "https://github.com/opengrep/opengrep/releases/latest/download/opengrep-windows-amd64.exe" -OutFile C:\tools\sandman\opengrep.exe
```

### OWASP ZAP — dynamic application security testing

Used by: `scan dast`

**Linux**
```bash
# Via snap
sudo snap install zaproxy --classic

# Or download from https://github.com/zaproxy/zaproxy/releases
# Ensure zap-baseline.py and zap-full-scan.py are in PATH
```

**Windows**

1. Download the Windows installer from [zaproxy.org](https://www.zaproxy.org/download/)
2. Run the installer
3. Add the ZAP installation directory to `PATH` so `zap-baseline.py` is accessible
4. Ensure Python 3 is installed (ZAP scan scripts require it)

> **DAST via Docker** is the easiest cross-platform approach — the Docker image handles ZAP automatically.

---

## Usage

### scan image — container vulnerability scan

Scans a container image for known CVEs in OS packages and application dependencies.

```bash
sandman scan image nginx:latest
sandman scan image myapp:1.2.3 --severity CRITICAL
sandman scan image alpine:3.18 --format json --output image-report.json
sandman scan image ubuntu:22.04 --format sarif --output image.sarif
```

### scan secrets — hardcoded secret detection

Scans source code and configuration files for accidentally committed secrets: API keys, tokens, passwords, certificates, and more.

```bash
sandman scan secrets .
sandman scan secrets ./src --severity HIGH,CRITICAL
sandman scan secrets /var/app --output secrets.json --format json
```

### scan code — static application security testing (SAST)

Analyses source code for security vulnerabilities including SQL injection, XSS, path traversal, insecure deserialization, and hundreds of other patterns via the Opengrep `p/default` ruleset.

```bash
sandman scan code ./src
sandman scan code ./src --format sarif --output sast.sarif
sandman scan code /app --format json --output sast.json
```

### scan iac — Infrastructure as Code misconfiguration

Scans Terraform plans, CloudFormation templates, Kubernetes manifests, Helm charts, and Dockerfiles for security misconfigurations such as open security groups, missing encryption, over-privileged roles, and container policy violations.

```bash
sandman scan iac ./infra
sandman scan iac ./k8s --severity CRITICAL
sandman scan iac ./terraform --format json --output iac.json
```

### scan vuln — OS and package vulnerability scan

Scans a live filesystem for vulnerabilities in installed OS packages and language runtime libraries. Works on both Linux and Windows paths.

```bash
# Linux — scan the root filesystem
sandman scan vuln /

# Windows — scan the system drive
sandman scan vuln C:\

# Scan a specific app directory
sandman scan vuln /opt/myapp --severity HIGH,CRITICAL
sandman scan vuln C:\inetpub\wwwroot --format json --output vuln.json
```

### scan malware — malware and virus detection

Recursively scans a directory for malware, viruses, worms, and trojans using ClamAV. By default only infected files are printed to keep output clean.

```bash
sandman scan malware /home
sandman scan malware C:\Users --output malware-report.txt
sandman scan malware /uploads
```

### scan dast — dynamic application security testing

Runs OWASP ZAP against a live web target to find vulnerabilities that only appear at runtime: reflected XSS, CSRF, broken authentication, insecure redirects, and more.

```bash
# Passive baseline scan (safe, no active attacks)
sandman scan dast https://staging.myapp.com

# Full active scan (sends attack payloads — use only against targets you own)
sandman scan dast https://staging.myapp.com --full

# API scan with OpenAPI / Swagger spec
sandman scan dast https://api.myapp.com --api-spec ./openapi.yaml

# Save report
sandman scan dast https://staging.myapp.com --output dast.html
sandman scan dast https://staging.myapp.com --format json --output dast.json
```

> **Warning:** The `--full` flag performs active scanning — it sends real attack payloads. Only use it against systems you own or have explicit permission to test.

### scan all — run every scan at once

Runs all applicable scan types in sequence. Continues through individual failures and prints a combined summary at the end. Exits non-zero if any scan found issues.

```bash
# Full pipeline: image + filesystem + DAST
sandman scan all \
  --image myapp:latest \
  --path ./src \
  --target https://staging.myapp.com

# Just filesystem scans (secrets, code, IaC, vuln, malware)
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
| `--severity` | `HIGH,CRITICAL` | Comma-separated severity filter: `LOW`, `MEDIUM`, `HIGH`, `CRITICAL` |
| `--format` | `table` | Output format: `table`, `json`, `sarif`. DAST also accepts `html` and `xml` |
| `--output` | _(stdout)_ | Write results to a file instead of stdout |

### scan dast flags

| Flag | Default | Description |
|---|---|---|
| `--full` | `false` | Run a full active scan (`zap-full-scan.py`) instead of the passive baseline |
| `--api-spec` | — | Path or URL to an OpenAPI/Swagger spec — triggers `zap-api-scan.py` |

### scan all flags

| Flag | Description |
|---|---|
| `--image` | Container image to scan (e.g. `nginx:latest`) |
| `--path` | Filesystem path — runs secrets, code, IaC, vuln, and malware scans |
| `--target` | Live URL — runs DAST baseline scan |
| `--full` | Use full active ZAP scan instead of baseline |

---

## Running in Docker

All engines are bundled. Mount your code or pass targets directly.

```bash
# Container image scan
docker run --rm rajvanshi/sandman scan image nginx:latest

# Scan local source code
docker run --rm -v $(pwd):/src rajvanshi/sandman scan code /src

# Scan IaC configs
docker run --rm -v $(pwd)/infra:/infra rajvanshi/sandman scan iac /infra

# Scan for OS vulnerabilities (mount root)
docker run --rm -v /:/host:ro rajvanshi/sandman scan vuln /host

# Malware scan with fresh definitions (mount ClamAV DB)
docker run --rm \
  -v /var/lib/clamav:/var/lib/clamav \
  -v $(pwd):/scan \
  rajvanshi/sandman scan malware /scan

# DAST against a live target
docker run --rm rajvanshi/sandman scan dast https://staging.myapp.com

# Full pipeline
docker run --rm \
  -v $(pwd):/src \
  rajvanshi/sandman scan all \
    --image nginx:latest \
    --path /src \
    --target https://staging.myapp.com
```

---

## Kubernetes

A sample Job manifest is included in [`k8s/job.yaml`](k8s/job.yaml).

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
        image: rajvanshi/sandman:latest
        args: ["scan", "image", "nginx:latest"]
      restartPolicy: Never
```

```bash
kubectl apply -f k8s/job.yaml
kubectl logs job/sandman-scan-nginx
```

---

## CI/CD integration

Sandman exits non-zero when any scan finds issues, making it a natural gate in CI pipelines.

### GitHub Actions

```yaml
- name: Security scan
  run: |
    docker run --rm \
      -v ${{ github.workspace }}:/src \
      rajvanshi/sandman scan all \
        --image ${{ env.IMAGE }}:${{ github.sha }} \
        --path /src \
        --severity HIGH,CRITICAL
```

### GitLab CI

```yaml
security-scan:
  image: rajvanshi/sandman:latest
  script:
    - scan image $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - scan code /src --format sarif --output gl-sast-report.json
  artifacts:
    reports:
      sast: gl-sast-report.json
```

---

## How it works

Sandman is a thin, opinionated wrapper. It translates a unified flag set into the correct invocation for each underlying tool, checks the binary exists before running, streams stdout/stderr directly to the terminal, and maps tool exit codes to a consistent pass/fail signal.

```
sandman scan image nginx:latest --severity CRITICAL
         └─▶  trivy image --severity CRITICAL nginx:latest

sandman scan secrets ./src --format json --output out.json
         └─▶  trivy fs --scanners secret --severity HIGH,CRITICAL
                        --format json --output out.json ./src

sandman scan code ./src
         └─▶  opengrep scan --config p/default ./src

sandman scan iac ./infra
         └─▶  trivy config --severity HIGH,CRITICAL ./infra

sandman scan vuln /
         └─▶  trivy fs --scanners vuln --severity HIGH,CRITICAL /

sandman scan malware /app
         └─▶  clamscan --recursive --infected /app

sandman scan dast https://example.com
         └─▶  zap-baseline.py -t https://example.com

sandman scan dast https://example.com --full
         └─▶  zap-full-scan.py -t https://example.com
```

---

## Built with

| Tool | Purpose |
|---|---|
| [Trivy](https://github.com/aquasecurity/trivy) | Container image, filesystem, secrets, IaC, and vulnerability scanning |
| [Opengrep](https://github.com/opengrep/opengrep) | SAST — static source code analysis |
| [OWASP ZAP](https://github.com/zaproxy/zaproxy) | DAST — dynamic application security testing |
| [ClamAV](https://www.clamav.net) | Malware and virus detection |
| [Cobra](https://github.com/spf13/cobra) | CLI framework |
| [GoReleaser](https://goreleaser.com) | Cross-platform builds and packaging |

---

## License

MIT
