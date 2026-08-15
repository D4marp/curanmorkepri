// Package config memuat konfigurasi aplikasi dari environment variable.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv          string
	HTTPPort        string
	DatabaseURL     string
	JWTSecret       string
	JWTIssuer       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	EncryptionKey   string // 32 byte key (hex/base64) untuk AES-256-GCM (enkripsi NIK)
	UploadDir       string
	MaxUploadSizeMB int64
	CORSOrigins     string
	CookieSecure    bool
	CookieDomain    string
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable %s wajib diisi", key))
	}
	return v
}

func envOr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

// Load membaca konfigurasi dari environment variable.
// Variabel wajib (tanpa default aman) akan menyebabkan panic saat start-up
// agar kesalahan konfigurasi keamanan terdeteksi lebih awal, bukan diam-diam
// memakai nilai default yang tidak aman di lingkungan produksi.
func Load() *Config {
	cfg := &Config{
		AppEnv:          envOr("APP_ENV", "development"),
		HTTPPort:        envOr("HTTP_PORT", "8080"),
		DatabaseURL:     mustEnv("DATABASE_URL"),
		JWTSecret:       mustEnv("JWT_SECRET"),
		JWTIssuer:       envOr("JWT_ISSUER", "curanmor-ai.polda-kepri.go.id"),
		AccessTokenTTL:  time.Duration(envInt64("ACCESS_TOKEN_TTL_MINUTES", 60)) * time.Minute,
		RefreshTokenTTL: time.Duration(envInt64("REFRESH_TOKEN_TTL_HOURS", 168)) * time.Hour,
		EncryptionKey:   mustEnv("APP_ENCRYPTION_KEY"),
		UploadDir:       envOr("UPLOAD_DIR", "/app/uploads"),
		MaxUploadSizeMB: envInt64("MAX_UPLOAD_SIZE_MB", 10),
		CORSOrigins:     envOr("CORS_ALLOWED_ORIGINS", "*"),
		CookieSecure:    envBool("COOKIE_SECURE", true),
		CookieDomain:    envOr("COOKIE_DOMAIN", ""),
	}
	if len(cfg.JWTSecret) < 32 {
		panic("JWT_SECRET minimal 32 karakter (gunakan string acak yang kuat)")
	}
	if len(cfg.EncryptionKey) < 32 {
		panic("APP_ENCRYPTION_KEY minimal 32 karakter (dipakai sebagai kunci AES-256)")
	}
	return cfg
}
