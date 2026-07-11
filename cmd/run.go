package cmd

import (
	"fmt"
	"io"

	"github.com/AbsolutOD/zipline/internal/execute"
	"github.com/AbsolutOD/zipline/internal/store"
	"github.com/spf13/cobra"
)

// testExecer, when non-nil, replaces the real syscall execer in tests.
var testExecer execute.Execer

var runCmd = &cobra.Command{
	Use:   "run <alias> [-- <args>...]",
	Short: "Run a stored alias, appending any extra arguments",
	Long: `Run the command stored under <alias>, appending any extra arguments.
The zipline process is replaced by the command (exec), so interactive
containers, signals, and exit codes behave as if you ran it directly.`,
	// Docker-style args (-it, --rm, ...) must pass through untouched.
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
			return cmd.Help()
		}
		st, err := openStore()
		if err != nil {
			return err
		}
		passArgs := args[1:]
		if len(passArgs) > 0 && passArgs[0] == "--" {
			passArgs = passArgs[1:]
		}
		var ex execute.Execer = execute.SyscallExecer{}
		if testExecer != nil {
			ex = testExecer
		}
		return runAlias(st, ex, args[0], passArgs, cmd.ErrOrStderr())
	},
}

// runAlias looks up the alias, records the use, and execs the command.
func runAlias(st *store.Store, ex execute.Execer, aliasName string, args []string, errW io.Writer) error {
	a, err := getAlias(st, aliasName)
	if err != nil {
		return err
	}
	// Tracking must never block execution.
	if err := st.Touch(a.Name); err != nil {
		_, _ = fmt.Fprintf(errW, "zipline: warning: failed to record usage: %v\n", err)
	}
	return execute.Run(ex, a.Command, a.Name, args)
}

func init() {
	rootCmd.AddCommand(runCmd)
}
