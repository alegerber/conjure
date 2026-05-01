package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alegerber/conjure/internal/api"
	"github.com/alegerber/conjure/internal/heuristic"
	"github.com/alegerber/conjure/internal/keyring"
	"github.com/alegerber/conjure/internal/prompt"
)

type generateOpts struct {
	explain bool
	noEx    bool
	copy    bool
	run     bool
}

func defaultModel() string {
	if v := os.Getenv("CONJURE_MODEL"); v != "" {
		return v
	}
	return "claude-haiku-4-5"
}

func defaultMaxTokens() int {
	if v := os.Getenv("CONJURE_MAX_TOKENS"); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		if n > 0 {
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

func runGenerate(opts generateOpts, description string) error {
	_, key, err := keyring.Resolve()
	if err != nil {
		return fmt.Errorf("API key not found. Run: conjure setup  (or set %s)", keyring.EnvVar)
	}
	osHint := defaultOS()
	systemPrompt := prompt.BuildSystem(osHint)
	client := api.New(key, defaultModel(), defaultMaxTokens())

	ctx := context.Background()
	var cmdLine, explanation string
	if opts.explain {
		out, err := client.GenerateExplain(ctx, systemPrompt, description)
		if err != nil {
			return err
		}
		cmdLine, explanation = out.Command, out.Explanation
	} else {
		out, err := client.GeneratePlain(ctx, systemPrompt, description)
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
