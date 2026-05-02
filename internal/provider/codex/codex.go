// Package codex reuses the Codex CLI's stored OpenAI credentials at
// ~/.codex/auth.json so users on a ChatGPT subscription don't pay again.
package codex

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alegerber/conjure/internal/provider"
	"github.com/alegerber/conjure/internal/provider/openai"
)

// DefaultAuthFile returns the conventional path to Codex's auth file.
func DefaultAuthFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex", "auth.json"), nil
}

type codexAuth struct {
	OpenAIAPIKey string `json:"OPENAI_API_KEY"`
}

// LoadAPIKey reads the Codex auth file and extracts the OpenAI API key.
func LoadAPIKey(path string) (string, error) {
	if path == "" {
		var err error
		path, err = DefaultAuthFile()
		if err != nil {
			return "", err
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("no Codex auth at %s — run `codex login` first", path)
		}
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	var a codexAuth
	if err := json.Unmarshal(b, &a); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}
	if a.OpenAIAPIKey == "" {
		return "", fmt.Errorf("no OPENAI_API_KEY in %s — run `codex login` first", path)
	}
	return a.OpenAIAPIKey, nil
}

// New returns a Provider that talks to OpenAI using the Codex CLI's stored
// credentials. authFile may be empty (use the default).
func New(authFile, model string, maxTokens int) (provider.Provider, error) {
	key, err := LoadAPIKey(authFile)
	if err != nil {
		return nil, err
	}
	c := openai.New(key, model, maxTokens)
	c.ProviderName = string(provider.KindCodex)
	return c, nil
}
