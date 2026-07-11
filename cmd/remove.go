package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <alias>",
	Aliases: []string{"rm"},
	Short:   "Remove an alias",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		if _, err := getAlias(st, args[0]); err != nil {
			return err
		}
		if err := st.Delete(args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Removed alias %q\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
