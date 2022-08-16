package ghoststring

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strings"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"golang.org/x/crypto/scrypt"
)

const (
	prefix               = "👻:"
	namespaceSeparator   = "::"
	namespacePartsLength = 2
	maxNamespaceLength   = 32

	aesKeyLen = 32
	aesRecN   = 32_768
	aesRecr   = 8
	aesRecp   = 1

	nonceLength = 12
)

var (
	ghostifyers     = map[string]Ghostifyer{}
	ghostifyersLock = &sync.Mutex{}

	namespaceMatch = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]*$")
)

// Ghostifyer encrypts and encodes a *GhostString into a string representation that is
// acceptable for inclusion in JSON. The structure of a ghostified string is:
//
//	  {prefix}base64({namespace}{namespace separator}{value})
type Ghostifyer interface {
	Ghostify(*GhostString) (string, error)
	Unghostify(string) (*GhostString, error)
}

func metaUnghostify(s string) (*GhostString, error) {
	nsValueBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, prefix))
	if err != nil {
		return nil, err
	}

	nsParts := strings.SplitN(string(nsValueBytes), namespaceSeparator, namespacePartsLength)
	if len(nsParts) != namespacePartsLength {
		return nil, errors.Wrap(Err, "invalid namespacing")
	}

	ghostifyer, ok := ghostifyers[nsParts[0]]
	if !ok {
		return nil, errors.Wrapf(Err, "no ghostifyer set for namespace %[1]q", nsParts[0])
	}

	return ghostifyer.Unghostify(s)
}

func SetGhostifyer(namespace, key string, nonce []byte) error {
	ghostifyersLock.Lock()
	defer ghostifyersLock.Unlock()

	if err := validateNamespace(namespace); err != nil {
		return err
	}

	if len(nonce) != nonceLength {
		return errors.Wrap(Err, "invalid nonce length")
	}

	dk, err := scrypt.Key([]byte(key), nonce, aesRecN, aesRecr, aesRecp, aesKeyLen)
	if err != nil {
		return err
	}

	ghostifyers[namespace] = &aes256GcmGhostifyer{key: []byte(dk), nonce: nonce}

	return nil
}

func SetGhostifyerFromEnv(namespace string) error {
	ghostifyersLock.Lock()
	defer ghostifyersLock.Unlock()

	if err := validateNamespace(namespace); err != nil {
		return err
	}

	cfg, err := newConfig(namespace)
	if err != nil {
		return err
	}

	decNonce, err := hex.DecodeString(cfg.Nonce)
	if err != nil {
		return err
	}

	if len(decNonce) != nonceLength {
		return errors.Wrap(Err, "invalid nonce length")
	}

	dk, err := scrypt.Key([]byte(cfg.Key), decNonce, aesRecN, aesRecr, aesRecp, aesKeyLen)
	if err != nil {
		return err
	}

	ghostifyers[namespace] = &aes256GcmGhostifyer{key: []byte(dk), nonce: decNonce}

	return nil
}

type config struct {
	Key   string `envconfig:"KEY"`
	Nonce string `envconfig:"NONCE"`
}

func newConfig(namespace string) (*config, error) {
	cfg := &config{}

	if err := envconfig.Process("ghoststring_"+namespace, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateNamespace(namespace string) error {
	if namespace != strings.TrimSpace(namespace) {
		return errors.Wrapf(Err, "invalid namespace with blankspace %[1]q", namespace)
	}

	if len(namespace) > maxNamespaceLength {
		return errors.Wrapf(Err, "invalid namespace is too long %[1]q", namespace)
	}

	if !namespaceMatch.MatchString(namespace) {
		return errors.Wrapf(Err, "invalid namespace %[1]q", namespace)
	}

	return nil
}
