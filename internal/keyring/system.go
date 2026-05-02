package keyring

import (
	"errors"
	"os/user"

	gkeyring "github.com/zalando/go-keyring"
)

type systemStore struct {
	user    string
	service string
}

var _ Store = systemStore{}

// NewSystemStore returns a system-keyring store bound to the legacy
// anthropic-api-key service. Prefer NewSystemStoreFor for new code.
func NewSystemStore() (Store, error) {
	return NewSystemStoreFor(Service)
}

// NewSystemStoreFor returns a system-keyring store bound to the given service.
func NewSystemStoreFor(service string) (Store, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	return systemStore{user: u.Username, service: service}, nil
}

func (s systemStore) Get() (string, error) {
	v, err := gkeyring.Get(s.service, s.user)
	if err != nil {
		if errors.Is(err, gkeyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return v, nil
}

func (s systemStore) Set(key string) error {
	return gkeyring.Set(s.service, s.user, key)
}

func (s systemStore) Source() string { return "system keyring (service=" + s.service + ")" }
