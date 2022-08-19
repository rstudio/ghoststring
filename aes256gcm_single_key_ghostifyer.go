package ghoststring

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

const (
	saltPrefix = "github.com/rstudio/ghoststring:"

	aesKeyLen     = 32
	argon2Mem     = 64 * 1024
	argon2Threads = 4
	argon2Time    = 1
)

// NewAES256GCMSingleKeyGhostifyer creates a Ghostifyer with a
// single key that uses AES-256-GCM encryption with nonce assigned
// at the individual string level.
func NewAES256GCMSingleKeyGhostifyer(namespace, key string) (Ghostifyer, error) {
	if err := validateNamespace(namespace); err != nil {
		return nil, err
	}

	salt := fmt.Sprintf(
		"%x",
		sha1.Sum(append([]byte(saltPrefix), []byte(namespace)...)),
	)

	dk := argon2.IDKey(
		[]byte(key),
		[]byte(salt),
		argon2Time,
		argon2Mem,
		argon2Threads,
		aesKeyLen,
	)

	return &aes256GcmSingleKeyGhostifyer{
		ns:  namespace,
		key: []byte(dk),
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

	nonce := make([]byte, nonceBytes)
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
