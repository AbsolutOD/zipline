package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <alias>",
	Short: "Show full details of an alias",
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
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, ' ', 0)
		fmt.Fprintf(w, "Name:\t%s\n", a.Name)
		fmt.Fprintf(w, "Command:\t%s\n", a.Command)
		if a.Description != "" {
			fmt.Fprintf(w, "Description:\t%s\n", a.Description)
		}
		fmt.Fprintf(w, "Uses:\t%d\n", a.UseCount)
		fmt.Fprintf(w, "Last used:\t%s\n", humanTime(a.LastUsedAt))
		fmt.Fprintf(w, "Created:\t%s\n", a.CreatedAt.Format("2006-01-02"))
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
