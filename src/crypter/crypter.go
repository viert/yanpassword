package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"term"

	"golang.org/x/crypto/pbkdf2"
)

const (
	iterCount = 10
)

func createHash(passphrase string) ([]byte, error) {
	bytePasswd := []byte(passphrase)
	hasher := md5.New()
	hasher.Write(bytePasswd)
	salt := hasher.Sum(nil)

	phash := pbkdf2.Key(bytePasswd, salt, iterCount, 4096, sha1.New)
	hasher.Reset()
	hasher.Write(phash)
	pwdKey := hex.EncodeToString(hasher.Sum(nil))
	return []byte(pwdKey), nil
}

func createHashLegacy(passphrase string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(passphrase))
	pwdKey := hex.EncodeToString(hasher.Sum(nil))
	return []byte(pwdKey)
}

// Encrypt encrypts data with a passphrase
func Encrypt(data []byte, passphrase string) ([]byte, error) {
	// generate password hash
	pwdKey, err := createHash(passphrase)
	if err != nil {
		return nil, err
	}

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
	pwdKey, err := createHash(passphrase)
	if err != nil {
		return nil, err
	}

	legacyTried := false
	for {
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
			if legacyTried {
				return nil, err
			}
			term.Warnf("Error decrypting data, falling back to legacy MD5 crypter\n")
			pwdKey = createHashLegacy(passphrase)
			legacyTried = true
			continue
		} else {
			return data, nil
		}
	}
}
