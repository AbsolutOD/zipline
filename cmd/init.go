package cmd

import (
	"fmt"

	"github.com/AbsolutOD/zipline/internal/config"
	"github.com/AbsolutOD/zipline/internal/shell"
	"github.com/spf13/cobra"
)

var (
	initCmdName   string
	initFuncsOnly bool
)

var initCmd = &cobra.Command{
	Use:   "init <bash|zsh>",
	Short: "Print the shell hook script (add eval \"$(zipline init zsh)\" to your rc file)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		aliases, err := st.List("name")
		if err != nil {
			return err
		}
		names := make([]string, len(aliases))
		for i, a := range aliases {
			names[i] = a.Name
		}
		wrapper := initCmdName
		if wrapper == "" {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			wrapper = cfg.Cmd
		}
		script, err := shell.Script(shell.Data{
			Shell:     args[0],
			Cmd:       wrapper,
			Aliases:   names,
			FuncsOnly: initFuncsOnly,
		})
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), script)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initCmdName, "cmd", "", "wrapper function name (default from config, else zl)")
	initCmd.Flags().BoolVar(&initFuncsOnly, "funcs-only", false, "emit only the per-alias functions (used for refresh)")
	rootCmd.AddCommand(initCmd)
}
