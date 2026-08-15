// Package auth berisi implementasi JWT (HS256) minimal tanpa dependensi
// eksternal, otentikasi password (bcrypt), dan pembuatan token acak aman
// (API key, refresh token) menggunakan crypto/rand.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("token tidak valid")
	ErrExpiredToken = errors.New("token sudah kedaluwarsa")
)

// Claims merepresentasikan payload JWT untuk sesi pengguna dashboard.
type Claims struct {
	Subject     string `json:"sub"` // pengguna.id
	NRP         string `json:"nrp"`
	Nama        string `json:"nama"`
	PeranID     int64  `json:"peran_id"`
	PeranNama   string `json:"peran_nama"`
	SatkerID    int64  `json:"satker_id"`
	JenisSatker string `json:"jenis_satker"`
	Issuer      string `json:"iss"`
	IssuedAt    int64  `json:"iat"`
	ExpiresAt   int64  `json:"exp"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func b64encode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func b64decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// GenerateToken membuat JWT HS256 yang ditandatangani dengan secret.
func GenerateToken(secret string, c Claims, ttl time.Duration) (string, error) {
	now := time.Now()
	c.IssuedAt = now.Unix()
	c.ExpiresAt = now.Add(ttl).Unix()

	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	unsigned := b64encode(headerJSON) + "." + b64encode(payloadJSON)
	sig := sign(secret, unsigned)
	return unsigned + "." + b64encode(sig), nil
}

func sign(secret, data string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// ParseToken memvalidasi signature & expiry, mengembalikan Claims.
func ParseToken(secret, tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	unsigned := parts[0] + "." + parts[1]
	expectedSig := sign(secret, unsigned)

	actualSig, err := b64decode(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !hmac.Equal(expectedSig, actualSig) {
		return nil, ErrInvalidToken
	}

	payloadBytes, err := b64decode(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(payloadBytes, &c); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() > c.ExpiresAt {
		return nil, ErrExpiredToken
	}
	return &c, nil
}

// GenerateRandomToken membuat token acak aman (hex) untuk refresh token /
// API key mentah, menggunakan crypto/rand. nBytes=32 -> 64 karakter hex.
func GenerateRandomToken(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// HashAPIKey menghasilkan SHA-256 hash (hex) dari API key mentah. API key
// WA-bot disimpan sebagai hash saja di database (mirip password), sehingga
// kebocoran database tidak otomatis membocorkan kredensial layanan aktif.
func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// ConstantTimeEquals membandingkan dua string dengan waktu konstan untuk
// mencegah timing attack pada perbandingan token/hash.
func ConstantTimeEquals(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// NewAPIKey membuat pasangan (rawKey, hash, prefix) untuk wa_api_key.
// rawKey ditampilkan HANYA SEKALI ke pengguna; hash disimpan di DB.
func NewAPIKey() (rawKey, hash, prefix string, err error) {
	raw, err := GenerateRandomToken(32)
	if err != nil {
		return "", "", "", err
	}
	full := fmt.Sprintf("cma_%s", raw) // prefix "cma_" = CURANMOR-AI
	h := HashAPIKey(full)
	return full, h, full[:12], nil
}
