package repo

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/lib/pq"

	"curanmor-ai/internal/models"
)

// ---------- notifikasi_alert ----------

func ListNotifikasiUntukPengguna(db *sql.DB, penggunaID, peranID int64) ([]models.NotifikasiAlert, error) {
	rows, err := db.Query(`
		SELECT id, jenis_alert, deskripsi, target_peran_id, target_pengguna_id, status_baca, created_at
		FROM notifikasi_alert
		WHERE target_pengguna_id = $1 OR target_pengguna_id IS NULL AND (target_peran_id = $2 OR target_peran_id IS NULL)
		ORDER BY created_at DESC LIMIT 100
	`, penggunaID, peranID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.NotifikasiAlert
	for rows.Next() {
		var n models.NotifikasiAlert
		if err := rows.Scan(&n.ID, &n.JenisAlert, &n.Deskripsi, &n.TargetPeranID, &n.TargetPenggunaID, &n.StatusBaca, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func CreateNotifikasi(db *sql.DB, n *models.NotifikasiAlert) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO notifikasi_alert (jenis_alert, deskripsi, target_peran_id, target_pengguna_id)
		VALUES ($1,$2,$3,$4) RETURNING id
	`, n.JenisAlert, n.Deskripsi, n.TargetPeranID, n.TargetPenggunaID).Scan(&id)
	return id, err
}

// MarkNotifikasiRead menandai notifikasi dibaca HANYA jika notifikasi tersebut
// memang ditujukan ke pengguna ini (target_pengguna_id cocok, atau broadcast
// ke peran/semua orang) — mencegah IDOR di mana pengguna mana pun bisa
// menandai-dibaca notifikasi milik pengguna lain hanya dengan menebak ID.
// sql.ErrNoRows dikembalikan bila notifikasi tidak ditemukan ATAU bukan milik
// pengguna ini, supaya handler tidak membocorkan keberadaan ID tersebut.
func MarkNotifikasiRead(db *sql.DB, id, penggunaID, peranID int64) error {
	res, err := db.Exec(`
		UPDATE notifikasi_alert SET status_baca = true
		WHERE id = $1
		  AND (target_pengguna_id = $2 OR target_pengguna_id IS NULL AND (target_peran_id = $3 OR target_peran_id IS NULL))
	`, id, penggunaID, peranID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ---------- laporan_periodik ----------

func ListLaporanPeriodik(db *sql.DB, satkerScope []int64) ([]models.LaporanPeriodik, error) {
	query := `SELECT id, jenis_laporan, periode_mulai, periode_selesai, satker_id, COALESCE(file_url,''), dibuat_otomatis, created_at FROM laporan_periodik`
	var rows *sql.Rows
	var err error
	if satkerScope != nil {
		query += ` WHERE satker_id = ANY($1) OR satker_id IS NULL ORDER BY created_at DESC`
		rows, err = db.Query(query, pq.Array(satkerScope))
	} else {
		query += ` ORDER BY created_at DESC`
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.LaporanPeriodik
	for rows.Next() {
		var l models.LaporanPeriodik
		if err := rows.Scan(&l.ID, &l.JenisLaporan, &l.PeriodeMulai, &l.PeriodeSelesai, &l.SatkerID, &l.FileURL, &l.DibuatOtomatis, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func GetLaporanPeriodikByID(db *sql.DB, id int64) (*models.LaporanPeriodik, error) {
	var l models.LaporanPeriodik
	err := db.QueryRow(`SELECT id, jenis_laporan, periode_mulai, periode_selesai, satker_id, COALESCE(file_url,''), dibuat_otomatis, created_at FROM laporan_periodik WHERE id = $1`, id).
		Scan(&l.ID, &l.JenisLaporan, &l.PeriodeMulai, &l.PeriodeSelesai, &l.SatkerID, &l.FileURL, &l.DibuatOtomatis, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func CreateLaporanPeriodik(db *sql.DB, l *models.LaporanPeriodik) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO laporan_periodik (jenis_laporan, periode_mulai, periode_selesai, satker_id, file_url, dibuat_otomatis)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id
	`, l.JenisLaporan, l.PeriodeMulai, l.PeriodeSelesai, l.SatkerID, l.FileURL, l.DibuatOtomatis).Scan(&id)
	return id, err
}

// ---------- audit_log ----------

// ListAuditLog mengambil jurnal audit, dibatasi cakupan satuan kerja
// (satkerScope nil = Admin Polda/Super Admin, tanpa batas). Tanpa batasan ini
// Admin Polres/Viewer (yang punya akses modul laporan_analitik) dapat melihat
// aktivitas pengguna di satuan kerja lain — kebocoran jurnal akuntabilitas
// lintas-tenant yang tidak sesuai desain multi-tenant sistem ini.
func ListAuditLog(db *sql.DB, satkerScope []int64, page, pageSize int) ([]models.AuditLog, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	where := ""
	countArgs := []interface{}{}
	queryArgs := []interface{}{}
	if satkerScope != nil {
		where = "WHERE pg.satker_id = ANY($1)"
		countArgs = append(countArgs, pq.Array(satkerScope))
		queryArgs = append(queryArgs, pq.Array(satkerScope))
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM audit_log a LEFT JOIN pengguna pg ON pg.id = a.pengguna_id ` + where
	if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryArgs = append(queryArgs, pageSize, offset)
	limitOffset := fmt.Sprintf("ORDER BY a.waktu DESC LIMIT $%d OFFSET $%d", len(queryArgs)-1, len(queryArgs))
	rows, err := db.Query(`
		SELECT a.id, a.pengguna_id, COALESCE(pg.nama_lengkap,'-'), a.aktivitas, COALESCE(a.modul,''), COALESCE(a.ip_address,''), COALESCE(a.perangkat,''), a.waktu
		FROM audit_log a
		LEFT JOIN pengguna pg ON pg.id = a.pengguna_id
		`+where+`
		`+limitOffset+`
	`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []models.AuditLog
	for rows.Next() {
		var a models.AuditLog
		if err := rows.Scan(&a.ID, &a.PenggunaID, &a.PenggunaNama, &a.Aktivitas, &a.Modul, &a.IPAddress, &a.Perangkat, &a.Waktu); err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

// ---------- dashboard aggregates ----------

func DashboardRingkasan(db *sql.DB, satkerScope []int64) (*models.DashboardRingkasan, error) {
	var r models.DashboardRingkasan
	query := `
		SELECT
			COUNT(DISTINCT lp.id),
			COUNT(DISTINCT lp.id) FILTER (WHERE lp.status_kasus = 'Selesai'),
			COUNT(DISTINCT lp.id) FILTER (WHERE lp.status_kasus != 'Selesai'),
			COUNT(DISTINCT k.id) FILTER (WHERE k.status_kendaraan IN ('Ditemukan','Diamankan','Dikembalikan')),
			COUNT(DISTINCT k.id) FILTER (WHERE k.status_kendaraan = 'Terlapor Hilang')
		FROM laporan_polisi lp
		LEFT JOIN kendaraan k ON k.lp_id = lp.id
	`
	var err error
	if satkerScope != nil {
		query += ` WHERE lp.satker_id = ANY($1)`
		err = db.QueryRow(query, pq.Array(satkerScope)).Scan(&r.TotalKasus, &r.KasusSelesai, &r.KasusBelumTerungkap, &r.KendaraanDitemukan, &r.KendaraanBelumDitemukan)
	} else {
		err = db.QueryRow(query).Scan(&r.TotalKasus, &r.KasusSelesai, &r.KasusBelumTerungkap, &r.KendaraanDitemukan, &r.KendaraanBelumDitemukan)
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func DashboardTrenBulanan(db *sql.DB, satkerScope []int64, bulanTerakhir int) ([]models.TrenBulanan, error) {
	query := `
		SELECT to_char(date_trunc('month', tanggal_lp), 'YYYY-MM') AS bulan, COUNT(*) AS jumlah
		FROM laporan_polisi
		WHERE tanggal_lp >= (CURRENT_DATE - ($1 || ' months')::interval)
	`
	args := []interface{}{bulanTerakhir}
	if satkerScope != nil {
		query += ` AND satker_id = ANY($2)`
		args = append(args, pq.Array(satkerScope))
	}
	query += ` GROUP BY 1 ORDER BY 1`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.TrenBulanan
	for rows.Next() {
		var t models.TrenBulanan
		if err := rows.Scan(&t.Bulan, &t.JumlahKasus); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func DashboardKategoriKasus(db *sql.DB, satkerScope []int64) ([]models.KategoriKasus, error) {
	query := `SELECT jenis_perkara, COUNT(*) FROM laporan_polisi`
	args := []interface{}{}
	if satkerScope != nil {
		query += ` WHERE satker_id = ANY($1)`
		args = append(args, pq.Array(satkerScope))
	}
	query += ` GROUP BY jenis_perkara ORDER BY 2 DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.KategoriKasus
	for rows.Next() {
		var k models.KategoriKasus
		if err := rows.Scan(&k.JenisPerkara, &k.Jumlah); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// PetaSebaran mengembalikan titik lokasi TKP untuk kebutuhan peta sebaran
// (belum terungkap / sudah terungkap), dengan filter rentang tanggal
// opsional (mis. "curanmor di Batam dalam seminggu terakhir").
func PetaSebaran(db *sql.DB, satkerScope []int64, dari, sampai string) ([]models.PetaSebaranPoint, error) {
	query := `
		SELECT lp.id, lp.no_lp, lp.satker_id, sk.nama_satker, lp.tanggal_kejadian, COALESCE(lp.tkp_alamat,''),
			lp.tkp_latitude, lp.tkp_longitude, lp.status_kasus,
			CASE WHEN lp.status_kasus = 'Selesai' THEN true ELSE false END
		FROM laporan_polisi lp
		JOIN satuan_kerja sk ON sk.id = lp.satker_id
		WHERE lp.tkp_latitude IS NOT NULL AND lp.tkp_longitude IS NOT NULL
	`
	args := []interface{}{}
	argN := 1
	if satkerScope != nil {
		query += ` AND lp.satker_id = ANY($` + strconv.Itoa(argN) + `)`
		args = append(args, pq.Array(satkerScope))
		argN++
	}
	if dari != "" {
		query += ` AND lp.tanggal_kejadian >= $` + strconv.Itoa(argN)
		args = append(args, dari)
		argN++
	}
	if sampai != "" {
		query += ` AND lp.tanggal_kejadian <= $` + strconv.Itoa(argN)
		args = append(args, sampai)
		argN++
	}
	query += ` ORDER BY lp.tanggal_kejadian DESC LIMIT 2000`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.PetaSebaranPoint
	for rows.Next() {
		var p models.PetaSebaranPoint
		var tgl sql.NullString
		if err := rows.Scan(&p.LPID, &p.NoLP, &p.SatkerID, &p.NamaSatker, &tgl, &p.TKPAlamat, &p.Latitude, &p.Longitude, &p.StatusKasus, &p.SudahTerungkap); err != nil {
			return nil, err
		}
		if tgl.Valid {
			p.TanggalKejadian = &tgl.String
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
