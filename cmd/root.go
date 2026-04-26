// Package cmd defines the gitignore CLI commands.
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/keegan-ferrett/gitignore/internal/github"
	"github.com/spf13/cobra"
)

const gitignorePath = ".gitignore"

// rootCmd is the entry-point command. Invoking it with a template name
// (e.g. `gitignore C++`) downloads that template from github/gitignore and
// writes it to ./.gitignore. With no arguments it prints help.
var rootCmd = &cobra.Command{
	Use:   "gitignore [template]",
	Short: "Write a .gitignore file using a template from github/gitignore",
	Long: `gitignore is a small CLI that reads template files from the
github/gitignore repository and writes them into a local .gitignore.

Run "gitignore list" to see every available template name.`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return writeTemplate(cmd, args[0])
	},
}

// writeTemplate fetches the named template, prompts the user before
// overwriting an existing .gitignore, then writes the file.
func writeTemplate(cmd *cobra.Command, name string) error {
	body, err := github.FetchTemplate(cmd.Context(), name)
	if err != nil {
		if errors.Is(err, github.ErrTemplateNotFound) {
			return fmt.Errorf("%w (run `gitignore list` to see available templates)", err)
		}
		return err
	}

	if _, statErr := os.Stat(gitignorePath); statErr == nil {
		ok, err := confirmOverwrite(cmd.InOrStdin(), cmd.OutOrStdout())
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", gitignorePath, statErr)
	}

	if err := os.WriteFile(gitignorePath, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", gitignorePath, err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s from %s template (%d bytes)\n", gitignorePath, name, len(body))
	return nil
}

// confirmOverwrite asks the user whether to overwrite the existing
// .gitignore. It accepts y/yes (case-insensitive) as confirmation; anything
// else (including EOF on a piped stdin) is treated as a refusal.
func confirmOverwrite(in io.Reader, out io.Writer) (bool, error) {
	fmt.Fprintf(out, "%s already exists. Overwrite? [y/N]: ", gitignorePath)
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

// Execute runs the root command and exits with a non-zero status on error.
// Cobra prints the error itself (with an "Error: " prefix), so we only need
// to set the exit code here.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
