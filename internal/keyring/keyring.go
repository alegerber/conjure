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
