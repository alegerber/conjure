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

func TestResolve_ReturnsFromEnv_WhenSystemMisses(t *testing.T) {
	// We can't reliably test the system-keyring branch in unit tests (requires
	// platform daemon). This test covers the env-fallback path: assumes the
	// system keyring has no entry for the test user (true in CI, true on a
	// fresh dev machine), then sets the env var and asserts Resolve picks it up.
	t.Setenv(EnvVar, "sk-test-resolve-456")
	store, key, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// The store could be either env or system depending on what's in the
	// developer's actual keyring. Assert on the key value, which differs.
	if key != "sk-test-resolve-456" {
		// Don't fail — the developer might have a real key in Keychain that
		// takes precedence. Skip in that case.
		t.Skipf("system keyring returned %q (probably has a real entry); env-fallback path not exercised", key)
	}
	if store.Source() != EnvVar+" env var" {
		t.Errorf("Source = %q, want env-var", store.Source())
	}
}

func TestResolve_BothMissReturnsErrNotFound(t *testing.T) {
	// Same caveat: if the developer has a real Keychain entry, this test will
	// get a successful Resolve. Skip in that case.
	t.Setenv(EnvVar, "")
	_, _, err := Resolve()
	if err == nil {
		t.Skip("system keyring has an entry for this user; can't test the both-miss path here")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
