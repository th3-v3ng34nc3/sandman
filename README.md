# 🌙 Sandman

**Sandman** is a unified security scanning CLI that wraps [Trivy](https://github.com/aquasecurity/trivy), [Opengrep](https://github.com/opengrep/opengrep), and [OWASP ZAP](https://github.com/zaproxy/zaproxy) into a single, consistent interface.

Run container, code, secrets, IaC, and DAST scans — all with one tool.

---

## Features

| Scan Type | Command | Engine |
|---|---|---|
| Container image vulnerabilities | `scan image` | Trivy |
| Hardcoded secrets | `scan secrets` | Trivy |
| Source code (SAST) | `scan code` | Opengrep |
| Infrastructure as Code | `scan iac` | Trivy |
| Dynamic app security (DAST) | `scan dast` | OWASP ZAP |
| All of the above | `scan all` | All engines |

---

## Installation

### Docker (recommended)

```bash
docker pull rajvanshi/sandman:latest
```

### Build from source

Requires Go 1.22+.

```bash
git clone https://github.com/rajvanshi/sandman.git
cd sandman
go mod tidy
go build -o sandman .
```

### Releases

Pre-built binaries for Linux and Windows (amd64 / arm64) are available on the [Releases](../../releases) page as `.zip`, `.deb`, and `.rpm` packages.

---

## Usage

### Scan a container image

```bash
sandman scan image nginx:latest
sandman scan image nginx:latest --severity CRITICAL
sandman scan image nginx:latest --format json --output report.json
```

### Scan source code (SAST)

```bash
sandman scan code ./src
sandman scan code ./src --format sarif --output results.sarif
```

### Scan for secrets

```bash
sandman scan secrets ./
sandman scan secrets ./ --severity HIGH,CRITICAL,MEDIUM
```

### Scan Infrastructure as Code

Supports Terraform, CloudFormation, Kubernetes manifests, Helm charts, and Dockerfiles.

```bash
sandman scan iac ./infra
sandman scan iac ./k8s --format json --output iac-report.json
```

### Dynamic Application Security Testing (DAST)

```bash
# Passive baseline scan (default)
sandman scan dast https://example.com

# Full active scan
sandman scan dast https://example.com --full

# API scan with OpenAPI spec
sandman scan dast https://api.example.com --api-spec ./openapi.yaml

# Save report
sandman scan dast https://example.com --output report.html
sandman scan dast https://example.com --format json --output report.json
```

### Run all scans at once

```bash
sandman scan all --image nginx:latest --path ./src --target https://example.com

# With active DAST
sandman scan all --image myapp:latest --path . --target https://staging.myapp.com --full
```

---

## Flags

### Global scan flags

All `scan` subcommands inherit these flags:

| Flag | Default | Description |
|---|---|---|
| `--severity` | `HIGH,CRITICAL` | Comma-separated severity levels (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`) |
| `--format` | `table` | Output format: `table`, `json`, `sarif` (DAST also supports `html`, `xml`) |
| `--output` | _(stdout)_ | Write results to a file |

### `scan dast` flags

| Flag | Description |
|---|---|
| `--full` | Run a full active scan instead of the passive baseline |
| `--api-spec` | Path or URL to an OpenAPI/Swagger spec for API scanning |

### `scan all` flags

| Flag | Description |
|---|---|
| `--image` | Container image to scan |
| `--path` | Filesystem path for secrets, code, and IaC scans |
| `--target` | Live URL for DAST scan |
| `--full` | Use full active ZAP scan instead of baseline |

---

## Running in Docker

All scan engines are bundled in the Docker image. Mount your source code or pass image names directly.

```bash
# Scan a container image
docker run --rm rajvanshi/sandman scan image nginx:latest

# Scan local source code
docker run --rm -v $(pwd):/src rajvanshi/sandman scan code /src

# Scan IaC configs
docker run --rm -v $(pwd)/infra:/infra rajvanshi/sandman scan iac /infra

# Run DAST against a live target
docker run --rm rajvanshi/sandman scan dast https://example.com

# Run all scans
docker run --rm -v $(pwd):/src rajvanshi/sandman scan all \
  --image nginx:latest \
  --path /src \
  --target https://example.com
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

## How it works

```
sandman scan image nginx:latest
        │
        ▼
  trivy image --severity HIGH,CRITICAL nginx:latest

sandman scan code ./src
        │
        ▼
  opengrep scan --config p/default ./src

sandman scan dast https://example.com
        │
        ▼
  zap-baseline.py -t https://example.com
```

Sandman checks that each required binary is in `PATH` before running. If a tool is missing, it exits with a clear error rather than a cryptic failure.

---

## Built with

- [Trivy](https://github.com/aquasecurity/trivy) — vulnerability, secrets, and IaC scanning
- [Opengrep](https://github.com/opengrep/opengrep) — SAST / static analysis
- [OWASP ZAP](https://github.com/zaproxy/zaproxy) — dynamic application security testing
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [GoReleaser](https://goreleaser.com) — cross-platform builds and packaging

---

## License

MIT
