package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func createHash(passphrase string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(passphrase))
	pwdKey := hex.EncodeToString(hasher.Sum(nil))
	return []byte(pwdKey)
}

// Encrypt encrypts data with a passphrase
func Encrypt(data []byte, passphrase string) ([]byte, error) {
	// generate password hash
	pwdKey := createHash(passphrase)

	// Creating cipher
	block, _ := aes.NewCipher(pwdKey)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generating Nonce
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

// Decrypt decrypts data with a passphrase
func Decrypt(encrypted []byte, passphrase string) ([]byte, error) {
	// generate password hash
	pwdKey := createHash(passphrase)

	// Creating cipher
	block, _ := aes.NewCipher(pwdKey)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	data, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}
