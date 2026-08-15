package repo

import (
	"database/sql"

	"github.com/lib/pq"

	"curanmor-ai/internal/models"
)

// ListWhitelist mengembalikan entri whitelist_whatsapp, dibatasi ke satker
// pemilik akun pengguna terkait bila satkerScope tidak nil (Admin
// Polres/Polsek) — Admin Polda (satkerScope nil) melihat semuanya.
func ListWhitelist(db *sql.DB, satkerScope []int64) ([]models.WhitelistWhatsapp, error) {
	query := `SELECT ww.id, ww.no_whatsapp, ww.pengguna_id, ww.status, ww.tanggal_registrasi, ww.diregistrasi_oleh, ww.tanggal_verifikasi
		FROM whitelist_whatsapp ww JOIN pengguna pg ON pg.id = ww.pengguna_id`
	var rows *sql.Rows
	var err error
	if satkerScope != nil {
		query += ` WHERE pg.satker_id = ANY($1) ORDER BY ww.tanggal_registrasi DESC`
		rows, err = db.Query(query, pq.Array(satkerScope))
	} else {
		query += ` ORDER BY ww.tanggal_registrasi DESC`
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.WhitelistWhatsapp
	for rows.Next() {
		var w models.WhitelistWhatsapp
		if err := rows.Scan(&w.ID, &w.NoWhatsapp, &w.PenggunaID, &w.Status, &w.TanggalRegistrasi, &w.DiregistrasiOleh, &w.TanggalVerifikasi); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func GetWhitelistByID(db *sql.DB, id int64) (*models.WhitelistWhatsapp, error) {
	var w models.WhitelistWhatsapp
	err := db.QueryRow(`SELECT id, no_whatsapp, pengguna_id, status, tanggal_registrasi, diregistrasi_oleh, tanggal_verifikasi FROM whitelist_whatsapp WHERE id = $1`, id).
		Scan(&w.ID, &w.NoWhatsapp, &w.PenggunaID, &w.Status, &w.TanggalRegistrasi, &w.DiregistrasiOleh, &w.TanggalVerifikasi)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func CreateWhitelist(db *sql.DB, w *models.WhitelistWhatsapp) (int64, error) {
	var id int64
	err := db.QueryRow(
		`INSERT INTO whitelist_whatsapp (no_whatsapp, pengguna_id, status, diregistrasi_oleh) VALUES ($1,$2,$3,$4) RETURNING id`,
		w.NoWhatsapp, w.PenggunaID, w.Status, w.DiregistrasiOleh,
	).Scan(&id)
	return id, err
}

func UpdateWhitelistStatus(db *sql.DB, id int64, status string) error {
	verifCol := ""
	if status == "aktif" {
		verifCol = ", tanggal_verifikasi = now()"
	}
	_, err := db.Exec(`UPDATE whitelist_whatsapp SET status = $1`+verifCol+` WHERE id = $2`, status, id)
	return err
}

func DeleteWhitelist(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM whitelist_whatsapp WHERE id = $1`, id)
	return err
}

// CheckWhitelist dipakai oleh /api/v1/wa/whitelist/check — dipanggil chatbot
// AI SEBELUM memproses pesan, sesuai alur "whitelist nomor WA" pada paparan.
type WhitelistCheckResult struct {
	Diizinkan   bool   `json:"diizinkan"`
	PenggunaID  *int64 `json:"pengguna_id,omitempty"`
	NamaLengkap string `json:"nama_lengkap,omitempty"`
	PeranNama   string `json:"peran_nama,omitempty"`
	SatkerNama  string `json:"satker_nama,omitempty"`
	Alasan      string `json:"alasan,omitempty"`
}

func CheckWhitelist(db *sql.DB, noWhatsapp string) (*WhitelistCheckResult, error) {
	var status string
	var penggunaID int64
	var nama, peranNama, satkerNama string
	var statusAktif bool
	err := db.QueryRow(`
		SELECT ww.status, pg.id, pg.nama_lengkap, pr.nama_peran, sk.nama_satker, pg.status_aktif
		FROM whitelist_whatsapp ww
		JOIN pengguna pg ON pg.id = ww.pengguna_id
		JOIN peran pr ON pr.id = pg.peran_id
		JOIN satuan_kerja sk ON sk.id = pg.satker_id
		WHERE ww.no_whatsapp = $1
	`, noWhatsapp).Scan(&status, &penggunaID, &nama, &peranNama, &satkerNama, &statusAktif)

	if err == sql.ErrNoRows {
		return &WhitelistCheckResult{Diizinkan: false, Alasan: "Nomor tidak terdaftar di whitelist"}, nil
	}
	if err != nil {
		return nil, err
	}
	if status != "aktif" || !statusAktif {
		return &WhitelistCheckResult{Diizinkan: false, Alasan: "Nomor/akun berstatus " + status}, nil
	}
	return &WhitelistCheckResult{
		Diizinkan: true, PenggunaID: &penggunaID, NamaLengkap: nama, PeranNama: peranNama, SatkerNama: satkerNama,
	}, nil
}
