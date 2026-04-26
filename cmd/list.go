package cmd

import (
	"fmt"

	"github.com/keegan-ferrett/gitignore/internal/github"
	"github.com/spf13/cobra"
)

// listCmd prints every root-level template available in github/gitignore,
// one name per line, suitable for piping into other tools.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available .gitignore templates from github/gitignore",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := github.ListTemplates(cmd.Context())
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		for _, n := range names {
			fmt.Fprintln(out, n)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
