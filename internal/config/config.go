// Package config memuat konfigurasi aplikasi dari environment variable.
package config

import (
	"fmt"
	"net/url"
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

// resolveDatabaseURL menentukan connection string Postgres yang dipakai.
// Prioritas:
//  1. DATABASE_URL apa adanya, bila diisi (docker-compose lokal & override manual).
//  2. Dirakit dari DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME (+DB_SSLMODE opsional)
//     — dipakai saat platform hosting cuma kasih field terpisah (Host/Port/
//     Username/Password/Database Name), bukan satu string gabungan. Dirakit
//     lewat net/url.URL, BUKAN concat string manual, supaya password yang
//     mengandung karakter spesial (@, :, /, dst — umum pada password random
//     yang di-generate provider DB terkelola) tidak merusak format URL.
func resolveDatabaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	host := mustEnv("DB_HOST")
	port := envOr("DB_PORT", "5432")
	user := mustEnv("DB_USER")
	password := mustEnv("DB_PASSWORD")
	dbname := mustEnv("DB_NAME")
	sslmode := envOr("DB_SSLMODE", "require")

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     host + ":" + port,
		Path:     "/" + dbname,
		RawQuery: "sslmode=" + url.QueryEscape(sslmode),
	}
	return u.String()
}

// Load membaca konfigurasi dari environment variable.
// Variabel wajib (tanpa default aman) akan menyebabkan panic saat start-up
// agar kesalahan konfigurasi keamanan terdeteksi lebih awal, bukan diam-diam
// memakai nilai default yang tidak aman di lingkungan produksi.
func Load() *Config {
	cfg := &Config{
		AppEnv:          envOr("APP_ENV", "development"),
		HTTPPort:        envOr("HTTP_PORT", "8080"),
		DatabaseURL:     resolveDatabaseURL(),
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
