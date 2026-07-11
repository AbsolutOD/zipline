package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <alias>",
	Short: "Edit an alias's command in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		a, err := getAlias(st, args[0])
		if err != nil {
			return err
		}
		tmp, err := os.CreateTemp("", "zipline-edit-*.sh")
		if err != nil {
			return err
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.WriteString(a.Command + "\n"); err != nil {
			tmp.Close()
			return err
		}
		if err := tmp.Close(); err != nil {
			return err
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		// $EDITOR may carry arguments ("code --wait").
		parts := append(strings.Fields(editor), tmp.Name())
		ed := exec.Command(parts[0], parts[1:]...)
		ed.Stdin, ed.Stdout, ed.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := ed.Run(); err != nil {
			return fmt.Errorf("editor failed, alias unchanged: %w", err)
		}
		content, err := os.ReadFile(tmp.Name())
		if err != nil {
			return err
		}
		newCommand := strings.TrimSpace(string(content))
		if newCommand == "" {
			return errors.New("empty command, alias unchanged")
		}
		a.Command = newCommand
		if err := st.Update(a); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Updated alias %q\n", a.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
