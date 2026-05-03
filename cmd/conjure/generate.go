package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alegerber/conjure/internal/clipboard"
	"github.com/alegerber/conjure/internal/config"
	"github.com/alegerber/conjure/internal/heuristic"
	"github.com/alegerber/conjure/internal/keyring"
	"github.com/alegerber/conjure/internal/prompt"
	"github.com/alegerber/conjure/internal/provider"
	anthropicprov "github.com/alegerber/conjure/internal/provider/anthropic"
	"github.com/alegerber/conjure/internal/provider/factory"
	openaiprov "github.com/alegerber/conjure/internal/provider/openai"
	"github.com/alegerber/conjure/internal/runner"
	"github.com/alegerber/conjure/internal/sanitize"
)

type generateOpts struct {
	explain    bool
	noEx       bool
	copy       bool
	run        bool
	provider   string
	model      string
	ollamaHost string
}

const defaultMaxTokens = 256

// resolveProvider builds a Provider from flags + on-disk config. Order
// of precedence: --provider / --model flags > config file > per-kind
// defaults.
func resolveProvider(cfg *config.Config, opts generateOpts) (provider.Provider, error) {
	kindStr := opts.provider
	if kindStr == "" {
		kindStr = cfg.Provider
	}
	if kindStr == "" {
		// Back-compat: if a legacy keyring entry exists, default to anthropic
		// so existing users don't have to re-run setup.
		if _, _, err := keyring.ResolveFor(anthropicprov.KeyringService, anthropicprov.EnvVar); err == nil {
			kindStr = string(provider.KindAnthropic)
		} else {
			return nil, fmt.Errorf("no provider configured. Run: conjure setup")
		}
	}
	kind, err := provider.ParseKind(kindStr)
	if err != nil {
		return nil, err
	}

	model := opts.model
	if model == "" {
		model = cfg.Model
	}
	if model == "" {
		model = provider.DefaultModel(kind)
	}

	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	spec := provider.Spec{
		Kind:      kind,
		Model:     model,
		MaxTokens: maxTokens,
	}

	var key string
	switch kind {
	case provider.KindAnthropic:
		_, k, err := keyring.ResolveFor(anthropicprov.KeyringService, anthropicprov.EnvVar)
		if err != nil {
			return nil, fmt.Errorf("anthropic API key not found. Run: conjure setup (or set %s)", anthropicprov.EnvVar)
		}
		key = k
	case provider.KindOpenAI:
		_, k, err := keyring.ResolveFor(openaiprov.KeyringService, openaiprov.EnvVar)
		if err != nil {
			return nil, fmt.Errorf("OpenAI API key not found. Run: conjure setup (or set %s)", openaiprov.EnvVar)
		}
		key = k
		spec.BaseURL = cfg.OpenAIBaseURL
	case provider.KindOllama:
		spec.BaseURL = resolveOllamaHost(opts.ollamaHost, cfg.OllamaHost)
		if spec.Model == "" {
			return nil, fmt.Errorf("ollama: no model configured. Set with --model or `conjure setup`")
		}
	case provider.KindCodex:
		spec.AuthFile = cfg.CodexAuthFile
	case provider.KindClaudeCLI:
		// no credentials needed
	}

	return factory.New(spec, key)
}

func runGenerate(cfg *config.Config, opts generateOpts, description string) error {
	osHint := cfg.OS
	if osHint == "" {
		osHint = runtime.GOOS
	}
	systemPrompt := prompt.BuildSystem(osHint)

	prov, err := resolveProvider(cfg, opts)
	if err != nil {
		return err
	}

	ctx := context.Background()
	var cmdLine, explanation string
	if opts.explain {
		out, err := prov.GenerateExplain(ctx, systemPrompt, description)
		if err != nil {
			return err
		}
		cmdLine, explanation = out.Command, out.Explanation
	} else {
		out, err := prov.GeneratePlain(ctx, systemPrompt, description)
		if err != nil {
			return err
		}
		cmdLine = out
	}

	cmdLine = sanitize.CommandLine(cmdLine)
	explanation = sanitize.CommandLine(explanation)
	if cmdLine == "" {
		return fmt.Errorf("model returned no command after sanitization")
	}

	for _, h := range heuristic.CheckGNU(cmdLine, osHint) {
		fmt.Fprintln(os.Stderr, "conjure: heuristic warning — "+h)
	}

	fmt.Println(cmdLine)
	if explanation != "" {
		fmt.Printf("# %s\n", explanation)
	}

	if opts.copy {
		c, err := clipboard.New()
		if err != nil {
			return fmt.Errorf("--copy: %w", err)
		}
		if err := c.Copy(cmdLine); err != nil {
			return fmt.Errorf("--copy: %w", err)
		}
		fmt.Fprintf(os.Stderr, "conjure: copied via %s\n", c.Source())
	}

	if opts.run {
		err := runner.ConfirmAndRun(cmdLine, os.Stdin, os.Stdout)
		if errors.Is(err, runner.ErrAborted) {
			fmt.Fprintln(os.Stderr, "conjure: aborted")
			return nil
		}
		return err
	}
	return nil
}

func resolveExplain(flagExplain, flagNoExplain, cfgExplain bool) bool {
	if flagNoExplain {
		return false
	}
	if flagExplain {
		return true
	}
	return cfgExplain
}

func attachGenerateFlags(cmd *cobra.Command, opts *generateOpts) {
	cmd.Flags().BoolVarP(&opts.explain, "explain", "e", false, "Generate command + brief explanation (uses tool-use)")
	cmd.Flags().BoolVar(&opts.noEx, "no-explain", false, "Force plain output (overrides config explain=true)")
	cmd.Flags().BoolVar(&opts.copy, "copy", false, "Copy generated command to system clipboard")
	cmd.Flags().BoolVar(&opts.run, "run", false, "After printing, prompt to execute the command")
	cmd.Flags().StringVar(&opts.provider, "provider", "", "Override configured provider (anthropic|openai|ollama|codex|claude-cli)")
	cmd.Flags().StringVar(&opts.model, "model", "", "Override configured model")
	cmd.Flags().StringVar(&opts.ollamaHost, "ollama-host", "", "Override Ollama host URL (precedence: flag > $OLLAMA_HOST > config)")
}

// resolveOllamaHost picks the ollama host with precedence: --ollama-host flag
// > $OLLAMA_HOST env var > config file > default (handled downstream by
// ollama.New). An empty result means "use the default".
func resolveOllamaHost(flagVal, cfgVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if env := os.Getenv("OLLAMA_HOST"); env != "" {
		return env
	}
	return cfgVal
}

// rootRunE is wired into the root command so `conjure "<description>"` works
// without an explicit subcommand.
func rootRunE(opts *generateOpts) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		opts.explain = resolveExplain(opts.explain, opts.noEx, cfg.Explain)
		return runGenerate(cfg, *opts, strings.Join(args, " "))
	}
}
