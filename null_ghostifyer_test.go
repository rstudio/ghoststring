package ghoststring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullGhostifyer(t *testing.T) {
	r := require.New(t)

	ng := &nullGhostifyer{}

	r.Implements((*Ghostifyer)(nil), ng)
	r.Equal("", ng.Namespace())

	s, err := ng.Ghostify(&GhostString{})
	r.Equal("", s)
	r.Nil(err)

	gs, err := ng.Unghostify("")
	r.Equal(gs, &GhostString{})
	r.Nil(err)
}
