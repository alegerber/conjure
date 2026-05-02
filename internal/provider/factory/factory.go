// Package factory builds concrete provider.Provider instances from a Spec.
// It lives in a sub-package so the provider package itself can stay free of
// imports of every backend (avoids an import cycle).
package factory

import (
	"fmt"

	"github.com/alegerber/conjure/internal/provider"
	"github.com/alegerber/conjure/internal/provider/anthropic"
	"github.com/alegerber/conjure/internal/provider/claudecli"
	"github.com/alegerber/conjure/internal/provider/codex"
	"github.com/alegerber/conjure/internal/provider/ollama"
	"github.com/alegerber/conjure/internal/provider/openai"
)

// New constructs a concrete Provider for the given Spec. The key argument is
// interpreted per-Kind:
//
//   - anthropic / openai: API key (required)
//   - ollama / claude-cli: ignored
//   - codex: ignored (key is read from spec.AuthFile or default location)
func New(spec provider.Spec, key string) (provider.Provider, error) {
	switch spec.Kind {
	case provider.KindAnthropic:
		if key == "" {
			return nil, fmt.Errorf("anthropic: API key required")
		}
		return anthropic.New(key, spec.Model, spec.MaxTokens), nil
	case provider.KindOpenAI:
		if key == "" {
			return nil, fmt.Errorf("openai: API key required")
		}
		return openai.NewWithBase(key, spec.Model, spec.MaxTokens, spec.BaseURL), nil
	case provider.KindOllama:
		return ollama.New(spec.BaseURL, spec.Model), nil
	case provider.KindCodex:
		return codex.New(spec.AuthFile, spec.Model, spec.MaxTokens)
	case provider.KindClaudeCLI:
		return claudecli.New(), nil
	}
	return nil, fmt.Errorf("unknown provider kind %q", spec.Kind)
}
