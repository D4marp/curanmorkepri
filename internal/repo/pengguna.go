package repo

import (
	"database/sql"

	"github.com/lib/pq"

	"curanmor-ai/internal/models"
)

const penggunaSelectCols = `
	pg.id, pg.nrp, pg.nama_lengkap, COALESCE(pg.pangkat,''), COALESCE(pg.jabatan,''),
	pg.peran_id, pr.nama_peran, pg.satker_id, sk.nama_satker, pg.no_whatsapp,
	pg.status_aktif, pg.dibuat_oleh, pg.created_at, pg.updated_at
`

func scanPengguna(row interface{ Scan(...interface{}) error }) (*models.Pengguna, error) {
	var p models.Pengguna
	err := row.Scan(&p.ID, &p.NRP, &p.NamaLengkap, &p.Pangkat, &p.Jabatan,
		&p.PeranID, &p.PeranNama, &p.SatkerID, &p.SatkerNama, &p.NoWhatsapp,
		&p.StatusAktif, &p.DibuatOleh, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func ListPengguna(db *sql.DB, satkerScope []int64) ([]models.Pengguna, error) {
	query := `SELECT ` + penggunaSelectCols + ` FROM pengguna pg
		JOIN peran pr ON pr.id = pg.peran_id
		JOIN satuan_kerja sk ON sk.id = pg.satker_id`
	var rows *sql.Rows
	var err error
	if satkerScope != nil {
		query += ` WHERE pg.satker_id = ANY($1) ORDER BY pg.nama_lengkap`
		rows, err = db.Query(query, pq.Array(satkerScope))
	} else {
		query += ` ORDER BY pg.nama_lengkap`
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Pengguna
	for rows.Next() {
		p, err := scanPengguna(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func GetPenggunaByID(db *sql.DB, id int64) (*models.Pengguna, error) {
	row := db.QueryRow(`SELECT `+penggunaSelectCols+` FROM pengguna pg
		JOIN peran pr ON pr.id = pg.peran_id
		JOIN satuan_kerja sk ON sk.id = pg.satker_id
		WHERE pg.id = $1`, id)
	return scanPengguna(row)
}

// GetPenggunaByNRP juga mengembalikan password_hash — HANYA dipakai internal
// oleh handler login, tidak pernah diekspos ke response API.
func GetPenggunaByNRP(db *sql.DB, nrp string) (*models.Pengguna, error) {
	var p models.Pengguna
	err := db.QueryRow(`
		SELECT pg.id, pg.nrp, pg.nama_lengkap, COALESCE(pg.pangkat,''), COALESCE(pg.jabatan,''),
			pg.peran_id, pr.nama_peran, pg.satker_id, sk.nama_satker, pg.no_whatsapp,
			pg.status_aktif, pg.password_hash, sk.jenis_satker
		FROM pengguna pg
		JOIN peran pr ON pr.id = pg.peran_id
		JOIN satuan_kerja sk ON sk.id = pg.satker_id
		WHERE pg.nrp = $1
	`, nrp).Scan(&p.ID, &p.NRP, &p.NamaLengkap, &p.Pangkat, &p.Jabatan,
		&p.PeranID, &p.PeranNama, &p.SatkerID, &p.SatkerNama, &p.NoWhatsapp,
		&p.StatusAktif, &p.PasswordHash, &p.JenisSatker)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func CreatePengguna(db *sql.DB, p *models.Pengguna) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO pengguna (nrp, nama_lengkap, pangkat, jabatan, peran_id, satker_id, no_whatsapp, password_hash, dibuat_oleh)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id
	`, p.NRP, p.NamaLengkap, p.Pangkat, p.Jabatan, p.PeranID, p.SatkerID, p.NoWhatsapp, p.PasswordHash, p.DibuatOleh).Scan(&id)
	return id, err
}

func UpdatePengguna(db *sql.DB, p *models.Pengguna) error {
	_, err := db.Exec(`
		UPDATE pengguna SET nama_lengkap=$1, pangkat=$2, jabatan=$3, peran_id=$4, satker_id=$5,
			no_whatsapp=$6, status_aktif=$7 WHERE id=$8
	`, p.NamaLengkap, p.Pangkat, p.Jabatan, p.PeranID, p.SatkerID, p.NoWhatsapp, p.StatusAktif, p.ID)
	return err
}

func UpdatePenggunaPassword(db *sql.DB, id int64, hash string) error {
	_, err := db.Exec(`UPDATE pengguna SET password_hash = $1 WHERE id = $2`, hash, id)
	return err
}

func DeletePengguna(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM pengguna WHERE id = $1`, id)
	return err
}
