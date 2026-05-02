package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAPIKey_ReadsField(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(p, []byte(`{"OPENAI_API_KEY":"sk-test-codex"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := LoadAPIKey(p)
	if err != nil {
		t.Fatalf("LoadAPIKey: %v", err)
	}
	if got != "sk-test-codex" {
		t.Errorf("got %q", got)
	}
}

func TestLoadAPIKey_MissingFile(t *testing.T) {
	_, err := LoadAPIKey("/nonexistent/codex/auth.json")
	if err == nil || !strings.Contains(err.Error(), "codex login") {
		t.Errorf("err = %v, want hint to run codex login", err)
	}
}

func TestLoadAPIKey_EmptyKey(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(p, []byte(`{"OPENAI_API_KEY":""}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadAPIKey(p)
	if err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("err = %v", err)
	}
}

func TestLoadAPIKey_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(p, []byte(`not json`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadAPIKey(p)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestNew_ReportsCodexName(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(p, []byte(`{"OPENAI_API_KEY":"sk-test"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	prov, err := New(p, "gpt-4o-mini", 256)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if prov.Name() != "codex" {
		t.Errorf("Name = %q, want codex", prov.Name())
	}
}

func TestDefaultAuthFile_UsesHome(t *testing.T) {
	got, err := DefaultAuthFile()
	if err != nil {
		t.Fatalf("DefaultAuthFile: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	want := filepath.Join(home, ".codex", "auth.json")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
