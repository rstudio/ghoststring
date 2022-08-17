package ghoststring

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	namespaceSeparator = "::"
	prefix             = "👻:"

	maxNamespaceLength   = 32
	namespacePartsLength = 2
	nonceLength          = 12
	nonceLengthHex       = 24
)

var (
	Err = errors.New("ghoststring error")

	namespaceMatch = regexp.MustCompile("^[a-zA-Z][-\\._a-zA-Z0-9]*[a-zA-Z0-9]$")

	_ json.Marshaler   = &GhostString{}
	_ json.Unmarshaler = &GhostString{}
)

// GhostString wraps a string with a JSON marshaller that uses a
// namespace-scoped encrypting Ghostifyer registered via
// SetGhostifyer
type GhostString struct {
	String    string
	Namespace string
}

// IsValid checks that the wrapped string value is non-empty and
// the namespace is valid
func (gs *GhostString) IsValid() bool {
	return gs.String != "" && validateNamespace(gs.Namespace) == nil
}

// Equal compares this GhostString to another
func (gs *GhostString) Equal(other *GhostString) bool {
	return other != nil &&
		gs.String == other.String &&
		gs.Namespace == other.Namespace
}

// MarshalJSON allows GhostString to fulfill the json.Marshaler
// interface. The lack of a namespace is considered an error.
func (gs *GhostString) MarshalJSON() ([]byte, error) {
	ghostifyersLock.RLock()
	ghostifyer, ok := ghostifyers[gs.Namespace]
	ghostifyersLock.RUnlock()

	if !ok {
		ghostifyer = internalNullGhostifyer
	}

	s, err := ghostifyer.Ghostify(gs)
	if err != nil {
		return nil, err
	}

	return json.Marshal(s)
}

// UnmarshalJSON allows GhostString to fulfill the json.Unmarshaler
// interface. The bytes are first unmarshaled as a string and then
// if non-empty are passed through an "unghostify" step.
func (gs *GhostString) UnmarshalJSON(b []byte) error {
	s := ""
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" {
		gs.String = ""
		gs.Namespace = ""

		return nil
	}

	un, err := metaUnghostify(s)
	if err != nil {
		return err
	}

	gs.String = un.String
	gs.Namespace = un.Namespace

	return nil
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

	ghostifyersLock.RLock()
	ghostifyer, ok := ghostifyers[nsParts[0]]
	ghostifyersLock.RUnlock()

	if !ok {
		return nil, errors.Wrapf(Err, "no ghostifyer set for namespace %[1]q", nsParts[0])
	}

	return ghostifyer.Unghostify(s)
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
