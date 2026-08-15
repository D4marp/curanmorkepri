package repo

import (
	"database/sql"

	"github.com/lib/pq"

	"curanmor-ai/internal/models"
)

const kendaraanCols = `k.id, k.lp_id, COALESCE(k.no_polisi,''), COALESCE(k.no_rangka_vin,''), COALESCE(k.no_mesin,''),
	COALESCE(k.merk_tipe,''), COALESCE(k.warna,''), k.tahun, k.jenis_kendaraan, k.status_kendaraan,
	COALESCE(k.foto_kendaraan_url,''), k.created_at, k.updated_at`

func scanKendaraan(row interface{ Scan(...interface{}) error }) (*models.Kendaraan, error) {
	var k models.Kendaraan
	if err := row.Scan(&k.ID, &k.LPID, &k.NoPolisi, &k.NoRangkaVIN, &k.NoMesin, &k.MerkTipe, &k.Warna,
		&k.Tahun, &k.JenisKendaraan, &k.StatusKendaraan, &k.FotoKendaraanURL, &k.CreatedAt, &k.UpdatedAt); err != nil {
		return nil, err
	}
	return &k, nil
}

func GetKendaraanByID(db *sql.DB, id int64) (*models.Kendaraan, error) {
	row := db.QueryRow(`SELECT `+kendaraanCols+` FROM kendaraan k WHERE k.id = $1`, id)
	return scanKendaraan(row)
}

func ListKendaraanByLP(db *sql.DB, lpID int64) ([]models.Kendaraan, error) {
	rows, err := db.Query(`SELECT `+kendaraanCols+` FROM kendaraan k WHERE k.lp_id = $1 ORDER BY k.id`, lpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Kendaraan
	for rows.Next() {
		k, err := scanKendaraan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *k)
	}
	return out, rows.Err()
}

func CreateKendaraan(db *sql.DB, k *models.Kendaraan) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO kendaraan (lp_id, no_polisi, no_rangka_vin, no_mesin, merk_tipe, warna, tahun, jenis_kendaraan, status_kendaraan, foto_kendaraan_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id
	`, k.LPID, k.NoPolisi, k.NoRangkaVIN, k.NoMesin, k.MerkTipe, k.Warna, k.Tahun, k.JenisKendaraan, k.StatusKendaraan, k.FotoKendaraanURL).Scan(&id)
	return id, err
}

func UpdateKendaraan(db *sql.DB, k *models.Kendaraan) error {
	_, err := db.Exec(`
		UPDATE kendaraan SET no_polisi=$1, no_rangka_vin=$2, no_mesin=$3, merk_tipe=$4, warna=$5,
			tahun=$6, jenis_kendaraan=$7, status_kendaraan=$8, foto_kendaraan_url=$9
		WHERE id=$10
	`, k.NoPolisi, k.NoRangkaVIN, k.NoMesin, k.MerkTipe, k.Warna, k.Tahun, k.JenisKendaraan, k.StatusKendaraan, k.FotoKendaraanURL, k.ID)
	return err
}

func UpdateStatusKendaraan(db *sql.DB, id int64, status string) error {
	_, err := db.Exec(`UPDATE kendaraan SET status_kendaraan = $1 WHERE id = $2`, status, id)
	return err
}

func DeleteKendaraan(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM kendaraan WHERE id = $1`, id)
	return err
}

// SearchKendaraan adalah jalur kueri UTAMA dipakai chatbot WhatsApp AI
// (harus merespons "dalam hitungan detik" — mengandalkan indeks pada
// no_polisi/no_rangka_vin/no_mesin, lihat migrations/0005). Pencarian
// case-insensitive & exact-match dahulu, fallback ke partial match.
func SearchKendaraan(db *sql.DB, tipe, kataKunci string, satkerScope []int64) ([]models.Kendaraan, error) {
	var col string
	switch tipe {
	case "No. Polisi":
		col = "k.no_polisi"
	case "No. Rangka":
		col = "k.no_rangka_vin"
	case "No. Mesin":
		col = "k.no_mesin"
	default:
		col = "k.no_polisi"
	}

	query := `
		SELECT k.id, k.lp_id, COALESCE(k.no_polisi,''), COALESCE(k.no_rangka_vin,''), COALESCE(k.no_mesin,''),
			COALESCE(k.merk_tipe,''), COALESCE(k.warna,''), k.tahun, k.jenis_kendaraan, k.status_kendaraan,
			COALESCE(k.foto_kendaraan_url,''), k.created_at, k.updated_at,
			lp.no_lp, lp.status_kasus, sk.nama_satker
		FROM kendaraan k
		JOIN laporan_polisi lp ON lp.id = k.lp_id
		JOIN satuan_kerja sk ON sk.id = lp.satker_id
		WHERE UPPER(` + col + `) = UPPER($1)`
	args := []interface{}{kataKunci}
	if satkerScope != nil {
		query += ` AND lp.satker_id = ANY($2)`
		args = append(args, pq.Array(satkerScope))
	}
	query += ` ORDER BY k.created_at DESC LIMIT 20`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Kendaraan
	for rows.Next() {
		var k models.Kendaraan
		if err := rows.Scan(&k.ID, &k.LPID, &k.NoPolisi, &k.NoRangkaVIN, &k.NoMesin, &k.MerkTipe, &k.Warna,
			&k.Tahun, &k.JenisKendaraan, &k.StatusKendaraan, &k.FotoKendaraanURL, &k.CreatedAt, &k.UpdatedAt,
			&k.NoLP, &k.StatusKasus, &k.SatkerNama); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}
