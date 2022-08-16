package ghoststring_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

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
			g:   &ghoststring.GhostString{},
			s:   "null",
			err: ghoststring.Err,
		},
		{
			g: &ghoststring.GhostString{Namespace: "test"},
			s: "null",
		},
		{
			g: &ghoststring.GhostString{Namespace: "test", String: "maybe"},
			s: "\"👻:dGVzdDo6N2KgwoGvJ0/dNzpBtEzRQxX1/DsS\"",
		},
	} {
		t.Run(fmt.Sprintf("namespace=%[1]q,string=%[2]v", tc.g.Namespace, tc.g.String), func(t *testing.T) {
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

func ExampleGhostString() {
	if err := ghoststring.SetGhostifyer(
		"example",
		"correct horse battery staple",
		[]byte("arghhhhhhhhh"),
	); err != nil {
		panic(err)
	}

	type DiaryEntry struct {
		Timestamp time.Time               `json:"timestamp"`
		Text      ghoststring.GhostString `json:"text"`
	}

	type Diary struct {
		Author  string       `json:"author"`
		Entries []DiaryEntry `json:"entries"`
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(
		&Diary{
			Author: "Eagerly Anticipated",
			Entries: []DiaryEntry{
				{
					Timestamp: time.UnixMicro(4),
					Text: ghoststring.GhostString{
						Namespace: "example",
						String:    "Nights without you are so dark. I pray that someday you will return my flashlight.",
					},
				},
			},
		},
	); err != nil {
		panic(err)
	}

	// Output: {
	//   "author": "Eagerly Anticipated",
	//   "entries": [
	//     {
	//       "timestamp": "1969-12-31T19:00:00.000004-05:00",
	//       "text": "👻:ZXhhbXBsZTo6c+wSRuV0cACU48++XPBIVPyvnMGQLJA0Gak3osgO0EHHNTIU8HO67H7T3/XgGZsXm9XdCqYkip0H/7L8q20MUUmiod2ykXZiW0NSNkL7f7JwCV4GcqGbhmIGaEjzgCKtE2A="
	//     }
	//   ]
	// }
	//
}
