// Package execute replaces the zipline process with the stored command
// so TTY, signals, and exit codes behave as if the user ran it directly.
package execute

import (
	"os"
	"os/exec"
	"syscall"
)

// Execer abstracts syscall.Exec for testability.
type Execer interface {
	Exec(argv0 string, argv []string, env []string) error
}

// SyscallExecer replaces the current process; on success it never returns.
type SyscallExecer struct{}

func (SyscallExecer) Exec(argv0 string, argv, env []string) error {
	return syscall.Exec(argv0, argv, env)
}

// Argv builds the sh invocation for a stored command with extra args
// appended: sh -c '<command> "$@"' <alias> [args...]. Running through
// sh means stored commands may use quoting, env vars, and ~ without
// zipline parsing them; the alias name becomes $0 in error messages.
func Argv(command, alias string, args []string) []string {
	return append([]string{"sh", "-c", command + ` "$@"`, alias}, args...)
}

// Run resolves sh on PATH and execs the stored command through it.
func Run(e Execer, command, alias string, args []string) error {
	shPath, err := exec.LookPath("sh")
	if err != nil {
		return err
	}
	return e.Exec(shPath, Argv(command, alias, args), os.Environ())
}
