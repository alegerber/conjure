package keyring

import (
	"errors"
	"os"
)

type envStore struct{}

func NewEnvStore() Store { return envStore{} }

func (envStore) Get() (string, error) {
	v := os.Getenv(EnvVar)
	if v == "" {
		return "", ErrNotFound
	}
	return v, nil
}

func (envStore) Set(string) error {
	return errors.New("env-var store is read-only; use system keyring (run: conjure setup)")
}

func (envStore) Source() string { return EnvVar + " env var" }
