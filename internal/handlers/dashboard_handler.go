package handlers

import (
	"database/sql"
	"net/http"

	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/repo"
)

type DashboardHandler struct{ DB *sql.DB }

func (h *DashboardHandler) scope(r *http.Request) ([]int64, error) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	return repo.ResolveSatkerScope(h.DB, claims.SatkerID)
}

// GET /api/v1/dashboard/ringkasan
func (h *DashboardHandler) Ringkasan(w http.ResponseWriter, r *http.Request) {
	scope, err := h.scope(r)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	data, err := repo.DashboardRingkasan(h.DB, scope)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, data)
}

// GET /api/v1/dashboard/tren-bulanan?bulan=6
func (h *DashboardHandler) TrenBulanan(w http.ResponseWriter, r *http.Request) {
	scope, err := h.scope(r)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	bulan := queryInt(r, "bulan", 6)
	data, err := repo.DashboardTrenBulanan(h.DB, scope, bulan)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, data)
}

// GET /api/v1/dashboard/kategori-kasus
func (h *DashboardHandler) KategoriKasus(w http.ResponseWriter, r *http.Request) {
	scope, err := h.scope(r)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	data, err := repo.DashboardKategoriKasus(h.DB, scope)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, data)
}

// GET /api/v1/dashboard/peta-sebaran?dari=&sampai=
// Data untuk peta sebaran lokasi motor belum/sudah terungkap (Damar/Steven,
// diisi manual oleh anggota di lapangan lewat tkp_latitude/tkp_longitude).
func (h *DashboardHandler) PetaSebaran(w http.ResponseWriter, r *http.Request) {
	scope, err := h.scope(r)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	q := r.URL.Query()
	data, err := repo.PetaSebaran(h.DB, scope, q.Get("dari"), q.Get("sampai"))
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, data)
}
