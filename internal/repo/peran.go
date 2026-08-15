package repo

import (
	"database/sql"

	"curanmor-ai/internal/models"
)

func ListPeran(db *sql.DB) ([]models.Peran, error) {
	rows, err := db.Query(`SELECT id, nama_peran, COALESCE(deskripsi,'') FROM peran ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Peran
	for rows.Next() {
		var p models.Peran
		if err := rows.Scan(&p.ID, &p.NamaPeran, &p.Deskripsi); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func ListModulAkses(db *sql.DB) ([]models.ModulAkses, error) {
	rows, err := db.Query(`SELECT id, kode_modul, nama_modul FROM modul_akses ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ModulAkses
	for rows.Next() {
		var m models.ModulAkses
		if err := rows.Scan(&m.ID, &m.KodeModul, &m.NamaModul); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type MatriksRow struct {
	PeranID    int64  `json:"peran_id"`
	PeranNama  string `json:"peran_nama"`
	ModulID    int64  `json:"modul_id"`
	KodeModul  string `json:"kode_modul"`
	NamaModul  string `json:"nama_modul"`
	LevelAkses string `json:"level_akses"`
}

func ListMatriksRBAC(db *sql.DB) ([]MatriksRow, error) {
	rows, err := db.Query(`
		SELECT p.id, p.nama_peran, m.id, m.kode_modul, m.nama_modul, pma.level_akses
		FROM peran_modul_akses pma
		JOIN peran p ON p.id = pma.peran_id
		JOIN modul_akses m ON m.id = pma.modul_id
		ORDER BY p.id, m.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MatriksRow
	for rows.Next() {
		var r MatriksRow
		if err := rows.Scan(&r.PeranID, &r.PeranNama, &r.ModulID, &r.KodeModul, &r.NamaModul, &r.LevelAkses); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpsertPeranModulAkses dipakai oleh modul "Pengaturan Sistem" (Super Admin
// saja) untuk mengubah matriks RBAC.
func UpsertPeranModulAkses(db *sql.DB, peranID, modulID int64, level string) error {
	_, err := db.Exec(`
		INSERT INTO peran_modul_akses (peran_id, modul_id, level_akses) VALUES ($1,$2,$3)
		ON CONFLICT (peran_id, modul_id) DO UPDATE SET level_akses = EXCLUDED.level_akses
	`, peranID, modulID, level)
	return err
}
