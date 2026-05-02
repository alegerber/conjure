package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/alegerber/conjure/internal/keyring"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Store your Anthropic API key in the system keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := keyring.NewSystemStore()
			if err != nil {
				return fmt.Errorf("init keyring: %w", err)
			}
			if existing, err := store.Get(); err == nil && existing != "" {
				fmt.Print("A key is already stored. Overwrite? [y/N] ")
				reader := bufio.NewReader(os.Stdin)
				ans, _ := reader.ReadString('\n')
				if a := strings.TrimSpace(strings.ToLower(ans)); a != "y" && a != "yes" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			fmt.Print("Anthropic API key (hidden, get one at https://console.anthropic.com/settings/keys): ")
			keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("read input: %w", err)
			}
			key := strings.TrimSpace(string(keyBytes))
			if key == "" {
				return fmt.Errorf("empty input, nothing stored")
			}
			if err := store.Set(key); err != nil {
				return fmt.Errorf("write keyring: %w", err)
			}
			fmt.Printf("✓ API key stored in %s.\n", store.Source())
			return nil
		},
	}
}
