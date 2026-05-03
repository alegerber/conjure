package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
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

// kindDescriptions are the one-line summaries shown in the setup menu.
// Keep the keys in sync with provider.Kinds().
var kindDescriptions = map[provider.Kind]string{
	provider.KindAnthropic: "Anthropic API (per-token billing)",
	provider.KindOpenAI:    "OpenAI API (per-token billing)",
	provider.KindOllama:    "Local Ollama (no auth)",
	provider.KindCodex:     "ChatGPT Plus/Pro via Codex CLI credentials",
	provider.KindClaudeCLI: "Claude Pro/Max via the `claude` CLI (slower)",
}

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
	kinds := provider.Kinds()

	_, _ = fmt.Fprintln(stdout, "Which provider do you want to use?")
	for i, k := range kinds {
		_, _ = fmt.Fprintf(stdout, "  [%d] %-12s — %s\n", i+1, k, kindDescriptions[k])
	}
	_, _ = fmt.Fprint(stdout, "Choice [1]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" {
		choice = "1"
	}

	picked, err := resolveSetupChoice(choice, kinds)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	switch picked {
	case provider.KindAnthropic:
		if err := setupAPIKey(reader, stdout, anthropicprov.KeyringService, "Anthropic", "https://console.anthropic.com/settings/keys"); err != nil {
			return err
		}
		cfg.Provider = string(provider.KindAnthropic)
		cfg.Model = promptModel(reader, stdout, provider.KindAnthropic, cfg.Model)
	case provider.KindOpenAI:
		if err := setupAPIKey(reader, stdout, openaiprov.KeyringService, "OpenAI", "https://platform.openai.com/api-keys"); err != nil {
			return err
		}
		cfg.Provider = string(provider.KindOpenAI)
		cfg.Model = promptModel(reader, stdout, provider.KindOpenAI, cfg.Model)
	case provider.KindOllama:
		host := promptString(reader, stdout, "Ollama host", firstNonEmpty(cfg.OllamaHost, ollama.DefaultHost))
		cfg.Provider = string(provider.KindOllama)
		cfg.OllamaHost = host
		cfg.Model = promptModel(reader, stdout, provider.KindOllama, cfg.Model)
		if cfg.Model == "" {
			return fmt.Errorf("ollama: a model is required (e.g. `ollama pull llama3.1`)")
		}
	case provider.KindCodex:
		path := cfg.CodexAuthFile
		if path == "" {
			var err error
			path, err = codex.DefaultAuthFile()
			if err != nil {
				return err
			}
		}
		if _, err := codex.LoadAPIKey(path); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "OK Using Codex CLI credentials from %s.\n", path)
		cfg.Provider = string(provider.KindCodex)
		cfg.Model = promptModel(reader, stdout, provider.KindCodex, cfg.Model)
	case provider.KindClaudeCLI:
		if err := claudecli.LookPath(); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, "OK `claude` CLI found on PATH.")
		cfg.Provider = string(provider.KindClaudeCLI)
		cfg.Model = "" // claude CLI picks its own model
	default:
		return fmt.Errorf("unsupported provider kind %q", picked)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	p, _ := config.Path()
	_, _ = fmt.Fprintf(stdout, "OK Saved config to %s.\n", p)
	return nil
}

// resolveSetupChoice maps a user-entered menu choice (numeric index or
// canonical kind name) to a provider.Kind from the given ordered list.
func resolveSetupChoice(choice string, kinds []provider.Kind) (provider.Kind, error) {
	if i, err := strconv.Atoi(choice); err == nil {
		if i >= 1 && i <= len(kinds) {
			return kinds[i-1], nil
		}
		return "", fmt.Errorf("invalid choice %q (pick 1..%d)", choice, len(kinds))
	}
	k, err := provider.ParseKind(choice)
	if err != nil {
		return "", fmt.Errorf("invalid choice %q", choice)
	}
	return k, nil
}

func setupAPIKey(reader *bufio.Reader, stdout io.Writer, service, label, helpURL string) error {
	store, err := keyring.NewSystemStoreFor(service)
	if err != nil {
		return fmt.Errorf("init keyring: %w", err)
	}
	if existing, err := store.Get(); err == nil && existing != "" {
		_, _ = fmt.Fprintf(stdout, "A %s key is already stored. Overwrite? [y/N] ", label)
		ans, _ := reader.ReadString('\n')
		if a := strings.TrimSpace(strings.ToLower(ans)); a != "y" && a != "yes" {
			_, _ = fmt.Fprintln(stdout, "Keeping existing key.")
			return nil
		}
	}

	_, _ = fmt.Fprintf(stdout, "%s API key (hidden, get one at %s): ", label, helpURL)
	keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(stdout)
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
	_, _ = fmt.Fprintf(stdout, "OK %s key stored in %s.\n", label, store.Source())
	return nil
}

func promptString(reader *bufio.Reader, stdout io.Writer, label, dflt string) string {
	if dflt != "" {
		_, _ = fmt.Fprintf(stdout, "%s [%s]: ", label, dflt)
	} else {
		_, _ = fmt.Fprintf(stdout, "%s: ", label)
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
