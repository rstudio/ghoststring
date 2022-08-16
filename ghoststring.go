package ghoststring

import (
	"encoding/json"

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

	ghostifyer, ok := ghostifyers[gs.Namespace]
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
