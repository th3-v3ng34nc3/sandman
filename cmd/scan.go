package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var (
	severity string
	format   string
	output   string
)

const divider = "──────────────────────────────────────────────────────────"

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Security scanning operations",
}

// ------------------------------------------------------------
// Verbose helpers
// ------------------------------------------------------------

func printScanHeader(scanType, target, reportPath string) {
	fmt.Println()
	fmt.Println(divider)
	fmt.Printf("🌙 Sandman — %s\n", scanType)
	fmt.Printf("   Target   : %s\n", target)
	fmt.Printf("   Severity : %s\n", severity)
	fmt.Printf("   Report   : %s\n", reportPath)
	fmt.Println(divider)
	fmt.Println()
}

func printScanDone(scanType, reportPath string) {
	fmt.Println()
	fmt.Println(divider)
	fmt.Printf("✅ Sandman %s complete\n", scanType)
	fmt.Printf("   Report saved → %s\n", reportPath)
	fmt.Println(divider)
}

func printScanFailed(scanType string, err error) {
	fmt.Println()
	fmt.Println(divider)
	fmt.Printf("❌ Sandman %s failed: %v\n", scanType, err)
	fmt.Println(divider)
}

// resolveReport returns the output path for a scan.
// If --output is set, it uses that. Otherwise it auto-generates a timestamped filename.
func resolveReport(scanType, ts string) string {
	if output != "" {
		return output
	}
	return fmt.Sprintf("sandman-%s-%s.json", scanType, ts)
}

// resolveMalwareReport returns the log path for a ClamAV scan.
// ClamAV does not support JSON output natively, so the report is a .log file.
func resolveMalwareReport(ts string) string {
	if output != "" {
		return output
	}
	return fmt.Sprintf("sandman-malware-%s.log", ts)
}

// ------------------------------------------------------------
// scan image
// ------------------------------------------------------------

var scanImageCmd = &cobra.Command{
	Use:   "image [target]",
	Short: "Scan a container image for vulnerabilities",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		report := resolveReport("image", time.Now().Format("20060102-150405"))
		printScanHeader("Image Scan", args[0], report)
		if err := runTool("trivy", buildTrivyImageArgs(args[0], report)); err != nil {
			printScanFailed("image scan", err)
			os.Exit(1)
		}
		printScanDone("image scan", report)
	},
}

// ------------------------------------------------------------
// scan secrets
// ------------------------------------------------------------

var scanSecretsCmd = &cobra.Command{
	Use:   "secrets [path]",
	Short: "Scan a directory for hardcoded secrets and tokens",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		report := resolveReport("secrets", time.Now().Format("20060102-150405"))
		printScanHeader("Secrets Scan", args[0], report)
		if err := runTool("trivy", buildTrivySecretsArgs(args[0], report)); err != nil {
			printScanFailed("secrets scan", err)
			os.Exit(1)
		}
		printScanDone("secrets scan", report)
	},
}

// ------------------------------------------------------------
// scan code
// ------------------------------------------------------------

var scanCodeCmd = &cobra.Command{
	Use:   "code [path]",
	Short: "Scan source code for security flaws",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		report := resolveReport("code", time.Now().Format("20060102-150405"))
		printScanHeader("Code Scan (SAST)", args[0], report)
		if err := runTool("opengrep", buildOpengrepArgs(args[0], report)); err != nil {
			printScanFailed("code scan", err)
			os.Exit(1)
		}
		printScanDone("code scan", report)
	},
}

// ------------------------------------------------------------
// scan iac
// ------------------------------------------------------------

var scanIaCCmd = &cobra.Command{
	Use:   "iac [path]",
	Short: "Scan Infrastructure as Code for misconfigurations",
	Long:  "Scans Terraform, CloudFormation, Kubernetes manifests, Helm charts, and Dockerfiles for security misconfigurations using Trivy.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		report := resolveReport("iac", time.Now().Format("20060102-150405"))
		printScanHeader("IaC Scan", args[0], report)
		if err := runTool("trivy", buildTrivyIaCArgs(args[0], report)); err != nil {
			printScanFailed("IaC scan", err)
			os.Exit(1)
		}
		printScanDone("IaC scan", report)
	},
}

// ------------------------------------------------------------
// scan vuln
// ------------------------------------------------------------

var scanVulnCmd = &cobra.Command{
	Use:   "vuln [path]",
	Short: "Scan a filesystem for OS and package vulnerabilities",
	Long:  "Scans installed OS packages and language runtime dependencies for known CVEs on Linux and Windows using Trivy.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		report := resolveReport("vuln", time.Now().Format("20060102-150405"))
		printScanHeader("Vulnerability Scan", args[0], report)
		if err := runTool("trivy", buildTrivyVulnArgs(args[0], report)); err != nil {
			printScanFailed("vulnerability scan", err)
			os.Exit(1)
		}
		printScanDone("vulnerability scan", report)
	},
}

// ------------------------------------------------------------
// scan malware
// ------------------------------------------------------------

var scanMalwareCmd = &cobra.Command{
	Use:   "malware [path]",
	Short: "Scan a directory for malware using ClamAV",
	Long: `Recursively scans a directory for malware, viruses, and trojans using ClamAV.

ClamAV exit codes:
  0  No threats found
  1  Threats detected (scan completed — review the report)
  2  An error occurred

Note: ClamAV does not support JSON output natively. The report is saved as a .log file.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		report := resolveMalwareReport(time.Now().Format("20060102-150405"))
		printScanHeader("Malware Scan", args[0], report)
		found, err := runMalwareScan(buildClamAVArgs(args[0], report))
		if err != nil {
			printScanFailed("malware scan", err)
			os.Exit(2)
		}
		if found {
			fmt.Println()
			fmt.Println(divider)
			fmt.Println("⚠️  Sandman malware scan complete — threats detected!")
			fmt.Printf("   Report saved → %s\n", report)
			fmt.Println(divider)
			os.Exit(1)
		}
		printScanDone("malware scan", report)
	},
}

// ------------------------------------------------------------  
// scan sbom  
// ------------------------------------------------------------  
  
var scanSbomCmd = &cobra.Command{  
	Use:   "sbom [target]",  
	Short: "Generate a Software Bill of Materials (SBOM)",  
	Long: `Generate an SBOM for a container image or filesystem using Trivy.  
  
Supported SBOM formats (--sbom-format):  
  cyclonedx      CycloneDX JSON (default)  
  spdx-json      SPDX JSON  
  spdx           SPDX tag-value  
  
Target type (--type):  
  image          Container image (default)  
  fs             Filesystem path`,  
	Args: cobra.ExactArgs(1),  
	Run: func(cmd *cobra.Command, args []string) {  
		sbomFormat, _ := cmd.Flags().GetString("sbom-format")  
		targetType, _ := cmd.Flags().GetString("type")  
		report := resolveReport("sbom", time.Now().Format("20060102-150405"))  
		printScanHeader("SBOM Generation", args[0], report)  
		if err := runTool("trivy", buildTrivySbomArgs(args[0], sbomFormat, targetType, report)); err != nil {  
			printScanFailed("SBOM generation", err)  
			os.Exit(1)  
		}  
		printScanDone("SBOM generation", report)  
	},  
}

// ------------------------------------------------------------  
// scan license  
// ------------------------------------------------------------  
  
var scanLicenseCmd = &cobra.Command{  
	Use:   "license [path]",  
	Short: "Scan for license compliance issues in dependencies",  
	Long:  "Scans a filesystem path for software license compliance issues using Trivy. Detects restrictive or non-compliant licenses in project dependencies.",  
	Args:  cobra.ExactArgs(1),  
	Run: func(cmd *cobra.Command, args []string) {  
		report := resolveReport("license", time.Now().Format("20060102-150405"))  
		printScanHeader("License Scan", args[0], report)  
		if err := runTool("trivy", buildTrivyLicenseArgs(args[0], report)); err != nil {  
			printScanFailed("license scan", err)  
			os.Exit(1)  
		}  
		printScanDone("license scan", report)  
	},  
}

// ------------------------------------------------------------
// scan dast
// ------------------------------------------------------------

var scanDastCmd = &cobra.Command{
	Use:   "dast [target-url]",
	Short: "Run dynamic application security testing against a live target",
	Long: `Runs OWASP ZAP against a live target URL.

Scan modes:
  default        Passive baseline scan   (zap-baseline.py)
  --full         Active full scan        (zap-full-scan.py)
  --api-spec     OpenAPI/Swagger scan    (zap-api-scan.py)

Output format for --output:
  json           -J flag  (JSON report)
  html           -r flag  (HTML report, default when --output is set)
  xml            -x flag  (XML report)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		full, _ := cmd.Flags().GetBool("full")
		apiSpec, _ := cmd.Flags().GetString("api-spec")
		report := resolveReport("dast", time.Now().Format("20060102-150405"))
		binary, zapArgs := buildZapArgs(args[0], full, apiSpec, report)
		printScanHeader("DAST Scan", args[0], report)
		if err := runTool(binary, zapArgs); err != nil {
			printScanFailed("DAST scan", err)
			os.Exit(1)
		}
		printScanDone("DAST scan", report)
	},
}

// ------------------------------------------------------------
// scan all
// ------------------------------------------------------------

type scanResult struct {
	label  string
	status string // "pass", "findings", "failed"
	report string
}

var scanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all scans in sequence",
	Long: `Run every scan type in sequence. Reports are auto-saved per scan type.

Flags:
  --image     Container image   → image scan
  --path      Filesystem path   → secrets, code, IaC, vuln, malware, SBOM, and license scans
  --target    Live URL          → DAST scan (add --full for active)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		image, _ := cmd.Flags().GetString("image")
		path, _ := cmd.Flags().GetString("path")
		target, _ := cmd.Flags().GetString("target")
		full, _ := cmd.Flags().GetBool("full")

		if image == "" && path == "" && target == "" {
			return fmt.Errorf("provide at least one of --image, --path, or --target")
		}

		// Shared timestamp so all reports from this run share the same suffix
		ts := time.Now().Format("20060102-150405")
		var results []scanResult

		fmt.Println()
		fmt.Println("══════════════════════════════════════════════════════════")
		fmt.Printf("🌙 Sandman — Full Security Scan  [%s]\n", ts)
		if image != "" {
			fmt.Printf("   Image    : %s\n", image)
		}
		if path != "" {
			fmt.Printf("   Path     : %s\n", path)
		}
		if target != "" {
			fmt.Printf("   Target   : %s\n", target)
		}
		fmt.Println("══════════════════════════════════════════════════════════")

		// ── Image ────────────────────────────────────────────────────
		if image != "" {
			r := fmt.Sprintf("sandman-image-%s.json", ts)
			printScanHeader("Image Scan", image, r)
			if err := runTool("trivy", buildTrivyImageArgs(image, r)); err != nil {
				printScanFailed("image scan", err)
				results = append(results, scanResult{"image", "failed", r})
			} else {
				printScanDone("image scan", r)
				results = append(results, scanResult{"image", "pass", r})
			}
			rImageSbom := fmt.Sprintf("sandman-sbom-image-%s.json", ts)  
			printScanHeader("Image SBOM Generation", image, rImageSbom)  
			if err := runTool("trivy", buildTrivySbomArgs(image, "cyclonedx", "image", rImageSbom)); err != nil {  
				printScanFailed("image SBOM generation", err)  
				results = append(results, scanResult{"sbom(image)", "failed", rImageSbom})  
			} else {  
				printScanDone("image SBOM generation", rImageSbom)  
				results = append(results, scanResult{"sbom(image)", "pass", rImageSbom})  
			}
		}

		// ── Filesystem scans ─────────────────────────────────────────
		if path != "" {
			type fsJob struct {
				label   string
				scanKey string
				args    []string
				isMal   bool
			}

			rSecrets := fmt.Sprintf("sandman-secrets-%s.json", ts)
			rCode := fmt.Sprintf("sandman-code-%s.json", ts)
			rIaC := fmt.Sprintf("sandman-iac-%s.json", ts)
			rVuln := fmt.Sprintf("sandman-vuln-%s.json", ts)
			rMal := fmt.Sprintf("sandman-malware-%s.log", ts)

			// Secrets
			printScanHeader("Secrets Scan", path, rSecrets)
			if err := runTool("trivy", buildTrivySecretsArgs(path, rSecrets)); err != nil {
				printScanFailed("secrets scan", err)
				results = append(results, scanResult{"secrets", "failed", rSecrets})
			} else {
				printScanDone("secrets scan", rSecrets)
				results = append(results, scanResult{"secrets", "pass", rSecrets})
			}

			// Code
			printScanHeader("Code Scan (SAST)", path, rCode)
			if err := runTool("opengrep", buildOpengrepArgs(path, rCode)); err != nil {
				printScanFailed("code scan", err)
				results = append(results, scanResult{"code", "failed", rCode})
			} else {
				printScanDone("code scan", rCode)
				results = append(results, scanResult{"code", "pass", rCode})
			}

			// IaC
			printScanHeader("IaC Scan", path, rIaC)
			if err := runTool("trivy", buildTrivyIaCArgs(path, rIaC)); err != nil {
				printScanFailed("IaC scan", err)
				results = append(results, scanResult{"iac", "failed", rIaC})
			} else {
				printScanDone("IaC scan", rIaC)
				results = append(results, scanResult{"iac", "pass", rIaC})
			}

			// Vuln
			printScanHeader("Vulnerability Scan", path, rVuln)
			if err := runTool("trivy", buildTrivyVulnArgs(path, rVuln)); err != nil {
				printScanFailed("vulnerability scan", err)
				results = append(results, scanResult{"vuln", "failed", rVuln})
			} else {
				printScanDone("vulnerability scan", rVuln)
				results = append(results, scanResult{"vuln", "pass", rVuln})
			}

			// Malware
			printScanHeader("Malware Scan", path, rMal)
			found, err := runMalwareScan(buildClamAVArgs(path, rMal))
			if err != nil {
				printScanFailed("malware scan", err)
				results = append(results, scanResult{"malware", "failed", rMal})
			} else if found {
				fmt.Println()
				fmt.Println(divider)
				fmt.Println("⚠️  Sandman malware scan complete — threats detected!")
				fmt.Printf("   Report saved → %s\n", rMal)
				fmt.Println(divider)
				results = append(results, scanResult{"malware", "findings", rMal})
			} else {
				printScanDone("malware scan", rMal)
				results = append(results, scanResult{"malware", "pass", rMal})
			}
			
			// SBOM  
			rSbom := fmt.Sprintf("sandman-sbom-%s.json", ts)  
			printScanHeader("SBOM Generation", path, rSbom)  
			if err := runTool("trivy", buildTrivySbomArgs(path, "cyclonedx", "fs", rSbom)); err != nil {  
				printScanFailed("SBOM generation", err)  
				results = append(results, scanResult{"sbom", "failed", rSbom})  
			} else {  
				printScanDone("SBOM generation", rSbom)  
				results = append(results, scanResult{"sbom", "pass", rSbom})  
			}  
  
			// License  
			rLicense := fmt.Sprintf("sandman-license-%s.json", ts)  
			printScanHeader("License Scan", path, rLicense)  
			if err := runTool("trivy", buildTrivyLicenseArgs(path, rLicense)); err != nil {  
				printScanFailed("license scan", err)  
				results = append(results, scanResult{"license", "failed", rLicense})  
			} else {  
				printScanDone("license scan", rLicense)  
				results = append(results, scanResult{"license", "pass", rLicense})  
			}
		}

		// ── DAST ─────────────────────────────────────────────────────
		if target != "" {
			r := fmt.Sprintf("sandman-dast-%s.json", ts)
			binary, zapArgs := buildZapArgs(target, full, "", r)
			printScanHeader("DAST Scan", target, r)
			if err := runTool(binary, zapArgs); err != nil {
				printScanFailed("DAST scan", err)
				results = append(results, scanResult{"dast", "failed", r})
			} else {
				printScanDone("DAST scan", r)
				results = append(results, scanResult{"dast", "pass", r})
			}
		}

		// ── Summary ───────────────────────────────────────────────────
		fmt.Println()
		fmt.Println("══════════════════════════════════════════════════════════")
		fmt.Println("🌙 Sandman — Scan Summary")
		fmt.Println("──────────────────────────────────────────────────────────")
		anyFailed := false
		for _, r := range results {
			var icon string
			switch r.status {
			case "pass":
				icon = "✅"
			case "findings":
				icon = "⚠️ "
				anyFailed = true
			case "failed":
				icon = "❌"
				anyFailed = true
			}
			fmt.Printf("   %-10s %s  %s\n", r.label, icon, r.report)
		}
		fmt.Println("══════════════════════════════════════════════════════════")
		fmt.Println()

		if anyFailed {
			os.Exit(1)
		}
		return nil
	},
}

// ------------------------------------------------------------
// Argument builders
// ------------------------------------------------------------

func buildTrivyImageArgs(target, reportPath string) []string {
	return []string{"image", "--severity", severity, "--format", format, "--output", reportPath, target}
}

func buildTrivySecretsArgs(path, reportPath string) []string {
	return []string{"fs", "--scanners", "secret", "--severity", severity, "--format", format, "--output", reportPath, path}
}

func buildTrivyIaCArgs(path, reportPath string) []string {
	return []string{"config", "--severity", severity, "--format", format, "--output", reportPath, path}
}

func buildTrivyVulnArgs(path, reportPath string) []string {
	return []string{"fs", "--scanners", "vuln", "--severity", severity, "--format", format, "--output", reportPath, path}
}

func buildOpengrepArgs(path, reportPath string) []string {
	args := []string{"scan", "--config", "p/default"}
	switch format {
	case "json":
		args = append(args, "--json")
	case "sarif":
		args = append(args, "--sarif")
	}
	return append(args, "--output", reportPath, path)
}

// buildClamAVArgs constructs clamscan arguments.
// ClamAV uses --log for file output; JSON output is not supported natively.
func buildClamAVArgs(path, reportPath string) []string {
	return []string{"--recursive", "--infected", "--log=" + reportPath, path}
}

// buildZapArgs selects the correct ZAP script and constructs its arguments.
// ZAP uses -J (json), -r (html), -x (xml) for output rather than --format/--output.
func buildZapArgs(target string, full bool, apiSpec, reportPath string) (string, []string) {
	var binary string
	switch {
	case apiSpec != "":
		binary = "zap-api-scan.py"
	case full:
		binary = "zap-full-scan.py"
	default:
		binary = "zap-baseline.py"
	}

	args := []string{"-t", target}

	if apiSpec != "" {
		args = append(args, "-f", "openapi", "-S", apiSpec)
	}

	switch format {
	case "json":
		args = append(args, "-J", reportPath)
	case "xml":
		args = append(args, "-x", reportPath)
	default:
		args = append(args, "-r", reportPath)
	}

	return binary, args
}

func buildTrivySbomArgs(target, sbomFormat, targetType, reportPath string) []string {  
	var args []string  
	switch targetType {  
	case "fs":  
		args = []string{"fs", "--format", sbomFormat}  
	default:  
		args = []string{"image", "--format", sbomFormat}  
	}  
	args = append(args, "--output", reportPath)  
	return append(args, target)  
}  
  
func buildTrivyLicenseArgs(path, reportPath string) []string {  
	args := []string{"fs", "--scanners", "license", "--severity", severity, "--format", format, "--output", reportPath}  
	return append(args, path)  
}

// ------------------------------------------------------------
// Tool runners
// ------------------------------------------------------------

func runTool(binary string, args []string) error {
	if _, err := exec.LookPath(binary); err != nil {
		return fmt.Errorf("%s is not installed or not in PATH", binary)
	}
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s exited with: %w", binary, err)
	}
	return nil
}

// runMalwareScan handles ClamAV's exit codes:
//
//	0 = no threats (clean)
//	1 = threats detected (scan succeeded — findings reported)
//	2 = scan error
func runMalwareScan(args []string) (bool, error) {
	if _, err := exec.LookPath("clamscan"); err != nil {
		return false, fmt.Errorf("clamscan is not installed or not in PATH")
	}
	cmd := exec.Command("clamscan", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("clamscan exited with: %w", err)
}

// ------------------------------------------------------------
// init
// ------------------------------------------------------------

func init() {
	scanCmd.PersistentFlags().StringVar(&severity, "severity", "HIGH,CRITICAL", "Severity levels to report (e.g. HIGH,CRITICAL,MEDIUM)")
	scanCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format: json, table, sarif (html/xml for DAST)")
	scanCmd.PersistentFlags().StringVar(&output, "output", "", "Override report file path (default: auto-generated per scan type)")

	// scan sbom flags  
	scanSbomCmd.Flags().String("sbom-format", "cyclonedx", "SBOM format: cyclonedx, spdx-json, spdx")  
	scanSbomCmd.Flags().String("type", "image", "Target type: image or fs")

	scanDastCmd.Flags().Bool("full", false, "Run a full active scan instead of the passive baseline scan")
	scanDastCmd.Flags().String("api-spec", "", "OpenAPI/Swagger spec URL or file path for API scanning")

	scanAllCmd.Flags().String("image", "", "Container image to scan (e.g. nginx:latest)")
	scanAllCmd.Flags().String("path", "", "Filesystem path for secrets, code, IaC, vuln, and malware scans")
	scanAllCmd.Flags().String("target", "", "Live URL for DAST scan (e.g. https://example.com)")
	scanAllCmd.Flags().Bool("full", false, "Use full active ZAP scan instead of baseline")

	scanCmd.AddCommand(scanSbomCmd)  
	scanCmd.AddCommand(scanLicenseCmd)
	scanCmd.AddCommand(scanImageCmd)
	scanCmd.AddCommand(scanSecretsCmd)
	scanCmd.AddCommand(scanCodeCmd)
	scanCmd.AddCommand(scanIaCCmd)
	scanCmd.AddCommand(scanVulnCmd)
	scanCmd.AddCommand(scanMalwareCmd)
	scanCmd.AddCommand(scanDastCmd)
	scanCmd.AddCommand(scanAllCmd)
	rootCmd.AddCommand(scanCmd)
}
