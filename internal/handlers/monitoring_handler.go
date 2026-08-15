package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

type MonitoringHandler struct{ DB *sql.DB }

// GET /api/v1/notifikasi
func (h *MonitoringHandler) ListNotifikasi(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	list, err := repo.ListNotifikasiUntukPengguna(h.DB, parseID(claims.Subject), claims.PeranID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

// PATCH /api/v1/notifikasi/{id}/baca
func (h *MonitoringHandler) MarkNotifikasiRead(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	err := repo.MarkNotifikasiRead(h.DB, pathID(r), parseID(claims.Subject), claims.PeranID)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Notifikasi tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Notifikasi ditandai dibaca"})
}

// GET /api/v1/laporan-periodik
func (h *MonitoringHandler) ListLaporanPeriodik(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	list, err := repo.ListLaporanPeriodik(h.DB, scope)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

type createLaporanPeriodikRequest struct {
	JenisLaporan   string `json:"jenis_laporan"`
	PeriodeMulai   string `json:"periode_mulai"`
	PeriodeSelesai string `json:"periode_selesai"`
	SatkerID       *int64 `json:"satker_id"`
	FileURL        string `json:"file_url"`
}

// POST /api/v1/laporan-periodik  — pembuatan manual (proses otomatis harian/
// mingguan/bulanan dapat dijalankan lewat scheduled job terpisah yang
// memanggil endpoint yang sama).
func (h *MonitoringHandler) CreateLaporanPeriodik(w http.ResponseWriter, r *http.Request) {
	var req createLaporanPeriodikRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	valid := map[string]bool{"Harian": true, "Mingguan": true, "Bulanan": true}
	if !valid[req.JenisLaporan] || req.PeriodeMulai == "" || req.PeriodeSelesai == "" {
		httpx.BadRequest(w, "jenis_laporan, periode_mulai, periode_selesai wajib diisi dengan benar")
		return
	}
	if req.SatkerID != nil {
		if ok, err := authorizeSatkerScope(h.DB, r, *req.SatkerID); err != nil {
			httpx.InternalError(w, err)
			return
		} else if !ok {
			httpx.Forbidden(w, "Anda tidak berwenang membuat laporan periodik untuk satuan kerja tersebut")
			return
		}
	}
	l := models.LaporanPeriodik{
		JenisLaporan: req.JenisLaporan, PeriodeMulai: req.PeriodeMulai, PeriodeSelesai: req.PeriodeSelesai,
		SatkerID: req.SatkerID, FileURL: req.FileURL, DibuatOtomatis: false,
	}
	id, err := repo.CreateLaporanPeriodik(h.DB, &l)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	created, err := repo.GetLaporanPeriodikByID(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

// GET /api/v1/audit-log?page=&page_size=  (Super Admin / Viewer terbatas)
// Dibatasi cakupan satuan kerja pemanggil — Admin Polres/Viewer hanya melihat
// jurnal aktivitas pengguna di satker sendiri & jajaran, bukan seluruh Polda.
func (h *MonitoringHandler) ListAuditLog(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 100)
	list, total, err := repo.ListAuditLog(h.DB, scope, page, pageSize)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OKMeta(w, list, map[string]interface{}{"total": total, "page": page, "page_size": pageSize})
}
