// Package provider defines the contract every LLM backend implements and a
// factory that constructs concrete providers from a Spec.
package provider

import (
	"context"
	"fmt"
)

// Kind identifies a backend implementation.
type Kind string

const (
	KindAnthropic Kind = "anthropic"
	KindOpenAI    Kind = "openai"
	KindOllama    Kind = "ollama"
	KindCodex     Kind = "codex"
	KindClaudeCLI Kind = "claude-cli"
)

// allKinds lists every supported provider kind in the order shown to users.
// Kept private so callers can't mutate the slice in place; expose via Kinds().
var allKinds = []Kind{KindAnthropic, KindOpenAI, KindOllama, KindCodex, KindClaudeCLI}

// Kinds returns a copy of the supported provider kinds in display order.
func Kinds() []Kind {
	out := make([]Kind, len(allKinds))
	copy(out, allKinds)
	return out
}

// ParseKind validates and converts a string to a Kind.
func ParseKind(s string) (Kind, error) {
	k := Kind(s)
	for _, v := range allKinds {
		if v == k {
			return k, nil
		}
	}
	return "", fmt.Errorf("unknown provider %q (want one of: %v)", s, allKinds)
}

// EmitCommand is the structured payload returned by GenerateExplain.
type EmitCommand struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
}

// Provider is the contract every backend implements.
type Provider interface {
	Name() string
	GeneratePlain(ctx context.Context, system, task string) (string, error)
	GenerateExplain(ctx context.Context, system, task string) (*EmitCommand, error)
}

// Spec captures the configuration needed to build a Provider. Fields not
// applicable to a given Kind are ignored.
type Spec struct {
	Kind         Kind
	Model        string
	MaxTokens    int
	BaseURL      string            // override endpoint base (openai/ollama)
	ExtraHeaders map[string]string // optional headers
	AuthFile     string            // codex: path to auth.json
}

// DefaultModel returns a sensible default model for the kind, or "" if the
// user must pick one (ollama).
func DefaultModel(k Kind) string {
	switch k {
	case KindAnthropic:
		return "claude-haiku-4-5"
	case KindOpenAI, KindCodex:
		return "gpt-4o-mini"
	case KindClaudeCLI:
		return "" // claude CLI picks its own
	}
	return ""
}
