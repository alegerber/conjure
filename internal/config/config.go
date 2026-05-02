// Package config persists conjure's user-level settings to
// $XDG_CONFIG_HOME/conjure/config.json (or the OS equivalent).
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Config is the on-disk representation. All fields are optional; a missing
// file produces a zero-value Config.
type Config struct {
	Provider      string `json:"provider,omitempty"`
	Model         string `json:"model,omitempty"`
	OllamaHost    string `json:"ollama_host,omitempty"`
	OpenAIBaseURL string `json:"openai_base_url,omitempty"`
}

// Path returns the absolute path to the config file. It honours
// $XDG_CONFIG_HOME on Linux via os.UserConfigDir.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate config dir: %w", err)
	}
	return filepath.Join(dir, "conjure", "config.json"), nil
}

// Load reads the config file. A missing file is not an error: a zero-value
// Config is returned.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", p, err)
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	return &c, nil
}

// Save writes the config atomically with mode 0600.
func Save(c *Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(p), err)
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(p, b, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	return nil
}
