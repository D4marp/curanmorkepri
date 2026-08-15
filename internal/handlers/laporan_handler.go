package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/cryptox"
	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

type LaporanHandler struct {
	DB     *sql.DB
	Crypto *cryptox.Cipher
}

// GET /api/v1/laporan?dari=&sampai=&status=&jenis_perkara=&satker_id=&keyword=&page=&page_size=
// Mendukung pencarian rentang waktu (Steven: "curanmor di batam dalam
// seminggu" / "dari tanggal ... sampai tanggal ...") lewat dari & sampai.
func (h *LaporanHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	q := r.URL.Query()
	f := repo.LaporanFilter{
		SatkerScope:   scope,
		SatkerID:      queryIntPtr64(r, "satker_id"),
		StatusKasus:   q.Get("status"),
		JenisPerkara:  q.Get("jenis_perkara"),
		DariTanggal:   q.Get("dari"),
		SampaiTanggal: q.Get("sampai"),
		Keyword:       q.Get("keyword"),
		Page:          queryInt(r, "page", 1),
		PageSize:      queryInt(r, "page_size", 25),
	}
	list, total, err := repo.SearchLaporan(h.DB, f)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OKMeta(w, list, map[string]interface{}{"total": total, "page": f.Page, "page_size": f.PageSize})
}

// authorizeLaporanScope memastikan satker pemilik laporan berada dalam cakupan
// satuan kerja pengguna yang login (Admin Polda: semua; Admin Polres: satker
// sendiri + jajaran Polsek; Admin Polsek: satker sendiri saja). Tanpa
// pengecekan ini, endpoint per-ID (Get/Update/UpdateStatus/Delete) dapat
// diakses lintas-tenant hanya dengan menebak ID — celah otorisasi kritis
// untuk sistem data kasus kepolisian ini.
func (h *LaporanHandler) authorizeLaporanScope(r *http.Request, satkerID int64) (bool, error) {
	return authorizeSatkerScope(h.DB, r, satkerID)
}

// GET /api/v1/laporan/{id}  — detail lengkap dengan seluruh relasi
func (h *LaporanHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	lp, err := repo.GetLaporanByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := h.authorizeLaporanScope(r, lp.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengakses laporan satuan kerja tersebut")
		return
	}

	if lp.Kendaraan, err = repo.ListKendaraanByLP(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	pihak, err := repo.ListPihakByLP(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	for i := range pihak {
		encRow, err := repo.GetPihakNikEnc(h.DB, pihak[i].ID)
		if err == nil && len(encRow) > 0 {
			if plain, derr := h.Crypto.Decrypt(encRow); derr == nil && plain != "" {
				pihak[i].NIKMasked = cryptox.MaskNIK(plain)
			}
		}
	}
	lp.PihakTerkait = pihak

	if lp.BarangBukti, err = repo.ListBarangBuktiByLP(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	if lp.Dokumentasi, err = repo.ListDokumentasiByLP(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	if lp.RiwayatStatus, err = repo.ListRiwayatStatusByLP(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	if lp.RiwayatPenyidikan, err = repo.ListRiwayatPenyidikanByLP(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, lp)
}

// POST /api/v1/laporan
func (h *LaporanHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	var lp models.LaporanPolisi
	if err := json.NewDecoder(r.Body).Decode(&lp); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if lp.NoLP == "" || lp.TanggalLP == "" || lp.SatkerID == 0 {
		httpx.BadRequest(w, "no_lp, tanggal_lp, satker_id wajib diisi")
		return
	}
	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if !inScope(scope, lp.SatkerID) {
		httpx.Forbidden(w, "Anda tidak berwenang membuat laporan untuk satuan kerja tersebut")
		return
	}
	if lp.JenisPerkara == "" {
		lp.JenisPerkara = "Pencurian Ranmor"
	}
	if lp.StatusKasus == "" {
		lp.StatusKasus = "LP"
	}
	if lp.PenyidikID == nil {
		pid := parseID(claims.Subject)
		lp.PenyidikID = &pid
	}
	id, err := repo.CreateLaporan(h.DB, &lp)
	if err != nil {
		httpx.Conflict(w, "Gagal membuat laporan (no_lp mungkin sudah dipakai): "+err.Error())
		return
	}
	created, err := repo.GetLaporanByID(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

// PUT /api/v1/laporan/{id}
func (h *LaporanHandler) Update(w http.ResponseWriter, r *http.Request) {
	var lp models.LaporanPolisi
	if err := json.NewDecoder(r.Body).Decode(&lp); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	lp.ID = pathID(r)

	existing, err := repo.GetLaporanByID(h.DB, lp.ID)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := h.authorizeLaporanScope(r, existing.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengubah laporan satuan kerja tersebut")
		return
	}
	if lp.SatkerID != 0 {
		if ok, err := h.authorizeLaporanScope(r, lp.SatkerID); err != nil {
			httpx.InternalError(w, err)
			return
		} else if !ok {
			httpx.Forbidden(w, "Anda tidak berwenang memindahkan laporan ke satuan kerja tersebut")
			return
		}
	} else {
		lp.SatkerID = existing.SatkerID
	}

	if err := repo.UpdateLaporan(h.DB, &lp); err != nil {
		httpx.InternalError(w, err)
		return
	}
	updated, err := repo.GetLaporanByID(h.DB, lp.ID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, updated)
}

type updateStatusRequest struct {
	StatusBaru string `json:"status_baru"`
	Keterangan string `json:"keterangan"`
}

// PATCH /api/v1/laporan/{id}/status
func (h *LaporanHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	valid := map[string]bool{"LP": true, "SP2HP": true, "DPO": true, "Selesai": true}
	if !valid[req.StatusBaru] {
		httpx.BadRequest(w, "status_baru harus salah satu dari: LP, SP2HP, DPO, Selesai")
		return
	}
	id := pathID(r)
	existing, err := repo.GetLaporanByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := h.authorizeLaporanScope(r, existing.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengubah status laporan satuan kerja tersebut")
		return
	}
	if err := repo.UpdateStatusKasus(h.DB, id, req.StatusBaru, req.Keterangan, parseID(claims.Subject)); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Status kasus diperbarui"})
}

// DELETE /api/v1/laporan/{id}
func (h *LaporanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetLaporanByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := h.authorizeLaporanScope(r, existing.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menghapus laporan satuan kerja tersebut")
		return
	}
	if err := repo.DeleteLaporan(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Laporan dihapus"})
}
