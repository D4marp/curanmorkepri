package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/repo"
)

type RBACHandler struct{ DB *sql.DB }

// GET /api/v1/peran
func (h *RBACHandler) ListPeran(w http.ResponseWriter, r *http.Request) {
	list, err := repo.ListPeran(h.DB)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

// GET /api/v1/modul-akses
func (h *RBACHandler) ListModul(w http.ResponseWriter, r *http.Request) {
	list, err := repo.ListModulAkses(h.DB)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

// GET /api/v1/rbac/matriks
func (h *RBACHandler) Matriks(w http.ResponseWriter, r *http.Request) {
	list, err := repo.ListMatriksRBAC(h.DB)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

type updateMatriksRequest struct {
	PeranID    int64  `json:"peran_id"`
	ModulID    int64  `json:"modul_id"`
	LevelAkses string `json:"level_akses"`
}

// PUT /api/v1/rbac/matriks  (Super Admin saja — modul pengaturan_sistem)
func (h *RBACHandler) UpdateMatriks(w http.ResponseWriter, r *http.Request) {
	var req updateMatriksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.LevelAkses != "penuh" && req.LevelAkses != "terbatas" && req.LevelAkses != "ditolak" {
		httpx.BadRequest(w, "level_akses harus salah satu dari: penuh, terbatas, ditolak")
		return
	}
	if err := repo.UpsertPeranModulAkses(h.DB, req.PeranID, req.ModulID, req.LevelAkses); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Matriks RBAC diperbarui"})
}
