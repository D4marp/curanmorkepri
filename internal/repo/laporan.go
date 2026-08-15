package repo

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"curanmor-ai/internal/models"
)

type LaporanFilter struct {
	SatkerScope   []int64 // hasil ResolveSatkerScope; nil = tanpa batas
	SatkerID      *int64  // filter tambahan dari query user (mis. pilih 1 wilayah spesifik)
	StatusKasus   string
	JenisPerkara  string
	DariTanggal   string // YYYY-MM-DD, terhadap tanggal_kejadian
	SampaiTanggal string
	Keyword       string // cocok ke no_lp atau tkp_alamat
	Page          int
	PageSize      int
}

const laporanListCols = `
	lp.id, lp.no_lp, lp.tanggal_lp, lp.jenis_perkara, lp.satker_id, sk.nama_satker,
	lp.penyidik_id, COALESCE(pg.nama_lengkap,''), lp.tanggal_kejadian, COALESCE(lp.tkp_alamat,''),
	lp.tkp_latitude, lp.tkp_longitude, lp.status_kasus, lp.created_at, lp.updated_at
`

func scanLaporanRow(row interface{ Scan(...interface{}) error }) (*models.LaporanPolisi, error) {
	var lp models.LaporanPolisi
	var tglKejadian sql.NullString
	err := row.Scan(&lp.ID, &lp.NoLP, &lp.TanggalLP, &lp.JenisPerkara, &lp.SatkerID, &lp.SatkerNama,
		&lp.PenyidikID, &lp.PenyidikNama, &tglKejadian, &lp.TKPAlamat,
		&lp.TKPLatitude, &lp.TKPLongitude, &lp.StatusKasus, &lp.CreatedAt, &lp.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if tglKejadian.Valid {
		lp.TanggalKejadian = &tglKejadian.String
	}
	return &lp, nil
}

func SearchLaporan(db *sql.DB, f LaporanFilter) ([]models.LaporanPolisi, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1

	add := func(cond string, val interface{}) {
		where = append(where, fmt.Sprintf(cond, argN))
		args = append(args, val)
		argN++
	}

	if f.SatkerScope != nil {
		add("lp.satker_id = ANY($%d)", pq.Array(f.SatkerScope))
	}
	if f.SatkerID != nil {
		add("lp.satker_id = $%d", *f.SatkerID)
	}
	if f.StatusKasus != "" {
		add("lp.status_kasus = $%d", f.StatusKasus)
	}
	if f.JenisPerkara != "" {
		add("lp.jenis_perkara = $%d", f.JenisPerkara)
	}
	if f.DariTanggal != "" {
		add("lp.tanggal_kejadian >= $%d", f.DariTanggal)
	}
	if f.SampaiTanggal != "" {
		add("lp.tanggal_kejadian <= $%d", f.SampaiTanggal)
	}
	if f.Keyword != "" {
		where = append(where, fmt.Sprintf("(lp.no_lp ILIKE $%d OR lp.tkp_alamat ILIKE $%d)", argN, argN))
		args = append(args, "%"+f.Keyword+"%")
		argN++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := `SELECT COUNT(*) FROM laporan_polisi lp WHERE ` + whereClause
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	query := `SELECT ` + laporanListCols + `
		FROM laporan_polisi lp
		JOIN satuan_kerja sk ON sk.id = lp.satker_id
		LEFT JOIN pengguna pg ON pg.id = lp.penyidik_id
		WHERE ` + whereClause + `
		ORDER BY lp.tanggal_lp DESC, lp.id DESC
		LIMIT $` + fmt.Sprint(argN) + ` OFFSET $` + fmt.Sprint(argN+1)
	args = append(args, pageSize, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LaporanPolisi
	for rows.Next() {
		lp, err := scanLaporanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *lp)
	}
	return out, total, rows.Err()
}

func GetLaporanByID(db *sql.DB, id int64) (*models.LaporanPolisi, error) {
	row := db.QueryRow(`SELECT `+laporanListCols+`
		FROM laporan_polisi lp
		JOIN satuan_kerja sk ON sk.id = lp.satker_id
		LEFT JOIN pengguna pg ON pg.id = lp.penyidik_id
		WHERE lp.id = $1`, id)
	return scanLaporanRow(row)
}

func GetLaporanByNoLP(db *sql.DB, noLP string) (*models.LaporanPolisi, error) {
	row := db.QueryRow(`SELECT `+laporanListCols+`
		FROM laporan_polisi lp
		JOIN satuan_kerja sk ON sk.id = lp.satker_id
		LEFT JOIN pengguna pg ON pg.id = lp.penyidik_id
		WHERE lp.no_lp = $1`, noLP)
	return scanLaporanRow(row)
}

func CreateLaporan(db *sql.DB, lp *models.LaporanPolisi) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO laporan_polisi (no_lp, tanggal_lp, jenis_perkara, satker_id, penyidik_id, tanggal_kejadian, tkp_alamat, tkp_latitude, tkp_longitude, status_kasus)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id
	`, lp.NoLP, lp.TanggalLP, lp.JenisPerkara, lp.SatkerID, lp.PenyidikID, lp.TanggalKejadian, lp.TKPAlamat, lp.TKPLatitude, lp.TKPLongitude, lp.StatusKasus).Scan(&id)
	return id, err
}

// UpdateLaporan mengubah data laporan. penyidik_id dan tanggal_kejadian memakai
// COALESCE agar permintaan yang tidak menyertakan field tersebut (mis. formulir
// UI yang hanya mengubah TKP/tanggal LP) TIDAK menghapus penyidik yang sudah
// ditugaskan sebelumnya — mencegah kehilangan data penting secara diam-diam.
func UpdateLaporan(db *sql.DB, lp *models.LaporanPolisi) error {
	_, err := db.Exec(`
		UPDATE laporan_polisi SET
			tanggal_lp=COALESCE(NULLIF($1, '')::date, tanggal_lp),
			jenis_perkara=COALESCE(NULLIF($2, '')::jenis_perkara_enum, jenis_perkara),
			satker_id=$3,
			penyidik_id=COALESCE($4, penyidik_id),
			tanggal_kejadian=COALESCE($5, tanggal_kejadian),
			tkp_alamat=COALESCE(NULLIF($6, ''), tkp_alamat),
			tkp_latitude=COALESCE($7, tkp_latitude),
			tkp_longitude=COALESCE($8, tkp_longitude)
		WHERE id=$9
	`, lp.TanggalLP, lp.JenisPerkara, lp.SatkerID, lp.PenyidikID, lp.TanggalKejadian, lp.TKPAlamat, lp.TKPLatitude, lp.TKPLongitude, lp.ID)
	return err
}

// UpdateStatusKasus mengubah status_kasus; trigger DB (trg_laporan_status_history)
// otomatis mencatat ke riwayat_status_kasus. keterangan opsional ditambahkan
// manual sebagai baris riwayat terpisah agar tercatat siapa & kenapa.
func UpdateStatusKasus(db *sql.DB, id int64, statusBaru string, keterangan string, diubahOleh int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE laporan_polisi SET status_kasus = $1 WHERE id = $2`, statusBaru, id); err != nil {
		_ = tx.Rollback()
		return err
	}
	if keterangan != "" {
		if _, err := tx.Exec(`
			UPDATE riwayat_status_kasus SET keterangan = $1, diubah_oleh = $2
			WHERE id = (SELECT id FROM riwayat_status_kasus WHERE lp_id = $3 ORDER BY id DESC LIMIT 1)
		`, keterangan, diubahOleh, id); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func DeleteLaporan(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM laporan_polisi WHERE id = $1`, id)
	return err
}
