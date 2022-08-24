package ghoststring

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"

	"golang.org/x/crypto/argon2"
)

const (
	argon2SaltPrefix = "github.com/rstudio/ghoststring:"

	aesKeyLen     = 32
	argon2Mem     = 64 * 1024
	argon2Threads = 4
	argon2Time    = 1
)

func newAES256GCMKey(namespace, key string) ([]byte, error) {
	if err := validateNamespace(namespace); err != nil {
		return nil, err
	}

	saltSHA1Bytes := sha1.Sum(append([]byte(argon2SaltPrefix), []byte(namespace)...))

	return argon2.IDKey(
		[]byte(key),
		saltSHA1Bytes[:],
		argon2Time,
		argon2Mem,
		argon2Threads,
		aesKeyLen,
	), nil
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
