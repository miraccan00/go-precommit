package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/miraccan00/go-precommit/internal/config"
	"github.com/miraccan00/go-precommit/internal/gitutil"
	"github.com/miraccan00/go-precommit/internal/runner"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var allFiles bool
	var files []string
	var configFile string
	var verbose bool
	var stage string

	cmd := &cobra.Command{
		Use:   "run [hook-id...]",
		Short: "Run hooks",
		Long: `Run configured hooks against staged files (or all files with --all-files).
Optionally pass specific hook IDs to run only those hooks.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHooks(configFile, allFiles, files, args, verbose, stage)
		},
	}

	cmd.Flags().BoolVar(&allFiles, "all-files", false, "Run on all tracked files instead of staged files")
	cmd.Flags().StringSliceVar(&files, "files", nil, "Run on specific files (comma-separated or repeated)")
	cmd.Flags().StringVarP(&configFile, "config", "c", config.ConfigFile, "Path to config file")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show hook output even when passing")
	cmd.Flags().StringVar(&stage, "stage", "", "Only run hooks matching this stage (pre-commit, pre-push)")

	return cmd
}

func runHooks(configFile string, allFiles bool, specificFiles, hookIDs []string, verbose bool, stage string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			fmt.Fprintln(os.Stderr, "[go-precommit] warning: no .pre-commit-config.yaml found — skipping hooks.")
			fmt.Fprintln(os.Stderr, "[go-precommit] To get started, see: https://github.com/miraccan00/go-precommit/wiki/Getting-Started")
			return nil
		}
		// parse / IO hatası — mesajı olduğu gibi döndür
		return err
	}

	var filesToCheck []string
	switch {
	case len(specificFiles) > 0:
		filesToCheck = specificFiles
	case allFiles:
		filesToCheck, err = gitutil.AllFiles(".")
		if err != nil {
			// Boş repo (henüz commit yok) — staged'e düş
			filesToCheck, err = gitutil.StagedFiles(".")
			if err != nil {
				return fmt.Errorf("getting staged files: %w", err)
			}
		}
	default:
		filesToCheck, err = gitutil.StagedFiles(".")
		if err != nil {
			return fmt.Errorf("getting staged files: %w", err)
		}
	}

	r := runner.New(cfg, verbose)
	if !r.Run(filesToCheck, hookIDs, stage) {
		os.Exit(1)
	}
	return nil
}
