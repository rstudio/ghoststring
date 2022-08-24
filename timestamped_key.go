package ghoststring

type TimestampedKey struct {
	Timestamp int64  `json:"timestamp"`
	Key       string `json:"key"`

	keyBytes []byte
}

type timestampedKeySlice []*TimestampedKey

func (ks timestampedKeySlice) Len() int {
	return len(ks)
}

func (ks timestampedKeySlice) Less(i, j int) bool {
	return ks[i].Timestamp < ks[j].Timestamp
}

func (ks timestampedKeySlice) Swap(i, j int) {
	ks[i], ks[j] = ks[j], ks[i]
}
