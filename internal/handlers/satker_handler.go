package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

type SatkerHandler struct{ DB *sql.DB }

// GET /api/v1/satuan-kerja
func (h *SatkerHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := repo.ListSatuanKerja(h.DB)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

// GET /api/v1/satuan-kerja/{id}
func (h *SatkerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	s, err := repo.GetSatuanKerja(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Satuan kerja tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeSatkerScope(h.DB, r, id); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang melihat satuan kerja tersebut")
		return
	}
	httpx.OK(w, s)
}

// POST /api/v1/satuan-kerja  (Super Admin saja, dijaga RBAC pengaturan_sistem)
func (h *SatkerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var s models.SatuanKerja
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if s.KodeSatker == "" || s.NamaSatker == "" || s.JenisSatker == "" {
		httpx.BadRequest(w, "kode_satker, nama_satker, jenis_satker wajib diisi")
		return
	}
	id, err := repo.CreateSatuanKerja(h.DB, &s)
	if err != nil {
		httpx.Conflict(w, "Gagal membuat satuan kerja (kode mungkin sudah dipakai): "+err.Error())
		return
	}
	created, err := repo.GetSatuanKerja(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

// PUT /api/v1/satuan-kerja/{id}
func (h *SatkerHandler) Update(w http.ResponseWriter, r *http.Request) {
	var s models.SatuanKerja
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	s.ID = pathID(r)
	if err := repo.UpdateSatuanKerja(h.DB, &s); err != nil {
		httpx.InternalError(w, err)
		return
	}
	updated, err := repo.GetSatuanKerja(h.DB, s.ID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, updated)
}

// DELETE /api/v1/satuan-kerja/{id}
func (h *SatkerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := repo.DeleteSatuanKerja(h.DB, pathID(r)); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Satuan kerja dihapus"})
}
