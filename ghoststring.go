package ghoststring

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

const (
	jsonNull = "null"
)

var (
	Err = errors.New("ghoststring error")

	_ json.Marshaler   = &GhostString{}
	_ json.Unmarshaler = &GhostString{}
)

type GhostString struct {
	String    string
	Namespace string
}

func (gs *GhostString) IsValid() bool {
	return gs.String != "" && validateNamespace(gs.Namespace) == nil
}

func (gs *GhostString) Equal(other *GhostString) bool {
	return other != nil &&
		gs.String == other.String &&
		gs.Namespace == other.Namespace
}

func (gs *GhostString) MarshalJSON() ([]byte, error) {
	if gs.Namespace == "" {
		return nil, errors.Wrap(Err, "no namespace set")
	}

	ghostifyersLock.RLock()
	ghostifyer, ok := ghostifyers[gs.Namespace]
	ghostifyersLock.RUnlock()

	if !ok {
		return nil, errors.Wrapf(Err, "no ghostifyer set for namespace %[1]q", gs.Namespace)
	}

	s, err := ghostifyer.Ghostify(gs)
	if err != nil {
		return nil, err
	}

	if s == jsonNull {
		return []byte(s), nil
	}

	return json.Marshal(s)
}

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
