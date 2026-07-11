// Package cmd implements the zipline CLI.
package cmd

import (
	"errors"
	"os"

	"github.com/AbsolutOD/zipline/internal/store"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zipline",
	Short: "Manage long docker run commands as shell aliases",
	Long: `Zipline stores long commands (docker run, podman run, ...) under short
aliases, tracks how often you use them, and hooks into bash/zsh so each
alias works like a native command.

Add the hook to your shell rc file:

  eval "$(zipline init zsh)"    # or: zipline init bash

Then:

  zl add runlike "docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro assaflavie/runlike"
  runlike <container>`,
	SilenceUsage: true,
}

// Root returns the root command; used by the docgen tool.
func Root() *cobra.Command { return rootCmd }

// Execute runs the CLI. Exit codes: 0 success, 2 unknown alias,
// 1 anything else.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
