package ghoststring

import (
	"crypto/sha1"
	"fmt"
	"regexp"
	"sync"

	"golang.org/x/crypto/argon2"
)

const (
	namespaceSeparator = "::"
	prefix             = "👻:"
	saltPrefix         = "github.com/rstudio/ghoststring:"

	aesKeyLen            = 32
	argon2Time           = 1
	argon2Mem            = 64 * 1024
	argon2Threads        = 4
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

	dk := argon2.IDKey(
		[]byte(key),
		[]byte(salt),
		argon2Time,
		argon2Mem,
		argon2Threads,
		aesKeyLen,
	)

	gh := &aes256GcmGhostifyer{key: []byte(dk)}

	ghostifyersLock.Lock()
	ghostifyers[namespace] = gh
	ghostifyersLock.Unlock()

	return gh, nil
}
