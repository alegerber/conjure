package keyring

import (
	"errors"
	"os/user"

	gkeyring "github.com/zalando/go-keyring"
)

type systemStore struct{ user string }

func NewSystemStore() (Store, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	return systemStore{user: u.Username}, nil
}

func (s systemStore) Get() (string, error) {
	v, err := gkeyring.Get(Service, s.user)
	if err != nil {
		if errors.Is(err, gkeyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return v, nil
}

func (s systemStore) Set(key string) error {
	return gkeyring.Set(Service, s.user, key)
}

func (systemStore) Source() string { return "system keyring (service=" + Service + ")" }
