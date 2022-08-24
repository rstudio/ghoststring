package ghoststring

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
)

// NewAES256GCMSingleKeyGhostifyer creates a Ghostifyer with a
// single key that uses AES-256-GCM encryption with nonce assigned
// at the individual string level.
func NewAES256GCMSingleKeyGhostifyer(namespace, key string) (Ghostifyer, error) {
	keyBytes, err := newAES256GCMKey(namespace, key)
	if err != nil {
		return nil, err
	}

	return &aes256GcmSingleKeyGhostifyer{
		ns:  namespace,
		key: keyBytes,
	}, nil
}

type aes256GcmSingleKeyGhostifyer struct {
	ns  string
	key []byte
}

func (g *aes256GcmSingleKeyGhostifyer) Namespace() string { return g.ns }

func (g *aes256GcmSingleKeyGhostifyer) Ghostify(gs *GhostString) (string, error) {
	if strings.TrimSpace(string(g.key)) == "" {
		return "", errors.Wrap(Err, "invalid key")
	}

	if gs == nil || !gs.IsValid() {
		return "", nil
	}

	nonce := make([]byte, Nonce)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	encBytes, err := aes256GcmEncrypt(g.key, nonce, gs.Str)
	if err != nil {
		return "", err
	}

	b64Value := base64.StdEncoding.EncodeToString(
		append(
			append(
				nonce,
				[]byte(gs.Namespace+NamespaceSeparator)...,
			),
			encBytes...,
		),
	)

	return Prefix + b64Value, nil
}

func (g *aes256GcmSingleKeyGhostifyer) Unghostify(s string) (*GhostString, error) {
	if strings.TrimSpace(string(g.key)) == "" {
		return nil, errors.Wrap(Err, "invalid key")
	}

	if s == Prefix || s == "" {
		return &GhostString{}, nil
	}

	unParts, err := toUnghostifyParts(s)
	if err != nil {
		return nil, err
	}

	plainBytes, err := aes256GcmDecrypt(g.key, unParts.nonce, unParts.opaque)
	if err != nil {
		return nil, err
	}

	return &GhostString{Namespace: unParts.namespace, Str: string(plainBytes)}, nil
}
