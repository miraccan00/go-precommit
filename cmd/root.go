package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-precommit",
	Short: "A fast pre-commit hook runner written in Go",
	Long: `go-precommit is a fast alternative to pre-commit, written in Go.
It uses go-git for git operations without requiring the git binary.

Compatible with .pre-commit-config.yaml format.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(newInstallCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newGlobalInstallCmd())
	rootCmd.AddCommand(newGlobalUninstallCmd())
}
