package repo

import (
	"database/sql"

	"curanmor-ai/internal/models"
)

func ListRiwayatStatusByLP(db *sql.DB, lpID int64) ([]models.RiwayatStatusKasus, error) {
	rows, err := db.Query(`
		SELECT r.id, r.lp_id, COALESCE(r.status_sebelumnya,''), r.status_baru, r.tanggal_perubahan,
			COALESCE(r.keterangan,''), r.diubah_oleh, COALESCE(pg.nama_lengkap,'')
		FROM riwayat_status_kasus r
		LEFT JOIN pengguna pg ON pg.id = r.diubah_oleh
		WHERE r.lp_id = $1 ORDER BY r.tanggal_perubahan DESC
	`, lpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.RiwayatStatusKasus
	for rows.Next() {
		var r models.RiwayatStatusKasus
		if err := rows.Scan(&r.ID, &r.LPID, &r.StatusSebelumnya, &r.StatusBaru, &r.TanggalPerubahan, &r.Keterangan, &r.DiubahOleh, &r.DiubahOlehNama); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func ListRiwayatPenyidikanByLP(db *sql.DB, lpID int64) ([]models.RiwayatPenyidikan, error) {
	rows, err := db.Query(`SELECT id, lp_id, tanggal, COALESCE(kegiatan,''), penyidik_id, COALESCE(catatan,'') FROM riwayat_penyidikan WHERE lp_id = $1 ORDER BY tanggal DESC`, lpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.RiwayatPenyidikan
	for rows.Next() {
		var r models.RiwayatPenyidikan
		if err := rows.Scan(&r.ID, &r.LPID, &r.Tanggal, &r.Kegiatan, &r.PenyidikID, &r.Catatan); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetRiwayatPenyidikanByID mengambil satu baris riwayat_penyidikan termasuk
// lp_id — dipakai handler untuk resolusi otorisasi cakupan satker sebelum
// hapus lewat endpoint per-ID.
func GetRiwayatPenyidikanByID(db *sql.DB, id int64) (*models.RiwayatPenyidikan, error) {
	var r models.RiwayatPenyidikan
	err := db.QueryRow(`SELECT id, lp_id, tanggal, COALESCE(kegiatan,''), penyidik_id, COALESCE(catatan,'') FROM riwayat_penyidikan WHERE id = $1`, id).
		Scan(&r.ID, &r.LPID, &r.Tanggal, &r.Kegiatan, &r.PenyidikID, &r.Catatan)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateRiwayatPenyidikan(db *sql.DB, r *models.RiwayatPenyidikan) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO riwayat_penyidikan (lp_id, tanggal, kegiatan, penyidik_id, catatan)
		VALUES ($1,$2,$3,$4,$5) RETURNING id
	`, r.LPID, r.Tanggal, r.Kegiatan, r.PenyidikID, r.Catatan).Scan(&id)
	return id, err
}

func DeleteRiwayatPenyidikan(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM riwayat_penyidikan WHERE id = $1`, id)
	return err
}
