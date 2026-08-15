// Package handlers berisi HTTP handler per domain, memanggil internal/repo
// untuk akses data dan internal/httpx untuk response envelope standar.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"curanmor-ai/internal/auth"
	"curanmor-ai/internal/config"
	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/middleware"
	"curanmor-ai/internal/repo"
)

type AuthHandler struct {
	DB  *sql.DB
	Cfg *config.Config
}

type loginRequest struct {
	NRP      string `json:"nrp"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token    string      `json:"token"`
	Pengguna interface{} `json:"pengguna"`
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.NRP == "" || req.Password == "" {
		httpx.BadRequest(w, "NRP dan password wajib diisi")
		return
	}

	p, err := repo.GetPenggunaByNRP(h.DB, req.NRP)
	if err == sql.ErrNoRows {
		middleware.RecordAuditManual(h.DB, nil, "login gagal (NRP tidak ditemukan): "+req.NRP, "auth", httpx.ClientIP(r), r.UserAgent())
		httpx.Unauthorized(w, "NRP atau password salah")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if !p.StatusAktif {
		httpx.Forbidden(w, "Akun dinonaktifkan, hubungi administrator")
		return
	}
	if !auth.CheckPassword(p.PasswordHash, req.Password) {
		middleware.RecordAuditManual(h.DB, &p.ID, "login gagal (password salah)", "auth", httpx.ClientIP(r), r.UserAgent())
		httpx.Unauthorized(w, "NRP atau password salah")
		return
	}

	claims := auth.Claims{
		Subject:     itoa64(p.ID),
		NRP:         p.NRP,
		Nama:        p.NamaLengkap,
		PeranID:     p.PeranID,
		PeranNama:   p.PeranNama,
		SatkerID:    p.SatkerID,
		JenisSatker: p.JenisSatker,
		Issuer:      h.Cfg.JWTIssuer,
	}
	token, err := auth.GenerateToken(h.Cfg.JWTSecret, claims, h.Cfg.AccessTokenTTL)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.Cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
		Domain:   h.Cfg.CookieDomain,
		Expires:  time.Now().Add(h.Cfg.AccessTokenTTL),
	})

	middleware.RecordAuditManual(h.DB, &p.ID, "login berhasil", "auth", httpx.ClientIP(r), r.UserAgent())

	full, err := repo.GetPenggunaByID(h.DB, p.ID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, loginResponse{Token: token, Pengguna: full})
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: middleware.CookieName, Value: "", Path: "/", HttpOnly: true,
		Secure: h.Cfg.CookieSecure, SameSite: http.SameSiteStrictMode, MaxAge: -1,
	})
	httpx.OK(w, map[string]string{"message": "Logout berhasil"})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w, "Sesi tidak ditemukan")
		return
	}
	p, err := repo.GetPenggunaByID(h.DB, parseID(claims.Subject))
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, p)
}

type changePasswordRequest struct {
	PasswordLama string `json:"password_lama"`
	PasswordBaru string `json:"password_baru"`
}

// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w, "Sesi tidak ditemukan")
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if len(req.PasswordBaru) < 10 {
		httpx.BadRequest(w, "Password baru minimal 10 karakter")
		return
	}
	id := parseID(claims.Subject)
	p, err := repo.GetPenggunaByNRP(h.DB, claims.NRP)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if !auth.CheckPassword(p.PasswordHash, req.PasswordLama) {
		httpx.Unauthorized(w, "Password lama tidak sesuai")
		return
	}
	hash, err := auth.HashPassword(req.PasswordBaru)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if err := repo.UpdatePenggunaPassword(h.DB, id, hash); err != nil {
		httpx.InternalError(w, err)
		return
	}
	middleware.RecordAuditManual(h.DB, &id, "ganti password", "auth", httpx.ClientIP(r), r.UserAgent())
	httpx.OK(w, map[string]string{"message": "Password berhasil diubah"})
}
