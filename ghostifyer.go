package ghoststring

import (
	"sync"
)

var (
	internalNullGhostifyer Ghostifyer = &nullGhostifyer{}

	ghostifyers     = map[string]Ghostifyer{}
	ghostifyersLock = &sync.RWMutex{}
)

// Ghostifyer encodes a GhostString into a string representation.
type Ghostifyer interface {
	Namespace() string
	Ghostify(*GhostString) (string, error)
	Unghostify(string) (*GhostString, error)
}

func SetGhostifyer(gh Ghostifyer) error {
	namespace := gh.Namespace()

	if err := validateNamespace(namespace); err != nil {
		return err
	}

	ghostifyersLock.Lock()
	ghostifyers[namespace] = gh
	ghostifyersLock.Unlock()

	return nil
}
