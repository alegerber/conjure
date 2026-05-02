package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alegerber/conjure/internal/shell"
)

func newShellInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "shell-init [zsh|bash|fish|powershell]",
		Short:     "Print shell-integration script for the given shell",
		Args:      cobra.ExactArgs(1),
		ValidArgs: shell.SupportedShells(),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := shell.Render(args[0])
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
}
