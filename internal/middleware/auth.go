// Package middleware berisi seluruh middleware HTTP: otentikasi JWT sesi
// dashboard, otentikasi API key layanan WA-bot, RBAC per modul, tenant
// scoping, audit trail otomatis, CORS, rate limiting, dan security headers.
package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"curanmor-ai/internal/auth"
	"curanmor-ai/internal/httpx"
)

const CookieName = "curanmor_token"

// AuthRequired memverifikasi JWT dari cookie httpOnly (login web) atau
// header Authorization: Bearer (dipakai dari Swagger UI / klien API).
func AuthRequired(jwtSecret string) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if c, err := r.Cookie(CookieName); err == nil {
				token = c.Value
			}
			if token == "" {
				h := r.Header.Get("Authorization")
				if strings.HasPrefix(h, "Bearer ") {
					token = strings.TrimPrefix(h, "Bearer ")
				}
			}
			if token == "" {
				httpx.Unauthorized(w, "Sesi tidak ditemukan, silakan login")
				return
			}
			claims, err := auth.ParseToken(jwtSecret, token)
			if err != nil {
				httpx.Unauthorized(w, "Sesi tidak valid atau sudah kedaluwarsa, silakan login ulang")
				return
			}
			ctx := httpx.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// APIKeyRequired dipakai pada grup /api/v1/wa/* — otentikasi service-to-service
// untuk chatbot WhatsApp AI, terpisah dari sesi JWT pengguna dashboard.
func APIKeyRequired(database *sql.DB) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("X-API-Key")
			if raw == "" {
				httpx.Unauthorized(w, "Header X-API-Key wajib diisi")
				return
			}
			hash := auth.HashAPIKey(raw)
			var id int64
			var statusAktif bool
			err := database.QueryRow(
				`SELECT id, status_aktif FROM wa_api_key WHERE key_hash = $1`, hash,
			).Scan(&id, &statusAktif)
			if err == sql.ErrNoRows {
				httpx.Unauthorized(w, "API key tidak dikenali")
				return
			}
			if err != nil {
				httpx.InternalError(w, err)
				return
			}
			if !statusAktif {
				httpx.Forbidden(w, "API key dinonaktifkan")
				return
			}
			_, _ = database.Exec(`UPDATE wa_api_key SET last_used_at = now() WHERE id = $1`, id)
			ctx := httpx.WithAPIKeyID(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
