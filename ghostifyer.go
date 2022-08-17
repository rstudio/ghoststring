package ghoststring

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

const (
	namespaceSeparator = "::"
	prefix             = "👻:"
	saltPrefix         = "github.com/rstudio/ghoststring:"

	aesKeyLen            = 32
	aesRecN              = 32_768
	maxNamespaceLength   = 32
	namespacePartsLength = 2
	nonceLength          = 12
	nonceLengthHex       = 24
)

var (
	ghostifyers     = map[string]Ghostifyer{}
	ghostifyersLock = &sync.Mutex{}

	namespaceMatch = regexp.MustCompile("^[a-zA-Z][-\\._a-zA-Z0-9]*[a-zA-Z0-9]$")
)

// Ghostifyer encrypts and encodes a *GhostString into a string representation that is
// acceptable for inclusion in JSON. The structure of a ghostified string is:
//
//	  {prefix}base64({nonce}{namespace}{namespace separator}{value})
type Ghostifyer interface {
	Ghostify(*GhostString) (string, error)
	Unghostify(string) (*GhostString, error)
}

func metaUnghostify(s string) (*GhostString, error) {
	nonceNsValueBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, prefix))
	if err != nil {
		return nil, err
	}

	nsValueBytes := nonceNsValueBytes[nonceLengthHex:]

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

func SetGhostifyer(namespace, key string) (Ghostifyer, error) {
	ghostifyersLock.Lock()
	defer ghostifyersLock.Unlock()

	if err := validateNamespace(namespace); err != nil {
		return nil, err
	}

	salt := fmt.Sprintf(
		"%x",
		sha1.Sum(append([]byte(saltPrefix), []byte(namespace)...)),
	)

	dk := pbkdf2.Key(
		[]byte(key),
		[]byte(salt),
		aesRecN,
		aesKeyLen,
		sha1.New,
	)

	ghostifyers[namespace] = &aes256GcmGhostifyer{key: []byte(dk)}

	return ghostifyers[namespace], nil
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
