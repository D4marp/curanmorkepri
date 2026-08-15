package repo

import (
	"database/sql"

	"curanmor-ai/internal/models"
)

// ---------- pihak_terkait (NIK disimpan sebagai BYTEA terenkripsi; enkripsi/
// dekripsi dilakukan di handler menggunakan internal/cryptox, repo hanya
// meneruskan []byte apa adanya agar lapisan data tidak memegang kunci) ----------

func ListPihakByLP(db *sql.DB, lpID int64) ([]models.PihakTerkait, error) {
	rows, err := db.Query(`SELECT id, lp_id, jenis_pihak, nama, nik_enc, COALESCE(alamat,''), COALESCE(no_telp,'') FROM pihak_terkait WHERE lp_id = $1 ORDER BY id`, lpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.PihakTerkait
	for rows.Next() {
		var p models.PihakTerkait
		var nikEnc []byte
		if err := rows.Scan(&p.ID, &p.LPID, &p.JenisPihak, &p.Nama, &nikEnc, &p.Alamat, &p.NoTelp); err != nil {
			return nil, err
		}
		p.NIK = "" // tidak pernah dikembalikan mentah dari list; gunakan NIKMasked (diisi handler)
		out = append(out, p)
	}
	return out, rows.Err()
}

func GetPihakNikEnc(db *sql.DB, id int64) ([]byte, error) {
	var nikEnc []byte
	err := db.QueryRow(`SELECT nik_enc FROM pihak_terkait WHERE id = $1`, id).Scan(&nikEnc)
	return nikEnc, err
}

// GetPihakByID mengambil pihak_terkait termasuk lp_id (tanpa nik_enc) —
// dipakai handler untuk resolusi otorisasi cakupan satker sebelum
// update/delete lewat endpoint per-ID.
func GetPihakByID(db *sql.DB, id int64) (*models.PihakTerkait, error) {
	var p models.PihakTerkait
	err := db.QueryRow(`SELECT id, lp_id, jenis_pihak, nama, COALESCE(alamat,''), COALESCE(no_telp,'') FROM pihak_terkait WHERE id = $1`, id).
		Scan(&p.ID, &p.LPID, &p.JenisPihak, &p.Nama, &p.Alamat, &p.NoTelp)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func CreatePihak(db *sql.DB, lpID int64, jenisPihak, nama string, nikEnc []byte, alamat, noTelp string) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO pihak_terkait (lp_id, jenis_pihak, nama, nik_enc, alamat, no_telp)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id
	`, lpID, jenisPihak, nama, nikEnc, alamat, noTelp).Scan(&id)
	return id, err
}

func UpdatePihak(db *sql.DB, id int64, jenisPihak, nama string, nikEnc []byte, alamat, noTelp string) error {
	_, err := db.Exec(`UPDATE pihak_terkait SET jenis_pihak=$1, nama=$2, nik_enc=$3, alamat=$4, no_telp=$5 WHERE id=$6`,
		jenisPihak, nama, nikEnc, alamat, noTelp, id)
	return err
}

func DeletePihak(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM pihak_terkait WHERE id = $1`, id)
	return err
}

// ---------- barang_bukti ----------

func ListBarangBuktiByLP(db *sql.DB, lpID int64) ([]models.BarangBukti, error) {
	rows, err := db.Query(`SELECT id, lp_id, COALESCE(jenis_bb,''), COALESCE(no_registrasi_bb,''), COALESCE(deskripsi,''), COALESCE(lokasi_penyimpanan,''), COALESCE(foto_bb_url,''), tanggal_diamankan FROM barang_bukti WHERE lp_id = $1 ORDER BY id`, lpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.BarangBukti
	for rows.Next() {
		var b models.BarangBukti
		var tgl sql.NullString
		if err := rows.Scan(&b.ID, &b.LPID, &b.JenisBB, &b.NoRegistrasiBB, &b.Deskripsi, &b.LokasiPenyimpanan, &b.FotoBBURL, &tgl); err != nil {
			return nil, err
		}
		if tgl.Valid {
			b.TanggalDiamankan = &tgl.String
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func CreateBarangBukti(db *sql.DB, b *models.BarangBukti) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO barang_bukti (lp_id, jenis_bb, no_registrasi_bb, deskripsi, lokasi_penyimpanan, foto_bb_url, tanggal_diamankan)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id
	`, b.LPID, b.JenisBB, b.NoRegistrasiBB, b.Deskripsi, b.LokasiPenyimpanan, b.FotoBBURL, b.TanggalDiamankan).Scan(&id)
	return id, err
}

// GetBarangBuktiByID mengambil satu baris barang_bukti termasuk lp_id —
// dipakai handler untuk resolusi otorisasi cakupan satker sebelum hapus
// lewat endpoint per-ID.
func GetBarangBuktiByID(db *sql.DB, id int64) (*models.BarangBukti, error) {
	var b models.BarangBukti
	var tgl sql.NullString
	err := db.QueryRow(`SELECT id, lp_id, COALESCE(jenis_bb,''), COALESCE(no_registrasi_bb,''), COALESCE(deskripsi,''), COALESCE(lokasi_penyimpanan,''), COALESCE(foto_bb_url,''), tanggal_diamankan FROM barang_bukti WHERE id = $1`, id).
		Scan(&b.ID, &b.LPID, &b.JenisBB, &b.NoRegistrasiBB, &b.Deskripsi, &b.LokasiPenyimpanan, &b.FotoBBURL, &tgl)
	if err != nil {
		return nil, err
	}
	if tgl.Valid {
		b.TanggalDiamankan = &tgl.String
	}
	return &b, nil
}

func DeleteBarangBukti(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM barang_bukti WHERE id = $1`, id)
	return err
}

// ---------- dokumentasi_pendukung ----------

func ListDokumentasiByLP(db *sql.DB, lpID int64) ([]models.DokumentasiPendukung, error) {
	rows, err := db.Query(`SELECT id, lp_id, jenis_dokumen, file_url, diunggah_oleh, waktu_unggah FROM dokumentasi_pendukung WHERE lp_id = $1 ORDER BY waktu_unggah DESC`, lpID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.DokumentasiPendukung
	for rows.Next() {
		var d models.DokumentasiPendukung
		if err := rows.Scan(&d.ID, &d.LPID, &d.JenisDokumen, &d.FileURL, &d.DiunggahOleh, &d.WaktuUnggah); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func GetDokumentasiByID(db *sql.DB, id int64) (*models.DokumentasiPendukung, error) {
	var d models.DokumentasiPendukung
	err := db.QueryRow(`SELECT id, lp_id, jenis_dokumen, file_url, diunggah_oleh, waktu_unggah FROM dokumentasi_pendukung WHERE id = $1`, id).
		Scan(&d.ID, &d.LPID, &d.JenisDokumen, &d.FileURL, &d.DiunggahOleh, &d.WaktuUnggah)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func CreateDokumentasi(db *sql.DB, d *models.DokumentasiPendukung) (int64, error) {
	var id int64
	err := db.QueryRow(`
		INSERT INTO dokumentasi_pendukung (lp_id, jenis_dokumen, file_url, diunggah_oleh)
		VALUES ($1,$2,$3,$4) RETURNING id
	`, d.LPID, d.JenisDokumen, d.FileURL, d.DiunggahOleh).Scan(&id)
	return id, err
}

func DeleteDokumentasi(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM dokumentasi_pendukung WHERE id = $1`, id)
	return err
}
