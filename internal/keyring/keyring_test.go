package keyring

import (
	"errors"
	"strings"
	"testing"
)

func TestEnvStore_GetReturnsValue(t *testing.T) {
	t.Setenv(EnvVar, "sk-test-123")
	got, err := NewEnvStore().Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "sk-test-123" {
		t.Errorf("got %q", got)
	}
}

func TestEnvStore_GetMissingReturnsErrNotFound(t *testing.T) {
	t.Setenv(EnvVar, "")
	_, err := NewEnvStore().Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestEnvStore_SetReturnsError(t *testing.T) {
	err := NewEnvStore().Set("anything")
	if err == nil {
		t.Fatal("Set should return error for read-only store")
	}
	if !strings.Contains(err.Error(), "conjure setup") {
		t.Errorf("Set error should mention 'conjure setup', got: %v", err)
	}
}

func TestEnvStore_Source(t *testing.T) {
	if got := NewEnvStore().Source(); got != EnvVar+" env var" {
		t.Errorf("Source = %q, want %q", got, EnvVar+" env var")
	}
}
