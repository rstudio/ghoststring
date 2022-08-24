package ghoststring

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

const (
	EnvKeyStoreKeyPrefix = "GHOSTSTRING_KEY_{{.Namespace}}_"
)

var (
	envKeyStoreKeyPrefixTmpl = template.Must(template.New("key_prefix").Parse(EnvKeyStoreKeyPrefix))
)

type KeyStore interface {
	Latest(ctx context.Context) ([]byte, error)
	All(ctx context.Context) ([][]byte, error)
}

func NewKeyStore(namespace string, keys []*TimestampedKey) (KeyStore, error) {
	if len(keys) == 0 {
		return nil, errors.Wrap(Err, "no keys found")
	}

	for i := range keys {
		kb, err := newAES256GCMKey(namespace, keys[i].Key)
		if err != nil {
			return nil, err
		}

		keys[i].keyBytes = kb
	}

	return &inMemoryKeyStore{keys: keys}, nil
}

func NewKeyStoreFromEnv(namespace string, env []string) (KeyStore, error) {
	if env == nil {
		env = os.Environ()
	}

	prefixBuf := &bytes.Buffer{}
	if err := envKeyStoreKeyPrefixTmpl.Execute(
		prefixBuf,
		map[string]string{"Namespace": envKeySafeNamespace(namespace)},
	); err != nil {
		return nil, err
	}

	keyPrefix := prefixBuf.String()

	keys := []*TimestampedKey{}

	for _, v := range env {
		if !strings.HasPrefix(v, keyPrefix) {
			continue
		}

		parts := strings.SplitN(v, "=", 2)
		if len(parts) < 2 {
			continue
		}

		tk := &TimestampedKey{}
		if err := json.Unmarshal([]byte(parts[1]), tk); err != nil {
			return nil, err
		}

		keys = append(keys, tk)
	}

	return NewKeyStore(namespace, keys)
}

type inMemoryKeyStore struct {
	keys timestampedKeySlice
}

func (ks *inMemoryKeyStore) Latest(context.Context) ([]byte, error) {
	if ks.keys.Len() == 0 {
		return nil, errors.Wrap(Err, "no keys available")
	}

	sort.Sort(sort.Reverse(ks.keys))

	return ks.keys[0].keyBytes, nil
}

func (ks *inMemoryKeyStore) All(context.Context) ([][]byte, error) {
	if ks.keys.Len() == 0 {
		return nil, errors.Wrap(Err, "no keys available")
	}

	sort.Sort(sort.Reverse(ks.keys))

	sl := make([][]byte, ks.keys.Len())

	for i, tk := range ks.keys {
		sl[i] = tk.keyBytes
	}

	return sl, nil
}
