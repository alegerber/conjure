package main

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conjure [description]",
		Short: "Natural-language → Unix command, via the Anthropic Messages API",
		Long: `conjure translates a natural-language description into a single-line
Unix command by calling the Anthropic Messages API directly.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	return cmd
}
