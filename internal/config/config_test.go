package config

import (
	"os"
	"runtime"
	"testing"
)

func setConfigHome(t *testing.T, dir string) {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		t.Setenv("HOME", dir)
	case "windows":
		t.Setenv("AppData", dir)
	default:
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
}

func TestLoad_MissingFileReturnsZeroConfig(t *testing.T) {
	dir := t.TempDir()
	setConfigHome(t, dir)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg == nil {
		t.Fatal("cfg is nil")
	}
	if cfg.Provider != "" || cfg.Model != "" {
		t.Errorf("expected zero-value config, got %+v", cfg)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	setConfigHome(t, dir)
	want := &Config{
		Provider:      "openai",
		Model:         "gpt-4o-mini",
		OllamaHost:    "http://localhost:11434",
		OpenAIBaseURL: "https://example.com",
	}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	p, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("perm = %o, want 0600", info.Mode().Perm())
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if *got != *want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
