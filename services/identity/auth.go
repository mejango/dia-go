package dia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

var rawURL = base64.RawURLEncoding.Strict()

const passwordIterations = 600000

// HashPassword encodes PBKDF2-HMAC-SHA256 parameters, a random salt and a 32-byte key.
func HashPassword(password string) (string, error) {
	if len(password) == 0 || len(password) > 4096 {
		return "", errors.New("invalid password length")
	}
	salt := make([]byte, 16)
	if _, e := rand.Read(salt); e != nil {
		return "", e
	}
	key, e := pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
	if e != nil {
		return "", e
	}
	return "pbkdf2-sha256$" + strconv.Itoa(passwordIterations) + "$" + rawURL.EncodeToString(salt) + "$" + rawURL.EncodeToString(key), nil
}

// CheckPassword takes the plaintext password followed by its stored hash.
func CheckPassword(password, encoded string) bool {
	if len(password) == 0 || len(password) > 4096 || len(encoded) > 256 {
		return false
	}
	p := strings.Split(encoded, "$")
	if len(p) != 4 || p[0] != "pbkdf2-sha256" {
		return false
	}
	n, e := strconv.Atoi(p[1])
	if e != nil || n < passwordIterations || n > 1000000 || strconv.Itoa(n) != p[1] {
		return false
	}
	salt, e := rawURL.DecodeString(p[2])
	if e != nil || len(salt) != 16 || rawURL.EncodeToString(salt) != p[2] {
		return false
	}
	want, e := rawURL.DecodeString(p[3])
	if e != nil || len(want) != 32 || rawURL.EncodeToString(want) != p[3] {
		return false
	}
	got, e := pbkdf2.Key(sha256.New, password, salt, n, 32)
	return e == nil && subtle.ConstantTimeCompare(got, want) == 1
}

// Seal prefixes a random GCM nonce to the authenticated ciphertext.
func Seal(key, plaintext []byte) ([]byte, error) {
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	g, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	nonce := make([]byte, g.NonceSize())
	if _, e = rand.Read(nonce); e != nil {
		return nil, e
	}
	return g.Seal(nonce, nonce, plaintext, nil), nil
}

func Open(key, ciphertext []byte) ([]byte, error) {
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	g, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	if len(ciphertext) < g.NonceSize()+g.Overhead() {
		return nil, errors.New("invalid ciphertext")
	}
	return g.Open(nil, ciphertext[:g.NonceSize()], ciphertext[g.NonceSize():], nil)
}
