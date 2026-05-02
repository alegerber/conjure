package keyring

import (
	"errors"
	"os"
)

type envStore struct{ envVar string }

// NewEnvStore returns an env-var store bound to the legacy ANTHROPIC_API_KEY
// variable. Prefer NewEnvStoreFor for new code.
func NewEnvStore() Store { return NewEnvStoreFor(EnvVar) }

// NewEnvStoreFor returns an env-var store bound to the given variable name.
func NewEnvStoreFor(envVar string) Store { return envStore{envVar: envVar} }

func (s envStore) Get() (string, error) {
	v := os.Getenv(s.envVar)
	if v == "" {
		return "", ErrNotFound
	}
	return v, nil
}

func (envStore) Set(string) error {
	return errors.New("env-var store is read-only; use system keyring (run: conjure setup)")
}

func (s envStore) Source() string { return s.envVar + " env var" }
