package ghoststring

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	NamespaceMatchRegexp = "^[a-zA-Z][-\\._a-zA-Z0-9]{1,254}[a-zA-Z0-9]$"
	NamespaceSeparator   = "::"
	Prefix               = "👻:"

	Nonce = 12

	namespacePartsLength = 2
	nonceBytes           = 24 / 2
)

var (
	Err = errors.New("ghoststring error")

	namespaceMatch = regexp.MustCompile(NamespaceMatchRegexp)

	_ json.Marshaler             = &GhostString{}
	_ json.Unmarshaler           = &GhostString{}
	_ fmt.Stringer               = &GhostString{}
	_ fmt.GoStringer             = &GhostString{}
	_ encoding.TextMarshaler     = &GhostString{}
	_ encoding.TextUnmarshaler   = &GhostString{}
	_ encoding.BinaryMarshaler   = &GhostString{}
	_ encoding.BinaryUnmarshaler = &GhostString{}
)

// GhostString wraps a string with a JSON marshaller that uses a
// namespace-scoped encrypting Ghostifyer registered via
// SetGhostifyer
type GhostString struct {
	Namespace string
	Str       string
}

type unghostifyParts struct {
	nonce     []byte
	namespace string
	opaque    string
}

// IsValid checks that the wrapped string value is non-empty and
// the namespace is valid
func (gs *GhostString) IsValid() bool {
	return gs.Str != "" && validateNamespace(gs.Namespace) == nil
}

// Equal compares this GhostString to another
func (gs *GhostString) Equal(other *GhostString) bool {
	return other != nil &&
		gs.Str == other.Str &&
		gs.Namespace == other.Namespace
}

func (gs *GhostString) String() string {
	s, err := gs.toString()
	if err != nil {
		return ""
	}

	return s
}

func (gs *GhostString) GoString() string {
	return fmt.Sprintf(
		"{%q, %q}",
		gs.Namespace,
		gs.String(),
	)
}

func (gs *GhostString) toString() (string, error) {
	ghostifyersLock.RLock()
	ghostifyer, ok := ghostifyers[gs.Namespace]
	ghostifyersLock.RUnlock()

	if !ok {
		ghostifyer = internalNullGhostifyer
	}

	s, err := ghostifyer.Ghostify(gs)
	if err != nil {
		return "", err
	}

	return s, nil
}

// MarshalJSON allows GhostString to fulfill the json.Marshaler
// interface. The lack of a namespace is considered an error.
func (gs *GhostString) MarshalJSON() ([]byte, error) {
	s, err := gs.toString()
	if err != nil {
		return nil, err
	}

	return json.Marshal(s)
}

// UnmarshalJSON allows GhostString to fulfill the json.Unmarshaler
// interface. The bytes are first unmarshaled as a string and then
// if non-empty are passed through an "unghostify" step. The
// expected structure of a marshalled GhostString is:
//
//	  "{Prefix}base64({nonce}{namespace}{NamespaceSeparator}{opaque-value})"
//
// where {nonce} has the length specified as Nonce.
func (gs *GhostString) UnmarshalJSON(b []byte) error {
	s := ""
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" {
		gs.Str = ""
		gs.Namespace = ""

		return nil
	}

	un, err := metaUnghostify(s)
	if err != nil {
		return err
	}

	gs.Str = un.Str
	gs.Namespace = un.Namespace

	return nil
}

// MarshalText allows GhostString to fulfill the encoding.TextMarshaler interface
func (gs *GhostString) MarshalText() ([]byte, error) {
	s, err := gs.toString()
	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

// UnmarshalText allows GhostString to fulfill the encoding.TextUnmarshaler interface
func (gs *GhostString) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		gs.Str = ""
		gs.Namespace = ""

		return nil
	}

	un, err := metaUnghostify(string(b))
	if err != nil {
		return err
	}

	gs.Str = un.Str
	gs.Namespace = un.Namespace

	return nil
}

// MarshalBinary allows GhostString to fulfill the encoding.BinaryMarshaler interface
func (gs *GhostString) MarshalBinary() ([]byte, error) {
	return gs.MarshalText()
}

// UnmarshalBinary allows GhostString to fulfill the encoding.BinaryUnmarshaler interface
func (gs *GhostString) UnmarshalBinary(b []byte) error {
	return gs.UnmarshalText(b)
}

func toUnghostifyParts(s string) (*unghostifyParts, error) {
	nonceNsValueBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, Prefix))
	if err != nil {
		return nil, err
	}

	nonce, nsValueBytes := nonceNsValueBytes[:Nonce], nonceNsValueBytes[Nonce:]

	nsParts := strings.SplitN(string(nsValueBytes), NamespaceSeparator, namespacePartsLength)
	if len(nsParts) != namespacePartsLength {
		return nil, errors.Wrap(Err, "invalid namespacing")
	}

	return &unghostifyParts{
		nonce:     nonce,
		namespace: nsParts[0],
		opaque:    nsParts[1],
	}, nil
}

func metaUnghostify(s string) (*GhostString, error) {
	unParts, err := toUnghostifyParts(s)
	if err != nil {
		return nil, err
	}

	ghostifyersLock.RLock()
	ghostifyer, ok := ghostifyers[unParts.namespace]
	ghostifyersLock.RUnlock()

	if !ok {
		return nil, errors.Wrapf(Err, "no ghostifyer set for namespace %[1]q", unParts.namespace)
	}

	return ghostifyer.Unghostify(s)
}

func validateNamespace(namespace string) error {
	if namespace != strings.TrimSpace(namespace) {
		return errors.Wrapf(Err, "invalid namespace with blankspace %[1]q", namespace)
	}

	if !namespaceMatch.MatchString(namespace) {
		return errors.Wrapf(Err, "invalid namespace %[1]q", namespace)
	}

	return nil
}
