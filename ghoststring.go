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
	Valid     bool
	String    string
	Namespace string
}

func (gs *GhostString) Equal(other *GhostString) bool {
	return other != nil &&
		gs.Valid == other.Valid &&
		gs.String == other.String &&
		gs.Namespace == other.Namespace
}

func (gs *GhostString) MarshalJSON() ([]byte, error) {
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
		gs.Valid = false
		gs.String = ""
		gs.Namespace = ""

		return nil
	}

	un, err := metaUnghostify(s)
	if err != nil {
		return err
	}

	gs.Valid = un.Valid
	gs.String = un.String
	gs.Namespace = un.Namespace

	return nil
}
