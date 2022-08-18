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
	gh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer(
		"test",
		"hang-glider-casserole-newt",
	)
	require.Nil(t, err)
	require.NotNil(t, gh)

	ghoststring.SetGhostifyer(gh)

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
			g: &ghoststring.GhostString{Namespace: "test", Str: "maybe"},
		},
	} {
		t.Run(fmt.Sprintf("namespace=%[1]q,string=%[2]v", tc.g.Namespace, tc.g.Str), func(t *testing.T) {
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

func TestGhostString_Stringified(t *testing.T) {
	gh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer(
		"test",
		"careful the beverage you're about to enjoy",
	)
	require.Nil(t, err)
	require.NotNil(t, gh)

	ghoststring.SetGhostifyer(gh)

	for _, tc := range []struct {
		g         *ghoststring.GhostString
		null      bool
		plusV     string
		plusHashV string
	}{
		{
			g:         &ghoststring.GhostString{},
			null:      true,
			plusHashV: `{"", ""}`,
		},
		{
			g:         &ghoststring.GhostString{Namespace: "test"},
			null:      true,
			plusHashV: `{"test", ""}`,
		},
		{
			g: &ghoststring.GhostString{Namespace: "test", Str: "maybe"},
		},
	} {
		t.Run(fmt.Sprintf("namespace=%[1]q,string=%[2]q", tc.g.Namespace, tc.g.Str), func(t *testing.T) {
			r := require.New(t)

			s := fmt.Sprintf("%s", tc.g)
			v := fmt.Sprintf("%v", tc.g)
			plusV := fmt.Sprintf("%+v", tc.g)
			plusHashV := fmt.Sprintf("%+#v", tc.g)

			if tc.null {
				r.Equalf("", s, "mismatched s form")
				r.Equalf("", v, "mismatched v form")
				r.Equalf(tc.plusV, plusV, "mismatched +v form")
				r.Equalf(tc.plusHashV, plusHashV, "mismatched +#v form")
				return
			}

			r.Contains(s, ghoststring.Prefix)
			r.Contains(v, ghoststring.Prefix)
			r.Contains(plusV, ghoststring.Prefix)
			r.Contains(plusHashV, ghoststring.Prefix)
		})
	}
}

func TestGhostString_TextMarshalRoundTrip(t *testing.T) {
	gh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer(
		"test",
		"let us add a little joy to your day",
	)
	require.Nil(t, err)
	require.NotNil(t, gh)

	ghoststring.SetGhostifyer(gh)

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
			g: &ghoststring.GhostString{Namespace: "test", Str: "maybe"},
		},
	} {
		t.Run(fmt.Sprintf("namespace=%[1]q,string=%[2]q", tc.g.Namespace, tc.g.Str), func(t *testing.T) {
			r := require.New(t)

			b, err := tc.g.MarshalText()
			if err != nil {
				r.ErrorIs(err, tc.err)
				return
			}

			if tc.null {
				r.Equal("", string(b))
				return
			}

			r.Contains(string(b), ghoststring.Prefix)

			gs := &ghoststring.GhostString{}
			r.Nil(gs.UnmarshalText(b))

			r.Equal(tc.g, gs)
		})
	}
}

func TestGhostString_BinaryMarshalRoundTrip(t *testing.T) {
	gh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer(
		"test",
		"do not microwave",
	)
	require.Nil(t, err)
	require.NotNil(t, gh)

	ghoststring.SetGhostifyer(gh)

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
			g: &ghoststring.GhostString{Namespace: "test", Str: "maybe"},
		},
	} {
		t.Run(fmt.Sprintf("namespace=%[1]q,string=%[2]q", tc.g.Namespace, tc.g.Str), func(t *testing.T) {
			r := require.New(t)

			b, err := tc.g.MarshalBinary()
			if err != nil {
				r.ErrorIs(err, tc.err)
				return
			}

			if tc.null {
				r.Len(b, 0)
				return
			}

			r.Contains(string(b), ghoststring.Prefix)

			gs := &ghoststring.GhostString{}
			r.Nil(gs.UnmarshalBinary(b))

			r.Equal(tc.g, gs)
		})
	}
}

func ExampleGhostString() {
	gh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer(
		"example",
		"correct horse battery staple",
	)
	if err != nil {
		panic(err)
	}

	if err := ghoststring.SetGhostifyer(gh); err != nil {
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
						Str:       "Nights without you are so dark. I pray that someday you will return my flashlight.",
					},
				},
				{
					Timestamp: time.UnixMicro(-8001),
					Text: ghoststring.GhostString{
						Namespace: "unknown",
						Str:       "We may never know.",
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

func exampleForREADME() error {
	// chunk 0
	type Message struct {
		Recipient string                  `json:"recipient"`
		Content   ghoststring.GhostString `json:"content"`
		Mood      ghoststring.GhostString `json:"mood"`
	}

	// chunk 1
	/*
		secretKeyBytes, err := os.ReadFile("/path/to/secret-key")
		if err != nil {
			return err
		}
	*/

	// Not included in example: {{
	secretKeyBytes := []byte("that first sip feeling")
	// }}

	// chunk 2
	gh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer("heck.example.org", string(secretKeyBytes))
	if err != nil {
		return err
	}

	// chunk 3
	if err := ghoststring.SetGhostifyer(gh); err != nil {
		return err
	}

	// chunk 4
	msg := &Message{
		Recipient: "morningstar@heck.example.org",
		Content: ghoststring.GhostString{
			Namespace: "heck.example.org",
			Str:       "We meet me at the fjord at dawn. Bring donuts, please.",
		},
		Mood: ghoststring.GhostString{
			Namespace: "heck.example.org",
			Str:       "giddy",
		},
	}

	// Not included in example: {{
	fmt.Printf("msg=%v\n", msg)
	return nil
	// }}
}
