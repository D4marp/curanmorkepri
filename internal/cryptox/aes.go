// Package cryptox menyediakan enkripsi field sensitif at-rest (mis. NIK
// pada pihak_terkait) menggunakan AES-256-GCM. Kunci diambil dari
// environment variable APP_ENCRYPTION_KEY (di-derive dengan SHA-256 agar
// panjang key input bebas, output selalu 32 byte / AES-256).
package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

type Cipher struct {
	gcm cipher.AEAD
}

// New membuat Cipher dari passphrase (string apa pun, minimal disarankan
// 32 karakter acak). Passphrase di-hash SHA-256 agar selalu menghasilkan
// key 32-byte yang valid untuk AES-256.
func New(passphrase string) (*Cipher, error) {
	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{gcm: gcm}, nil
}

// Encrypt mengenkripsi plaintext, mengembalikan nonce||ciphertext.
func (c *Cipher) Encrypt(plaintext string) ([]byte, error) {
	if plaintext == "" {
		return nil, nil
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// Decrypt mendekripsi hasil Encrypt.
func (c *Cipher) Decrypt(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	nonceSize := c.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("data terenkripsi tidak valid")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// MaskNIK menyamarkan NIK untuk ditampilkan di UI/log non-privileged,
// mis. "3271xxxxxxxx1234" -> "3271********1234".
func MaskNIK(nik string) string {
	if len(nik) <= 8 {
		return "****"
	}
	return nik[:4] + "********" + nik[len(nik)-4:]
}

// EncodeB64 / DecodeB64 dipakai bila field perlu direpresentasikan sebagai
// string (mis. pada JSON request/response), sedangkan kolom DB adalah BYTEA.
func EncodeB64(b []byte) string          { return base64.StdEncoding.EncodeToString(b) }
func DecodeB64(s string) ([]byte, error) { return base64.StdEncoding.DecodeString(s) }
