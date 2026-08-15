package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/auth"
	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/repo"
)

// APIKeyHandler mengelola kredensial layanan (wa_api_key) yang dipakai
// chatbot WhatsApp AI untuk mengakses grup route /api/v1/wa/*. Hanya
// Super Admin (modul pengaturan_sistem) yang boleh membuat/mencabut key.
type APIKeyHandler struct{ DB *sql.DB }

// GET /api/v1/api-keys
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := repo.ListAPIKeys(h.DB)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

type createAPIKeyRequest struct {
	NamaLayanan string `json:"nama_layanan"`
}

// POST /api/v1/api-keys
// PENTING: raw_key HANYA ditampilkan sekali di response ini. Simpan segera
// di konfigurasi layanan chatbot WA (header X-API-Key).
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.NamaLayanan == "" {
		httpx.BadRequest(w, "nama_layanan wajib diisi")
		return
	}
	rawKey, hash, prefix, err := auth.NewAPIKey()
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	creator := parseID(claims.Subject)
	id, err := repo.CreateAPIKey(h.DB, req.NamaLayanan, hash, prefix, creator)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, map[string]interface{}{
		"id": id, "nama_layanan": req.NamaLayanan, "raw_key": rawKey, "key_prefix": prefix,
		"peringatan": "Simpan raw_key sekarang — tidak akan ditampilkan lagi.",
	})
}

// DELETE /api/v1/api-keys/{id}  (mencabut/menonaktifkan)
func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	if err := repo.RevokeAPIKey(h.DB, pathID(r)); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "API key dicabut"})
}
