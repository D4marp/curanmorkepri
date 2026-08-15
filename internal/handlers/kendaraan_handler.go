package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

type KendaraanHandler struct{ DB *sql.DB }

// GET /api/v1/kendaraan/search?tipe=No.+Polisi&kata_kunci=BP1234AB
// Pencarian cepat sisi dashboard (terpisah dari /api/v1/wa/search milik chatbot).
func (h *KendaraanHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	tipe := r.URL.Query().Get("tipe")
	kataKunci := r.URL.Query().Get("kata_kunci")
	if kataKunci == "" {
		httpx.BadRequest(w, "Parameter kata_kunci wajib diisi")
		return
	}
	result, err := repo.SearchKendaraan(h.DB, tipe, kataKunci, scope)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, result)
}

// POST /api/v1/laporan/{id}/kendaraan
func (h *KendaraanHandler) Create(w http.ResponseWriter, r *http.Request) {
	lpID := pathID(r)
	if ok, err := authorizeLPScope(h.DB, r, lpID); err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	} else if err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menambah kendaraan pada laporan satuan kerja tersebut")
		return
	}
	var k models.Kendaraan
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	k.LPID = lpID
	if k.JenisKendaraan == "" {
		k.JenisKendaraan = "Roda 2"
	}
	if k.StatusKendaraan == "" {
		k.StatusKendaraan = "Terlapor Hilang"
	}
	id, err := repo.CreateKendaraan(h.DB, &k)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	created, err := repo.GetKendaraanByID(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

// PUT /api/v1/kendaraan/{id}
func (h *KendaraanHandler) Update(w http.ResponseWriter, r *http.Request) {
	var k models.Kendaraan
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	k.ID = pathID(r)

	existing, err := repo.GetKendaraanByID(h.DB, k.ID)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Kendaraan tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeLPScope(h.DB, r, existing.LPID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengubah kendaraan satuan kerja tersebut")
		return
	}

	if err := repo.UpdateKendaraan(h.DB, &k); err != nil {
		httpx.InternalError(w, err)
		return
	}
	updated, err := repo.GetKendaraanByID(h.DB, k.ID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, updated)
}

type updateStatusKendaraanRequest struct {
	Status string `json:"status_kendaraan"`
}

// PATCH /api/v1/kendaraan/{id}/status
func (h *KendaraanHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req updateStatusKendaraanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	valid := map[string]bool{"Terlapor Hilang": true, "Ditemukan": true, "Diamankan": true, "Dikembalikan": true}
	if !valid[req.Status] {
		httpx.BadRequest(w, "status_kendaraan tidak valid")
		return
	}
	id := pathID(r)
	existing, err := repo.GetKendaraanByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Kendaraan tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeLPScope(h.DB, r, existing.LPID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengubah status kendaraan satuan kerja tersebut")
		return
	}
	if err := repo.UpdateStatusKendaraan(h.DB, id, req.Status); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Status kendaraan diperbarui"})
}

// DELETE /api/v1/kendaraan/{id}
func (h *KendaraanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetKendaraanByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Kendaraan tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeLPScope(h.DB, r, existing.LPID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menghapus kendaraan satuan kerja tersebut")
		return
	}
	if err := repo.DeleteKendaraan(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Kendaraan dihapus"})
}
