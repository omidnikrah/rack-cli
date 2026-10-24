package secret

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	KeychainService = "rack-cli"
	KeychainAccount = "vault-key"
	KeySize         = 32
)

type KeySource interface {
	Key() ([]byte, error)
}

type Keychain struct {
}

func (Keychain) Key() ([]byte, error) {
	encoded, err := keyring.Get(KeychainService, KeychainAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return createKey()
	}

	if err != nil {
		return nil, fmt.Errorf("reading key from keychain: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != KeySize {
		return nil, errors.New("keychain key is corrupted")
	}

	return key, nil
}

func createKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}

	if err := keyring.Set(KeychainService, KeychainAccount, base64.StdEncoding.EncodeToString(key)); err != nil {
		return nil, fmt.Errorf("saving key to keychain: %w", err)
	}

	return key, nil
}
