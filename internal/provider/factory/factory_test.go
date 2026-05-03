package factory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alegerber/conjure/internal/provider"
)

func TestNew_Anthropic(t *testing.T) {
	spec := provider.Spec{
		Kind:      provider.KindAnthropic,
		Model:     "claude-haiku-4-5",
		MaxTokens: 256,
	}
	p, err := New(spec, "sk-ant-test")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != string(provider.KindAnthropic) {
		t.Errorf("Name = %q, want %q", p.Name(), provider.KindAnthropic)
	}
}

func TestNew_AnthropicMissingKey(t *testing.T) {
	spec := provider.Spec{Kind: provider.KindAnthropic, Model: "x", MaxTokens: 1}
	_, err := New(spec, "")
	if err == nil || !strings.Contains(err.Error(), "API key required") {
		t.Errorf("err = %v, want API-key-required", err)
	}
}

func TestNew_OpenAI(t *testing.T) {
	spec := provider.Spec{
		Kind:      provider.KindOpenAI,
		Model:     "gpt-4o-mini",
		MaxTokens: 256,
	}
	p, err := New(spec, "sk-openai-test")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != string(provider.KindOpenAI) {
		t.Errorf("Name = %q, want %q", p.Name(), provider.KindOpenAI)
	}
}

func TestNew_OpenAIMissingKey(t *testing.T) {
	spec := provider.Spec{Kind: provider.KindOpenAI, Model: "x", MaxTokens: 1}
	_, err := New(spec, "")
	if err == nil || !strings.Contains(err.Error(), "API key required") {
		t.Errorf("err = %v, want API-key-required", err)
	}
}

func TestNew_OpenAIWithBaseURL(t *testing.T) {
	spec := provider.Spec{
		Kind:      provider.KindOpenAI,
		Model:     "gpt-4o-mini",
		MaxTokens: 256,
		BaseURL:   "https://proxy.example.com",
	}
	p, err := New(spec, "sk-openai-test")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != string(provider.KindOpenAI) {
		t.Errorf("Name = %q, want %q", p.Name(), provider.KindOpenAI)
	}
}

func TestNew_Ollama(t *testing.T) {
	spec := provider.Spec{
		Kind:    provider.KindOllama,
		Model:   "llama3",
		BaseURL: "http://localhost:11434",
	}
	p, err := New(spec, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != string(provider.KindOllama) {
		t.Errorf("Name = %q, want %q", p.Name(), provider.KindOllama)
	}
}

func TestNew_Codex(t *testing.T) {
	dir := t.TempDir()
	authFile := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(authFile, []byte(`{"OPENAI_API_KEY":"sk-codex-test"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	spec := provider.Spec{
		Kind:      provider.KindCodex,
		Model:     "gpt-4o-mini",
		MaxTokens: 256,
		AuthFile:  authFile,
	}
	p, err := New(spec, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != string(provider.KindCodex) {
		t.Errorf("Name = %q, want %q", p.Name(), provider.KindCodex)
	}
}

func TestNew_CodexMissingAuth(t *testing.T) {
	spec := provider.Spec{
		Kind:      provider.KindCodex,
		Model:     "gpt-4o-mini",
		MaxTokens: 256,
		AuthFile:  "/definitely/not/a/real/path/auth.json",
	}
	_, err := New(spec, "")
	if err == nil {
		t.Fatal("expected error for missing codex auth file")
	}
}

func TestNew_ClaudeCLI(t *testing.T) {
	spec := provider.Spec{Kind: provider.KindClaudeCLI}
	p, err := New(spec, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if p.Name() != string(provider.KindClaudeCLI) {
		t.Errorf("Name = %q, want %q", p.Name(), provider.KindClaudeCLI)
	}
}

func TestNew_UnknownKind(t *testing.T) {
	spec := provider.Spec{Kind: provider.Kind("nope")}
	_, err := New(spec, "")
	if err == nil || !strings.Contains(err.Error(), "unknown provider kind") {
		t.Errorf("err = %v, want unknown-provider-kind", err)
	}
}
