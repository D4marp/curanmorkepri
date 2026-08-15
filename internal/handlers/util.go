package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"curanmor-ai/internal/repo"
)

func itoa64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func parseID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}

// pathID mengambil parameter path {id} (Go 1.22+ net/http routing).
func pathID(r *http.Request) int64 {
	return parseID(r.PathValue("id"))
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func queryIntPtr64(r *http.Request, key string) *int64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

// authorizeLPScope memastikan satker pemilik laporan (lp_id) berada dalam
// cakupan satuan kerja pengguna yang login saat ini. Dipakai oleh seluruh
// handler entitas anak laporan_polisi (kendaraan, pihak_terkait,
// barang_bukti, dokumentasi_pendukung, riwayat_penyidikan) agar akses
// tidak bisa lintas-tenant hanya dengan menebak ID entitas anak — celah
// otorisasi kritis yang sama seperti pada endpoint per-ID laporan.
func authorizeLPScope(db *sql.DB, r *http.Request, lpID int64) (bool, error) {
	lp, err := repo.GetLaporanByID(db, lpID)
	if err != nil {
		return false, err
	}
	return authorizeSatkerScope(db, r, lp.SatkerID)
}
