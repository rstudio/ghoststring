package ghoststring

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTimestampedKey(t *testing.T) {
	r := require.New(t)

	tk := &TimestampedKey{
		Timestamp: 42,
		Key:       "correct horse battery paperclip",
	}

	b, err := json.Marshal(tk)
	r.NotNil(b)
	r.Nil(err)

	keys := timestampedKeySlice{
		tk,
		&TimestampedKey{
			Timestamp: 11,
			Key:       "suspicious tote ukelele bongo",
		},
		&TimestampedKey{
			Timestamp: -1,
			Key:       "green smoothie watch fan",
		},
	}

	r.Implements((*sort.Interface)(nil), keys)
	r.Equal(3, keys.Len())
	r.True(keys.Less(2, 0))

	keys.Swap(0, 2)

	r.True(keys.Less(0, 2))
}
