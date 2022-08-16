package ghoststring_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rstudio/ghoststring"
	"github.com/stretchr/testify/require"
)

func TestGhostString_JSONRoundTrip(t *testing.T) {
	require.Nil(
		t,
		ghoststring.SetGhostifyer(
			"test",
			"hang-glider-casserole-newt",
			[]byte("4555679050a3"),
		),
	)

	for _, tc := range []struct {
		g   *ghoststring.GhostString
		s   string
		err error
	}{
		{
			g: &ghoststring.GhostString{Namespace: "test"},
			s: "null",
		},
		{
			g: &ghoststring.GhostString{Namespace: "test", String: "nope"},
			s: "null",
		},
		{
			g: &ghoststring.GhostString{Namespace: "test", Valid: true},
			s: "\"👻:dGVzdDo6vCuEIr0ZyfY3+RMNIeuYew==\"",
		},
		{
			g: &ghoststring.GhostString{Namespace: "test", String: "maybe", Valid: true},
			s: "\"👻:dGVzdDo6VTJUrg2wETXLUYVVAktVU0yz65bZ\"",
		},
	} {
		t.Run(fmt.Sprintf("string=%[1]q,valid=%[2]v", tc.g.String, tc.g.Valid), func(t *testing.T) {
			r := require.New(t)

			actualBytes, err := json.Marshal(tc.g)
			if err != nil {
				r.ErrorIs(err, tc.err)
				return
			}

			r.Equal(tc.s, string(actualBytes))

			if string(actualBytes) == "null" {
				return
			}

			fromJson := &ghoststring.GhostString{}
			if err := json.Unmarshal(actualBytes, fromJson); err != nil {
				r.Nilf(err, "non-nil error %+#[1]v", err)
			}

			r.Truef(tc.g.Equal(fromJson), "%+#[1]v != %+#[2]v", tc.g, fromJson)
		})
	}
}
