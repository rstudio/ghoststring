package ghoststring_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rstudio/ghoststring"
	"github.com/stretchr/testify/require"
)

func TestAES256GCMultiKeyGhostifyer(t *testing.T) {
	r := require.New(t)

	keyA := "correct horse battery staple"
	keyB := "grim cereal tardy octopus"

	ks, err := ghoststring.NewKeyStoreFromEnv(
		"test.local",
		[]string{
			fmt.Sprintf(`GHOSTSTRING_KEY_TEST_LOCAL_A={"key":"%s","timestamp":1661351759000}`, keyA),
			fmt.Sprintf(`GHOSTSTRING_KEY_TEST_LOCAL_B={"key":"%s","timestamp":1661351742000}`, keyB),
		},
	)
	r.Nil(err)

	gh := ghoststring.NewAES256GCMMultiKeyGhostifyer("test.local", ks)
	r.Nil(ghoststring.SetGhostifyer(gh))

	type tree struct {
		Phylogeny      int                     `json:"phylogeny"`
		HatShape       ghoststring.GhostString `json:"hat_shape"`
		SecretIdentity ghoststring.GhostString `json:"secret_identity"`
	}

	for _, tc := range []struct {
		name   string
		input  any
		verify func(*require.Assertions, []byte, error)
	}{
		{
			name: "typical",
			input: &tree{
				Phylogeny: 42,
				HatShape: ghoststring.GhostString{
					Namespace: "test.local",
					Str:       "pointy",
				},
				SecretIdentity: ghoststring.GhostString{
					Namespace: "test.local",
					Str:       "daisy",
				},
			},
			verify: func(r *require.Assertions, b []byte, err error) {
				r.NotNil(b)
				r.Nil(err)

				s := string(b)

				r.Contains(s, `"phylogeny":42`)
				r.Contains(s, fmt.Sprintf(`"secret_identity":"%s`, ghoststring.Prefix))
			},
		},
		{
			name: "different keys",
			input: &tree{
				Phylogeny: 11,
				HatShape: ghoststring.GhostString{
					Namespace: "NA",
					Str:       "frilly",
				},
				SecretIdentity: ghoststring.GhostString{
					Namespace: "test.local",
					Str:       "prototaxites",
				},
			},
			verify: func(r *require.Assertions, b []byte, err error) {
				r.NotNil(b)
				r.Nil(err)

				naiveMap := map[string]any{}
				r.Nil(json.Unmarshal(b, &naiveMap))

				gh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer("test.local", keyB)
				r.NotNil(gh)
				r.Nil(err)

				gs, err := gh.Unghostify(naiveMap["secret_identity"].(string))
				r.Nil(gs)
				r.NotNil(err)
			},
		},
	} {
		t.Run("json marshal "+tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			tc.verify(require.New(t), b, err)
		})
	}

	for _, tc := range []struct {
		name   string
		encStr string
		verify func(*require.Assertions, *ghoststring.GhostString, error)
	}{
		{
			name: "typical",
			encStr: func() string {
				s, err := gh.Ghostify(
					&ghoststring.GhostString{
						Namespace: "test.local",
						Str:       "pointy",
					},
				)
				r.Nil(err)
				return s
			}(),
			verify: func(r *require.Assertions, gs *ghoststring.GhostString, err error) {
				r.NotNil(gs)
				r.Nil(err)
				r.Equal("pointy", gs.Str)
			},
		},
		{
			name: "different keys",
			encStr: func() string {
				bgh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer("test.local", keyB)
				r.Nil(err)

				s, err := bgh.Ghostify(
					&ghoststring.GhostString{
						Namespace: "test.local",
						Str:       "prototaxites",
					},
				)
				r.Nil(err)
				return s
			}(),
			verify: func(r *require.Assertions, gs *ghoststring.GhostString, err error) {
				r.NotNil(gs)
				r.Nil(err)
				r.Equal("prototaxites", gs.Str)
			},
		},
	} {
		t.Run("unghostify "+tc.name, func(t *testing.T) {
			gs, err := gh.Unghostify(tc.encStr)
			tc.verify(require.New(t), gs, err)
		})
	}
}
