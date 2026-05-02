package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/alegerber/conjure/internal/config"
	"github.com/alegerber/conjure/internal/keyring"
	"github.com/alegerber/conjure/internal/provider"
	anthropicprov "github.com/alegerber/conjure/internal/provider/anthropic"
	"github.com/alegerber/conjure/internal/provider/claudecli"
	"github.com/alegerber/conjure/internal/provider/codex"
	"github.com/alegerber/conjure/internal/provider/ollama"
	openaiprov "github.com/alegerber/conjure/internal/provider/openai"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Pick an LLM provider and store its credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(os.Stdin, os.Stdout)
		},
	}
}

func runSetup(stdin io.Reader, stdout io.Writer) error {
	reader := bufio.NewReader(stdin)
	fmt.Fprintln(stdout, "Which provider do you want to use?")
	fmt.Fprintln(stdout, "  [1] anthropic    — Anthropic API (per-token billing)")
	fmt.Fprintln(stdout, "  [2] openai       — OpenAI API (per-token billing)")
	fmt.Fprintln(stdout, "  [3] ollama       — Local Ollama (no auth)")
	fmt.Fprintln(stdout, "  [4] codex        — ChatGPT Plus/Pro via Codex CLI credentials")
	fmt.Fprintln(stdout, "  [5] claude-cli   — Claude Pro/Max via the `claude` CLI (slower)")
	fmt.Fprint(stdout, "Choice [1]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" {
		choice = "1"
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	switch choice {
	case "1", "anthropic":
		if err := setupAPIKey(stdout, anthropicprov.KeyringService, "Anthropic", "https://console.anthropic.com/settings/keys"); err != nil {
			return err
		}
		cfg.Provider = string(provider.KindAnthropic)
		cfg.Model = promptModel(reader, stdout, provider.KindAnthropic, cfg.Model)
	case "2", "openai":
		if err := setupAPIKey(stdout, openaiprov.KeyringService, "OpenAI", "https://platform.openai.com/api-keys"); err != nil {
			return err
		}
		cfg.Provider = string(provider.KindOpenAI)
		cfg.Model = promptModel(reader, stdout, provider.KindOpenAI, cfg.Model)
	case "3", "ollama":
		host := promptString(reader, stdout, "Ollama host", firstNonEmpty(cfg.OllamaHost, ollama.DefaultHost))
		cfg.Provider = string(provider.KindOllama)
		cfg.OllamaHost = host
		cfg.Model = promptModel(reader, stdout, provider.KindOllama, cfg.Model)
		if cfg.Model == "" {
			return fmt.Errorf("ollama: a model is required (e.g. `ollama pull llama3.1`)")
		}
	case "4", "codex":
		path, err := codex.DefaultAuthFile()
		if err != nil {
			return err
		}
		if _, err := codex.LoadAPIKey(path); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "✓ Using Codex CLI credentials from %s.\n", path)
		cfg.Provider = string(provider.KindCodex)
		cfg.Model = promptModel(reader, stdout, provider.KindCodex, cfg.Model)
	case "5", "claude-cli", "claude":
		if err := claudecli.LookPath(); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "✓ `claude` CLI found on PATH.")
		cfg.Provider = string(provider.KindClaudeCLI)
		cfg.Model = "" // claude CLI picks its own model
	default:
		return fmt.Errorf("invalid choice %q", choice)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	p, _ := config.Path()
	fmt.Fprintf(stdout, "✓ Saved config to %s.\n", p)
	return nil
}

func setupAPIKey(stdout io.Writer, service, label, helpURL string) error {
	store, err := keyring.NewSystemStoreFor(service)
	if err != nil {
		return fmt.Errorf("init keyring: %w", err)
	}
	if existing, err := store.Get(); err == nil && existing != "" {
		fmt.Fprintf(stdout, "A %s key is already stored. Overwrite? [y/N] ", label)
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		if a := strings.TrimSpace(strings.ToLower(ans)); a != "y" && a != "yes" {
			fmt.Fprintln(stdout, "Keeping existing key.")
			return nil
		}
	}

	fmt.Fprintf(stdout, "%s API key (hidden, get one at %s): ", label, helpURL)
	keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(stdout)
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
	fmt.Fprintf(stdout, "✓ %s key stored in %s.\n", label, store.Source())
	return nil
}

func promptString(reader *bufio.Reader, stdout io.Writer, label, dflt string) string {
	if dflt != "" {
		fmt.Fprintf(stdout, "%s [%s]: ", label, dflt)
	} else {
		fmt.Fprintf(stdout, "%s: ", label)
	}
	v, _ := reader.ReadString('\n')
	v = strings.TrimSpace(v)
	if v == "" {
		return dflt
	}
	return v
}

func promptModel(reader *bufio.Reader, stdout io.Writer, kind provider.Kind, current string) string {
	dflt := current
	if dflt == "" {
		dflt = provider.DefaultModel(kind)
	}
	label := "Model"
	if kind == provider.KindOllama {
		label = "Ollama model (e.g. llama3.1)"
	}
	return promptString(reader, stdout, label, dflt)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
