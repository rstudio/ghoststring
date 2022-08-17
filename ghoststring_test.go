package ghoststring_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/rstudio/ghoststring"
	"github.com/stretchr/testify/require"
)

func TestGhostString_JSONRoundTrip(t *testing.T) {
	_, err := ghoststring.SetGhostifyer(
		"test",
		"hang-glider-casserole-newt",
	)
	require.Nil(t, err)

	for _, tc := range []struct {
		g    *ghoststring.GhostString
		null bool
		err  error
	}{
		{
			g:    &ghoststring.GhostString{},
			null: true,
			err:  ghoststring.Err,
		},
		{
			g:    &ghoststring.GhostString{Namespace: "test"},
			null: true,
		},
		{
			g: &ghoststring.GhostString{Namespace: "test", String: "maybe"},
		},
	} {
		t.Run(fmt.Sprintf("namespace=%[1]q,string=%[2]v", tc.g.Namespace, tc.g.String), func(t *testing.T) {
			r := require.New(t)

			actualBytes, err := json.Marshal(tc.g)
			if err != nil {
				r.ErrorIs(err, tc.err)
				return
			}

			r.NotEqual("", string(actualBytes))

			if tc.null {
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
	if _, err := ghoststring.SetGhostifyer(
		"example",
		"correct horse battery staple",
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

	enc, err := json.MarshalIndent(
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
				{
					Timestamp: time.UnixMicro(-8001),
					Text: ghoststring.GhostString{
						Namespace: "unknown",
						String:    "We may never know.",
					},
				},
			},
		},
		"",
		"  ",
	)
	if err != nil {
		panic(err)
	}

	encString := string(enc)

	if strings.Contains(encString, "We may never know.") || strings.Contains(encString, "Nights without you") {
		panic("not ghostly enough: contains cleartext")
	}

	if !strings.Contains(encString, `"text": ""`) {
		panic("not ghostly enough: lacking empty ghoststring")
	}

	if matched, err := regexp.MatchString(`.+"text": "👻:[^"]+"`, encString); !matched || err != nil {
		panic("not ghostly enough: lacking non-empty ghoststring")
	}

	fmt.Println("no peeking")

	// Output: no peeking
}
