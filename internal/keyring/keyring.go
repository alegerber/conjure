package keyring

import "errors"

const (
	// Service is the legacy default keyring service. New code should pass an
	// explicit service name to ResolveFor / NewSystemStoreFor.
	Service = "anthropic-api-key"
	// EnvVar is the legacy default env-var fallback.
	EnvVar = "ANTHROPIC_API_KEY"
)

var ErrNotFound = errors.New("key not found")

type Store interface {
	Get() (string, error)
	Set(key string) error
	Source() string // human-readable origin, used in error messages
}

// Resolve is preserved for back-compat and resolves the legacy
// anthropic-api-key service / ANTHROPIC_API_KEY env var.
func Resolve() (Store, string, error) {
	return ResolveFor(Service, EnvVar)
}

// ResolveFor returns the first store that yields a key for the given keyring
// service / env-var pair. It tries the system keyring first, then falls back
// to the env var. The returned Store reports where the key came from via
// Source(). If neither has a key, ErrNotFound.
//
// A system-keyring error (e.g. D-Bus unavailable on headless Linux) is
// treated as a miss and silently triggers the env-var fallback. This is
// intentional: a misconfigured keyring daemon must not block users who have
// the env var set.
func ResolveFor(service, envVar string) (Store, string, error) {
	if sys, err := NewSystemStoreFor(service); err == nil {
		if v, err := sys.Get(); err == nil {
			return sys, v, nil
		}
	}
	env := NewEnvStoreFor(envVar)
	if v, err := env.Get(); err == nil {
		return env, v, nil
	}
	return nil, "", ErrNotFound
}
