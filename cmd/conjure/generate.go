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
)

type generateOpts struct {
	explain  bool
	noEx     bool
	copy     bool
	run      bool
	provider string
	model    string
}

const ProviderEnvVar = "CONJURE_PROVIDER"

func defaultMaxTokens() int {
	if v := os.Getenv("CONJURE_MAX_TOKENS"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n <= 0 {
			fmt.Fprintf(os.Stderr, "conjure: ignoring invalid CONJURE_MAX_TOKENS=%q (want positive integer)\n", v)
		} else {
			return n
		}
	}
	return 256
}

func defaultOS() string {
	if v := os.Getenv("CONJURE_OS"); v != "" {
		return v
	}
	return runtime.GOOS
}

func envExplain() bool {
	return os.Getenv("CONJURE_EXPLAIN") == "1"
}

// resolveProvider builds a Provider from flags + env + on-disk config. Order
// of precedence: --provider / --model flags > env vars > config file >
// per-kind defaults.
func resolveProvider(opts generateOpts) (provider.Provider, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	kindStr := opts.provider
	if kindStr == "" {
		kindStr = os.Getenv(ProviderEnvVar)
	}
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
		model = os.Getenv("CONJURE_MODEL")
	}
	if model == "" {
		model = cfg.Model
	}
	if model == "" {
		model = provider.DefaultModel(kind)
	}

	spec := provider.Spec{
		Kind:      kind,
		Model:     model,
		MaxTokens: defaultMaxTokens(),
	}

	var key string
	switch kind {
	case provider.KindAnthropic:
		_, k, err := keyring.ResolveFor(anthropicprov.KeyringService, anthropicprov.EnvVar)
		if err != nil {
			return nil, fmt.Errorf("Anthropic API key not found. Run: conjure setup (or set %s)", anthropicprov.EnvVar)
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
		host := cfg.OllamaHost
		if v := os.Getenv("OLLAMA_HOST"); v != "" {
			host = v
		}
		spec.BaseURL = host
		if spec.Model == "" {
			return nil, fmt.Errorf("ollama: no model configured. Set with --model or `conjure setup`")
		}
	case provider.KindCodex:
		// codex.New() reads the auth file itself.
	case provider.KindClaudeCLI:
		// no credentials needed
	}

	return factory.New(spec, key)
}

func runGenerate(opts generateOpts, description string) error {
	osHint := defaultOS()
	systemPrompt := prompt.BuildSystem(osHint)

	prov, err := resolveProvider(opts)
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

func resolveExplain(flagExplain, flagNoExplain bool) bool {
	if flagNoExplain {
		return false
	}
	if flagExplain {
		return true
	}
	return envExplain()
}

func attachGenerateFlags(cmd *cobra.Command, opts *generateOpts) {
	cmd.Flags().BoolVarP(&opts.explain, "explain", "e", false, "Generate command + brief explanation (uses tool-use)")
	cmd.Flags().BoolVar(&opts.noEx, "no-explain", false, "Force plain output (overrides CONJURE_EXPLAIN)")
	cmd.Flags().BoolVar(&opts.copy, "copy", false, "Copy generated command to system clipboard")
	cmd.Flags().BoolVar(&opts.run, "run", false, "After printing, prompt to execute the command")
	cmd.Flags().StringVar(&opts.provider, "provider", "", "Override configured provider (anthropic|openai|ollama|codex|claude-cli)")
	cmd.Flags().StringVar(&opts.model, "model", "", "Override configured model")
}

// rootRunE is wired into the root command so `conjure "<description>"` works
// without an explicit subcommand.
func rootRunE(opts *generateOpts) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		opts.explain = resolveExplain(opts.explain, opts.noEx)
		return runGenerate(*opts, strings.Join(args, " "))
	}
}
