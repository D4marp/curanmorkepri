package repo

import (
	"database/sql"

	"curanmor-ai/internal/models"
)

func ListSatuanKerja(db *sql.DB) ([]models.SatuanKerja, error) {
	rows, err := db.Query(`SELECT id, kode_satker, nama_satker, jenis_satker, induk_id, COALESCE(wilayah,''), COALESCE(alamat,''), created_at FROM satuan_kerja ORDER BY jenis_satker, nama_satker`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.SatuanKerja
	for rows.Next() {
		var s models.SatuanKerja
		if err := rows.Scan(&s.ID, &s.KodeSatker, &s.NamaSatker, &s.JenisSatker, &s.IndukID, &s.Wilayah, &s.Alamat, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func GetSatuanKerja(db *sql.DB, id int64) (*models.SatuanKerja, error) {
	var s models.SatuanKerja
	err := db.QueryRow(`SELECT id, kode_satker, nama_satker, jenis_satker, induk_id, COALESCE(wilayah,''), COALESCE(alamat,''), created_at FROM satuan_kerja WHERE id = $1`, id).
		Scan(&s.ID, &s.KodeSatker, &s.NamaSatker, &s.JenisSatker, &s.IndukID, &s.Wilayah, &s.Alamat, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func CreateSatuanKerja(db *sql.DB, s *models.SatuanKerja) (int64, error) {
	var id int64
	err := db.QueryRow(
		`INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah, alamat) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		s.KodeSatker, s.NamaSatker, s.JenisSatker, s.IndukID, s.Wilayah, s.Alamat,
	).Scan(&id)
	return id, err
}

func UpdateSatuanKerja(db *sql.DB, s *models.SatuanKerja) error {
	_, err := db.Exec(
		`UPDATE satuan_kerja SET nama_satker=$1, jenis_satker=$2, induk_id=$3, wilayah=$4, alamat=$5 WHERE id=$6`,
		s.NamaSatker, s.JenisSatker, s.IndukID, s.Wilayah, s.Alamat, s.ID,
	)
	return err
}

func DeleteSatuanKerja(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM satuan_kerja WHERE id = $1`, id)
	return err
}
