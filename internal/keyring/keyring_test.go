package keyring

import (
	"errors"
	"os"
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
	os.Unsetenv(EnvVar)
	_, err := NewEnvStore().Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestEnvStore_SetReturnsError(t *testing.T) {
	if err := NewEnvStore().Set("anything"); err == nil {
		t.Errorf("Set should return error for read-only store")
	}
}
