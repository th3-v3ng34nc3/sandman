package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "sandman",
	Short: "Sandman: A multi-purpose security scanner (Trivy & Opengrep)",
	Long:  `A unified security tool for scanning container images (vulnerabilities) and source code (SAST) across Linux, Windows, and K8s.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Sandman version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sandman %s\n", Version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
}