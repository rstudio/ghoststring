package ghoststring

import (
	"crypto/sha1"
	"fmt"
	"regexp"
	"sync"

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
	ghostifyersLock = &sync.RWMutex{}

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

func SetGhostifyer(namespace, key string) (Ghostifyer, error) {
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

	gh := &aes256GcmGhostifyer{key: []byte(dk)}

	ghostifyersLock.Lock()
	ghostifyers[namespace] = gh
	ghostifyersLock.Unlock()

	return gh, nil
}
