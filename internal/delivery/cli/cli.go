package cli

import (
	"context"
	"io"

	"github.com/spf13/cobra"
)

// Run builds the root Cobra command, attaches the provided subcommands, and
// executes it against the given args. Callers are responsible for constructing
// the subcommands with their own dependencies before calling Run.
func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer, cmds ...*cobra.Command) error {
	rootCmd := newRootCommand(cmds...)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)
	rootCmd.SetContext(ctx)

	return rootCmd.Execute()
}

func newRootCommand(cmds ...*cobra.Command) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "app",
		Short:         "URL shortener",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.AddCommand(cmds...)

	return rootCmd
}
