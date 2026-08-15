package repo

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"curanmor-ai/internal/models"

	"github.com/lib/pq"
)

func CreateLogInteraksi(db *sql.DB, l *models.LogInteraksiWhatsapp) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO log_interaksi_whatsapp (no_whatsapp, pengguna_id, jenis_input, isi_pesan_masuk, status_akses)
		VALUES ($1,$2,$3,$4,$5) RETURNING id
	`, l.NoWhatsapp, l.PenggunaID, l.JenisInput, l.IsiPesanMasuk, l.StatusAkses).Scan(&id)
	return id, err
}

func GetLogInteraksiByID(db *sql.DB, id int64) (*models.LogInteraksiWhatsapp, error) {
	var l models.LogInteraksiWhatsapp
	err := db.QueryRow(`SELECT id, no_whatsapp, pengguna_id, jenis_input, COALESCE(isi_pesan_masuk,''), status_akses, waktu_masuk FROM log_interaksi_whatsapp WHERE id = $1`, id).
		Scan(&l.ID, &l.NoWhatsapp, &l.PenggunaID, &l.JenisInput, &l.IsiPesanMasuk, &l.StatusAkses, &l.WaktuMasuk)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func CreateHasilOcr(db *sql.DB, h *models.HasilOcrAI) (int64, error) {
	var jsonBytes []byte
	if h.HasilEkstraksiJSON != nil {
		var err error
		jsonBytes, err = json.Marshal(h.HasilEkstraksiJSON)
		if err != nil {
			return 0, err
		}
	}
	var id int64
	err := db.QueryRow(`
		INSERT INTO hasil_ocr_ai (log_id, jenis_proses, teks_mentah, skor_keyakinan, hasil_ekstraksi_json)
		VALUES ($1,$2,$3,$4,$5) RETURNING id
	`, h.LogID, h.JenisProses, h.TeksMentah, h.SkorKeyakinan, nullableJSON(jsonBytes)).Scan(&id)
	return id, err
}

func nullableJSON(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

func CreateLogPencarian(db *sql.DB, l *models.LogPencarian) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO log_pencarian (log_id, tipe_pencarian, kata_kunci, ditemukan, kendaraan_id, waktu_respons_ms)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id
	`, l.LogID, l.TipePencarian, l.KataKunci, l.Ditemukan, l.KendaraanID, l.WaktuResponsMs).Scan(&id)
	return id, err
}

func CreateResponsAI(db *sql.DB, r *models.ResponsAI) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO respons_ai (log_id, isi_respons, status_pengiriman)
		VALUES ($1,$2,$3) RETURNING id
	`, r.LogID, r.IsiRespons, r.StatusPengiriman).Scan(&id)
	return id, err
}

func GetResponsAIByID(db *sql.DB, id int64) (*models.ResponsAI, error) {
	var r models.ResponsAI
	err := db.QueryRow(`SELECT id, log_id, isi_respons, waktu_kirim, status_pengiriman FROM respons_ai WHERE id = $1`, id).
		Scan(&r.ID, &r.LogID, &r.IsiRespons, &r.WaktuKirim, &r.StatusPengiriman)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

type LogInteraksiFilter struct {
	NoWhatsapp string
	// SatkerScope, jika non-nil, membatasi hasil ke interaksi milik pengguna
	// yang terdaftar di salah satu satker ini (lihat repo.ResolveSatkerScope).
	// nil berarti tanpa filter satker (dipakai bridge WA lintas-satker).
	SatkerScope []int64
	Page        int
	PageSize    int
}

func ListLogInteraksi(db *sql.DB, f LogInteraksiFilter) ([]models.LogInteraksiWhatsapp, int, error) {
	from := "log_interaksi_whatsapp l"
	where := "1=1"
	args := []interface{}{}
	if f.NoWhatsapp != "" {
		args = append(args, f.NoWhatsapp)
		where += ` AND l.no_whatsapp = $` + strconv.Itoa(len(args))
	}
	if f.SatkerScope != nil {
		from = "log_interaksi_whatsapp l JOIN pengguna p ON p.id = l.pengguna_id"
		args = append(args, pq.Array(f.SatkerScope))
		where += ` AND p.satker_id = ANY($` + strconv.Itoa(len(args)) + `)`
	}
	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM `+from+` WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	page, pageSize := f.Page, f.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	query := `SELECT l.id, l.no_whatsapp, l.pengguna_id, l.jenis_input, COALESCE(l.isi_pesan_masuk,''), l.status_akses, l.waktu_masuk
		FROM ` + from + ` WHERE ` + where + ` ORDER BY l.waktu_masuk DESC LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []models.LogInteraksiWhatsapp
	for rows.Next() {
		var l models.LogInteraksiWhatsapp
		if err := rows.Scan(&l.ID, &l.NoWhatsapp, &l.PenggunaID, &l.JenisInput, &l.IsiPesanMasuk, &l.StatusAkses, &l.WaktuMasuk); err != nil {
			return nil, 0, err
		}
		out = append(out, l)
	}
	return out, total, rows.Err()
}
