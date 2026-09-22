package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"sync"
)

const prefix = "v1:"

var ErrDecrypt = errors.New("cannot decrypt password")

type Sealed string

func (s Sealed) IsSealed() bool {
	return strings.HasPrefix(string(s), prefix)
}

type Vault struct {
	aead func() (cipher.AEAD, error)
}

func NewVault(keys KeySource) *Vault {
	return &Vault{
		aead: sync.OnceValues(func() (cipher.AEAD, error) {
			key, err := keys.Key()
			if err != nil {
				return nil, err
			}
			block, err := aes.NewCipher(key)
			if err != nil {
				return nil, err
			}
			return cipher.NewGCMWithRandomNonce(block)
		}),
	}
}

func (v *Vault) Seal(plain string, id int) (Sealed, error) {
	if plain == "" {
		return "", nil
	}

	aead, err := v.aead()
	if err != nil {
		return "", err
	}

	out := aead.Seal(nil, nil, []byte(plain), additionalData(id))
	return Sealed(prefix + base64.StdEncoding.EncodeToString(out)), nil
}

func (v *Vault) Open(sealed Sealed, id int) (string, error) {
	if sealed == "" {
		return "", nil
	}

	encoded, ok := strings.CutPrefix(string(sealed), prefix)
	if !ok {
		return "", ErrDecrypt
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrDecrypt
	}

	aead, err := v.aead()
	if err != nil {
		return "", err
	}

	plain, err := aead.Open(nil, nil, raw, additionalData(id))
	if err != nil {
		return "", ErrDecrypt
	}

	return string(plain), nil
}

func additionalData(id int) []byte {
	return []byte("rack-cli/server/" + strconv.Itoa(id))
}
