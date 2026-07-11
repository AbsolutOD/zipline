package cmd

import (
	"errors"
	"fmt"

	"github.com/AbsolutOD/zipline/internal/name"
	"github.com/AbsolutOD/zipline/internal/store"
	"github.com/spf13/cobra"
)

var (
	addDescription string
	addForce       bool
)

var addCmd = &cobra.Command{
	Use:     "add <alias> <command>",
	Short:   "Store a command under an alias",
	Example: `  zipline add runlike "docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro assaflavie/runlike"`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliasName, command := args[0], args[1]
		if err := name.Validate(aliasName, reservedNames()); err != nil {
			return err
		}
		st, err := openStore()
		if err != nil {
			return err
		}
		err = st.Add(&store.Alias{Name: aliasName, Command: command, Description: addDescription})
		if errors.Is(err, store.ErrExists) {
			if !addForce {
				return fmt.Errorf("alias %q already exists (use --force to overwrite)", aliasName)
			}
			existing, getErr := st.Get(aliasName)
			if getErr != nil {
				return getErr
			}
			existing.Command = command
			existing.Description = addDescription
			if err := st.Update(existing); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated alias %q\n", aliasName)
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Added alias %q\n", aliasName)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "description of the alias")
	addCmd.Flags().BoolVar(&addForce, "force", false, "overwrite an existing alias")
	rootCmd.AddCommand(addCmd)
}
