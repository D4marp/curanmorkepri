package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

// EvidenceHandler menangani barang_bukti, dokumentasi_pendukung, dan
// riwayat_penyidikan — dikelompokkan dalam satu handler karena ketiganya
// adalah entitas anak sederhana (create/delete) di bawah laporan_polisi.
type EvidenceHandler struct{ DB *sql.DB }

// ---------- barang_bukti ----------

// POST /api/v1/laporan/{id}/barang-bukti
func (h *EvidenceHandler) CreateBarangBukti(w http.ResponseWriter, r *http.Request) {
	lpID := pathID(r)
	if ok, err := authorizeLPScope(h.DB, r, lpID); err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	} else if err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menambah barang bukti pada laporan satuan kerja tersebut")
		return
	}
	var b models.BarangBukti
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	b.LPID = lpID
	id, err := repo.CreateBarangBukti(h.DB, &b)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	b.ID = id
	httpx.Created(w, b)
}

// DELETE /api/v1/barang-bukti/{id}
func (h *EvidenceHandler) DeleteBarangBukti(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetBarangBuktiByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Barang bukti tidak ditemukan")
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
		httpx.Forbidden(w, "Anda tidak berwenang menghapus barang bukti satuan kerja tersebut")
		return
	}
	if err := repo.DeleteBarangBukti(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Barang bukti dihapus"})
}

// ---------- dokumentasi_pendukung ----------

// POST /api/v1/laporan/{id}/dokumentasi
func (h *EvidenceHandler) CreateDokumentasi(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	lpID := pathID(r)
	if ok, err := authorizeLPScope(h.DB, r, lpID); err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	} else if err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menambah dokumentasi pada laporan satuan kerja tersebut")
		return
	}
	var d models.DokumentasiPendukung
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if d.FileURL == "" || d.JenisDokumen == "" {
		httpx.BadRequest(w, "file_url dan jenis_dokumen wajib diisi")
		return
	}
	d.LPID = lpID
	uid := parseID(claims.Subject)
	d.DiunggahOleh = &uid
	id, err := repo.CreateDokumentasi(h.DB, &d)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	created, err := repo.GetDokumentasiByID(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

// DELETE /api/v1/dokumentasi/{id}
func (h *EvidenceHandler) DeleteDokumentasi(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetDokumentasiByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Dokumentasi tidak ditemukan")
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
		httpx.Forbidden(w, "Anda tidak berwenang menghapus dokumentasi satuan kerja tersebut")
		return
	}
	if err := repo.DeleteDokumentasi(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Dokumentasi dihapus"})
}

// ---------- riwayat_penyidikan ----------

// POST /api/v1/laporan/{id}/riwayat-penyidikan
func (h *EvidenceHandler) CreateRiwayatPenyidikan(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	lpID := pathID(r)
	if ok, err := authorizeLPScope(h.DB, r, lpID); err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	} else if err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menambah riwayat penyidikan pada laporan satuan kerja tersebut")
		return
	}
	var rp models.RiwayatPenyidikan
	if err := json.NewDecoder(r.Body).Decode(&rp); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if rp.Tanggal == "" {
		httpx.BadRequest(w, "tanggal wajib diisi")
		return
	}
	rp.LPID = lpID
	if rp.PenyidikID == nil {
		pid := parseID(claims.Subject)
		rp.PenyidikID = &pid
	}
	id, err := repo.CreateRiwayatPenyidikan(h.DB, &rp)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	rp.ID = id
	httpx.Created(w, rp)
}

// DELETE /api/v1/riwayat-penyidikan/{id}
func (h *EvidenceHandler) DeleteRiwayatPenyidikan(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetRiwayatPenyidikanByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Riwayat penyidikan tidak ditemukan")
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
		httpx.Forbidden(w, "Anda tidak berwenang menghapus riwayat penyidikan satuan kerja tersebut")
		return
	}
	if err := repo.DeleteRiwayatPenyidikan(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Riwayat penyidikan dihapus"})
}
