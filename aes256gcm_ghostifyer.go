package ghoststring

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
)

type aes256GcmGhostifyer struct {
	key []byte
}

func (g *aes256GcmGhostifyer) Ghostify(gs *GhostString) (string, error) {
	if strings.TrimSpace(string(g.key)) == "" {
		return "", errors.Wrap(Err, "invalid key")
	}

	if gs == nil || !gs.IsValid() {
		return jsonNull, nil
	}

	nonce := make([]byte, nonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	encBytes, err := aes256GcmEncrypt(g.key, nonce, gs.String)
	if err != nil {
		return "", err
	}

	b64Value := base64.StdEncoding.EncodeToString(
		append(
			append(
				[]byte(hex.EncodeToString(nonce)),
				[]byte(gs.Namespace+namespaceSeparator)...,
			),
			encBytes...,
		),
	)

	return prefix + b64Value, nil
}

func (g *aes256GcmGhostifyer) Unghostify(s string) (*GhostString, error) {
	if strings.TrimSpace(string(g.key)) == "" {
		return nil, errors.Wrap(Err, "invalid key")
	}

	if s == jsonNull || s == "" {
		return &GhostString{}, nil
	}

	nonceNsValueBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, prefix))
	if err != nil {
		return nil, err
	}

	hexNonce, nsValueBytes := nonceNsValueBytes[:nonceLengthHex], nonceNsValueBytes[nonceLengthHex:]

	nonce, err := hex.DecodeString(string(hexNonce))
	if err != nil {
		return nil, err
	}

	nsParts := strings.SplitN(string(nsValueBytes), namespaceSeparator, namespacePartsLength)
	if len(nsParts) != namespacePartsLength {
		return nil, errors.Wrap(Err, "invalid namespacing")
	}

	plainBytes, err := aes256GcmDecrypt(g.key, nonce, nsParts[1])
	if err != nil {
		return nil, err
	}

	return &GhostString{Namespace: nsParts[0], String: string(plainBytes)}, nil
}

func aes256GcmEncrypt(key, nonce []byte, plainText string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Seal(nil, nonce, []byte(plainText), nil), nil
}

func aes256GcmDecrypt(key, nonce []byte, cipherText string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plainText, err := aesgcm.Open(nil, nonce, []byte(cipherText), nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
