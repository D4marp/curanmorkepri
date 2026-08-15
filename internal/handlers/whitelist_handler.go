package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

type WhitelistHandler struct{ DB *sql.DB }

// GET /api/v1/whitelist-whatsapp
func (h *WhitelistHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	list, err := repo.ListWhitelist(h.DB, scope)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

type createWhitelistRequest struct {
	NoWhatsapp string `json:"no_whatsapp"`
	PenggunaID int64  `json:"pengguna_id"`
}

// POST /api/v1/whitelist-whatsapp
func (h *WhitelistHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	var req createWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.NoWhatsapp == "" || req.PenggunaID == 0 {
		httpx.BadRequest(w, "no_whatsapp dan pengguna_id wajib diisi")
		return
	}
	target, err := repo.GetPenggunaByID(h.DB, req.PenggunaID)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Pengguna tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeSatkerScope(h.DB, r, target.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mendaftarkan whitelist untuk pengguna di satuan kerja tersebut")
		return
	}
	registrar := parseID(claims.Subject)
	wwl := models.WhitelistWhatsapp{NoWhatsapp: req.NoWhatsapp, PenggunaID: req.PenggunaID, Status: "aktif", DiregistrasiOleh: &registrar}
	id, err := repo.CreateWhitelist(h.DB, &wwl)
	if err != nil {
		httpx.Conflict(w, "Nomor sudah terdaftar di whitelist: "+err.Error())
		return
	}
	created, err := repo.GetWhitelistByID(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

type updateWhitelistStatusRequest struct {
	Status string `json:"status"`
}

// authorizeWhitelistScope memastikan satker pemilik akun pengguna yang
// terkait entri whitelist_whatsapp berada dalam cakupan satuan kerja
// pengguna yang login saat ini.
func (h *WhitelistHandler) authorizeWhitelistScope(r *http.Request, ww *models.WhitelistWhatsapp) (bool, error) {
	target, err := repo.GetPenggunaByID(h.DB, ww.PenggunaID)
	if err != nil {
		return false, err
	}
	return authorizeSatkerScope(h.DB, r, target.SatkerID)
}

// PATCH /api/v1/whitelist-whatsapp/{id}/status
func (h *WhitelistHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req updateWhitelistStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.Status != "aktif" && req.Status != "nonaktif" && req.Status != "diblokir" {
		httpx.BadRequest(w, "status harus salah satu dari: aktif, nonaktif, diblokir")
		return
	}
	id := pathID(r)
	existing, err := repo.GetWhitelistByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Whitelist tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := h.authorizeWhitelistScope(r, existing); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengubah whitelist satuan kerja tersebut")
		return
	}
	if err := repo.UpdateWhitelistStatus(h.DB, id, req.Status); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Status whitelist diperbarui"})
}

// DELETE /api/v1/whitelist-whatsapp/{id}
func (h *WhitelistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetWhitelistByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Whitelist tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := h.authorizeWhitelistScope(r, existing); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menghapus whitelist satuan kerja tersebut")
		return
	}
	if err := repo.DeleteWhitelist(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Whitelist dihapus"})
}
