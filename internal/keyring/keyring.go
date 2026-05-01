package keyring

import "errors"

const (
	Service = "anthropic-api-key"
	EnvVar  = "ANTHROPIC_API_KEY"
)

var ErrNotFound = errors.New("key not found")

type Store interface {
	Get() (string, error)
	Set(key string) error
	Source() string // human-readable origin, used in error messages
}

// Resolve returns the first store that yields a key. It tries the system
// keyring first, then falls back to the env var. The returned Store reports
// where the key came from via Source(). If neither has a key, ErrNotFound.
//
// A system-keyring error (e.g. D-Bus unavailable on headless Linux) is
// treated as a miss and silently triggers the env-var fallback. This is
// intentional: a misconfigured keyring daemon must not block users who have
// set ANTHROPIC_API_KEY.
func Resolve() (Store, string, error) {
	if sys, err := NewSystemStore(); err == nil {
		if v, err := sys.Get(); err == nil {
			return sys, v, nil
		}
	}
	env := NewEnvStore()
	if v, err := env.Get(); err == nil {
		return env, v, nil
	}
	return nil, "", ErrNotFound
}
