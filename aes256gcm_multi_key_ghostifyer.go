package ghoststring

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/pkg/errors"
)

// NewAES256GCMMultiKeyGhostifyer creates a Ghostifyer with
// multiple timestamped keys that uses AES-256-GCM encryption with
// nonce assigned at the individual string level. The
// keystore.Latest will be used for encryption and any key in
// keystore.All may be used for decryption.
func NewAES256GCMMultiKeyGhostifyer(namespace string, keys KeyStore) Ghostifyer {
	return &aes256GcmMultiKeyGhostifyer{
		ns:   namespace,
		keys: keys,
	}
}

type aes256GcmMultiKeyGhostifyer struct {
	ns   string
	keys KeyStore
}

func (g *aes256GcmMultiKeyGhostifyer) Namespace() string { return g.ns }

func (g *aes256GcmMultiKeyGhostifyer) Ghostify(gs *GhostString) (string, error) {
	encKey, err := g.keys.Latest(context.TODO())
	if err != nil {
		return "", err
	}

	if gs == nil || !gs.IsValid() {
		return "", nil
	}

	nonce := make([]byte, Nonce)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	encBytes, err := aes256GcmEncrypt(encKey, nonce, gs.Str)
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

func (g *aes256GcmMultiKeyGhostifyer) Unghostify(s string) (*GhostString, error) {
	if s == Prefix || s == "" {
		return &GhostString{}, nil
	}

	allKeys, err := g.keys.All(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, kb := range allKeys {
		if gs, err := g.tryUnghostify(kb, s); err == nil {
			return gs, nil
		}
	}

	return nil, errors.Wrap(Err, "no valid decryption key")
}

func (g *aes256GcmMultiKeyGhostifyer) tryUnghostify(kb []byte, s string) (*GhostString, error) {
	unParts, err := toUnghostifyParts(s)
	if err != nil {
		return nil, err
	}

	plainBytes, err := aes256GcmDecrypt(kb, unParts.nonce, unParts.opaque)
	if err != nil {
		return nil, err
	}

	return &GhostString{Namespace: unParts.namespace, Str: string(plainBytes)}, nil
}
