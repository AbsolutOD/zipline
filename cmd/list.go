package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	listSort  string
	listQuiet bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List stored aliases",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if listSort != "usage" && listSort != "name" {
			return fmt.Errorf("invalid --sort %q (valid: usage, name)", listSort)
		}
		st, err := openStore()
		if err != nil {
			return err
		}
		aliases, err := st.List(listSort)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if listQuiet {
			for _, a := range aliases {
				fmt.Fprintln(out, a.Name)
			}
			return nil
		}
		if len(aliases) == 0 {
			fmt.Fprintln(out, `No aliases stored. Add one with: zipline add <alias> "<command>"`)
			return nil
		}
		w := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
		fmt.Fprintln(w, "ALIAS\tUSES\tLAST USED\tCOMMAND")
		for _, a := range aliases {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", a.Name, a.UseCount, humanTime(a.LastUsedAt), truncate(a.Command, 60))
		}
		return w.Flush()
	},
}

func init() {
	listCmd.Flags().StringVar(&listSort, "sort", "usage", "sort order: usage or name")
	listCmd.Flags().BoolVarP(&listQuiet, "quiet", "q", false, "print alias names only")
	rootCmd.AddCommand(listCmd)
}
