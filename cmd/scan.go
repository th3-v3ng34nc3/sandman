package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	severity string
	format   string
	output   string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Security scanning operations",
}

// ------------------------------------------------------------
// scan image
// ------------------------------------------------------------

var scanImageCmd = &cobra.Command{
	Use:   "image [target]",
	Short: "Scan a container image for vulnerabilities",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runTool("trivy", buildTrivyImageArgs(args[0]), "🌙 Sandman is inspecting image"); err != nil {
			fmt.Printf("❌ Image scan failed: %v\n", err)
			os.Exit(1)
		}
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
		if err := runTool("trivy", buildTrivySecretsArgs(args[0]), "🌙 Sandman is hunting for secrets"); err != nil {
			fmt.Printf("❌ Secrets scan failed: %v\n", err)
			os.Exit(1)
		}
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
		if err := runTool("opengrep", buildOpengrepArgs(args[0]), "🌙 Sandman is auditing code"); err != nil {
			fmt.Printf("❌ Code scan failed: %v\n", err)
			os.Exit(1)
		}
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
		if err := runTool("trivy", buildTrivyIaCArgs(args[0]), "🌙 Sandman is reviewing IaC"); err != nil {
			fmt.Printf("❌ IaC scan failed: %v\n", err)
			os.Exit(1)
		}
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

		binary, zapArgs := buildZapArgs(args[0], full, apiSpec)
		if err := runTool(binary, zapArgs, "🌙 Sandman is probing the target"); err != nil {
			fmt.Printf("❌ DAST scan failed: %v\n", err)
			os.Exit(1)
		}
	},
}

// ------------------------------------------------------------
// scan all
// ------------------------------------------------------------

var scanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all scans in sequence",
	Long: `Run every scan type in sequence and report combined results.

Flags:
  --image     Container image   → image scan
  --path      Filesystem path   → secrets, code, and IaC scans
  --target    Live URL          → DAST baseline scan (add --full for active)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		image, _ := cmd.Flags().GetString("image")
		path, _ := cmd.Flags().GetString("path")
		target, _ := cmd.Flags().GetString("target")
		full, _ := cmd.Flags().GetBool("full")

		if image == "" && path == "" && target == "" {
			return fmt.Errorf("provide at least one of --image, --path, or --target")
		}

		var failed []string

		if image != "" {
			fmt.Println("\n─── Container Image Scan ───────────────────────────────")
			if err := runTool("trivy", buildTrivyImageArgs(image), "🌙 Sandman is inspecting image"); err != nil {
				fmt.Printf("⚠️  %v\n", err)
				failed = append(failed, "image")
			}
		}

		if path != "" {
			fmt.Println("\n─── Secrets Scan ───────────────────────────────────────")
			if err := runTool("trivy", buildTrivySecretsArgs(path), "🌙 Sandman is hunting for secrets"); err != nil {
				fmt.Printf("⚠️  %v\n", err)
				failed = append(failed, "secrets")
			}

			fmt.Println("\n─── Code Scan ──────────────────────────────────────────")
			if err := runTool("opengrep", buildOpengrepArgs(path), "🌙 Sandman is auditing code"); err != nil {
				fmt.Printf("⚠️  %v\n", err)
				failed = append(failed, "code")
			}

			fmt.Println("\n─── IaC Scan ───────────────────────────────────────────")
			if err := runTool("trivy", buildTrivyIaCArgs(path), "🌙 Sandman is reviewing IaC"); err != nil {
				fmt.Printf("⚠️  %v\n", err)
				failed = append(failed, "iac")
			}
		}

		if target != "" {
			fmt.Println("\n─── DAST Scan ──────────────────────────────────────────")
			binary, zapArgs := buildZapArgs(target, full, "")
			if err := runTool(binary, zapArgs, "🌙 Sandman is probing the target"); err != nil {
				fmt.Printf("⚠️  %v\n", err)
				failed = append(failed, "dast")
			}
		}

		fmt.Println()
		if len(failed) > 0 {
			fmt.Printf("⚠️  Scans with findings or failures: %s\n", strings.Join(failed, ", "))
			os.Exit(1)
		}

		fmt.Println("✅ All scans complete. No issues found.")
		return nil
	},
}

// ------------------------------------------------------------
// Argument builders
// ------------------------------------------------------------

func buildTrivyImageArgs(target string) []string {
	args := []string{"image", "--severity", severity}
	if format != "table" {
		args = append(args, "--format", format)
	}
	if output != "" {
		args = append(args, "--output", output)
	}
	return append(args, target)
}

func buildTrivySecretsArgs(path string) []string {
	args := []string{"fs", "--scanners", "secret", "--severity", severity}
	if format != "table" {
		args = append(args, "--format", format)
	}
	if output != "" {
		args = append(args, "--output", output)
	}
	return append(args, path)
}

func buildTrivyIaCArgs(path string) []string {
	args := []string{"config", "--severity", severity}
	if format != "table" {
		args = append(args, "--format", format)
	}
	if output != "" {
		args = append(args, "--output", output)
	}
	return append(args, path)
}

func buildOpengrepArgs(path string) []string {
	args := []string{"scan", "--config", "p/default"}
	switch format {
	case "json":
		args = append(args, "--json")
	case "sarif":
		args = append(args, "--sarif")
	}
	if output != "" {
		args = append(args, "--output", output)
	}
	return append(args, path)
}

// buildZapArgs selects the correct ZAP script and constructs its arguments.
// ZAP uses -J (json), -r (html), -x (xml) for output rather than --format/--output.
func buildZapArgs(target string, full bool, apiSpec string) (string, []string) {
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

	if output != "" {
		switch format {
		case "json":
			args = append(args, "-J", output)
		case "xml":
			args = append(args, "-x", output)
		default:
			// html is ZAP's richest report format
			args = append(args, "-r", output)
		}
	}

	return binary, args
}

// ------------------------------------------------------------
// Tool runner
// ------------------------------------------------------------

func runTool(binary string, args []string, message string) error {
	fmt.Printf("%s: %s...\n", message, args[len(args)-1])

	if _, err := exec.LookPath(binary); err != nil {
		return fmt.Errorf("%s is not installed or not in PATH", binary)
	}

	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s exited with: %w", binary, err)
	}

	fmt.Printf("✅ %s scan complete.\n", binary)
	return nil
}

// ------------------------------------------------------------
// init
// ------------------------------------------------------------

func init() {
	// Persistent flags inherited by all scan subcommands
	scanCmd.PersistentFlags().StringVar(&severity, "severity", "HIGH,CRITICAL", "Severity levels to report (e.g. HIGH,CRITICAL,MEDIUM)")
	scanCmd.PersistentFlags().StringVar(&format, "format", "table", "Output format: table, json, sarif (html/xml for DAST)")
	scanCmd.PersistentFlags().StringVar(&output, "output", "", "Write results to this file path")

	// scan dast flags
	scanDastCmd.Flags().Bool("full", false, "Run a full active scan instead of the passive baseline scan")
	scanDastCmd.Flags().String("api-spec", "", "OpenAPI/Swagger spec URL or file path for API scanning")

	// scan all flags
	scanAllCmd.Flags().String("image", "", "Container image to scan (e.g. nginx:latest)")
	scanAllCmd.Flags().String("path", "", "Filesystem path for secrets, code, and IaC scans")
	scanAllCmd.Flags().String("target", "", "Live URL for DAST scan (e.g. https://example.com)")
	scanAllCmd.Flags().Bool("full", false, "Use full active ZAP scan instead of baseline (applies to DAST in scan all)")

	scanCmd.AddCommand(scanImageCmd)
	scanCmd.AddCommand(scanSecretsCmd)
	scanCmd.AddCommand(scanCodeCmd)
	scanCmd.AddCommand(scanIaCCmd)
	scanCmd.AddCommand(scanDastCmd)
	scanCmd.AddCommand(scanAllCmd)
	rootCmd.AddCommand(scanCmd)
}
