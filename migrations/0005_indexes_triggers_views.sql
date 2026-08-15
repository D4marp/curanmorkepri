-- =====================================================================
-- Migration 0005: Indeks, trigger updated_at, view agregat dashboard
-- =====================================================================

-- ---------- Indeks pencarian utama (Catatan Desain: wajib diindeks) ----------
CREATE INDEX IF NOT EXISTS idx_kendaraan_no_polisi   ON kendaraan (no_polisi);
CREATE INDEX IF NOT EXISTS idx_kendaraan_no_rangka   ON kendaraan (no_rangka_vin);
CREATE INDEX IF NOT EXISTS idx_kendaraan_no_mesin    ON kendaraan (no_mesin);
CREATE INDEX IF NOT EXISTS idx_kendaraan_lp_id       ON kendaraan (lp_id);
CREATE INDEX IF NOT EXISTS idx_kendaraan_status      ON kendaraan (status_kendaraan);

CREATE INDEX IF NOT EXISTS idx_laporan_satker        ON laporan_polisi (satker_id);
CREATE INDEX IF NOT EXISTS idx_laporan_tanggal_lp     ON laporan_polisi (tanggal_lp);
CREATE INDEX IF NOT EXISTS idx_laporan_tanggal_kejadian ON laporan_polisi (tanggal_kejadian);
CREATE INDEX IF NOT EXISTS idx_laporan_status        ON laporan_polisi (status_kasus);
CREATE INDEX IF NOT EXISTS idx_laporan_geo           ON laporan_polisi (tkp_latitude, tkp_longitude);

CREATE INDEX IF NOT EXISTS idx_pihak_lp              ON pihak_terkait (lp_id);
CREATE INDEX IF NOT EXISTS idx_bb_lp                 ON barang_bukti (lp_id);
CREATE INDEX IF NOT EXISTS idx_dok_lp                ON dokumentasi_pendukung (lp_id);
CREATE INDEX IF NOT EXISTS idx_riwayat_status_lp     ON riwayat_status_kasus (lp_id);
CREATE INDEX IF NOT EXISTS idx_riwayat_penyidikan_lp ON riwayat_penyidikan (lp_id);

CREATE INDEX IF NOT EXISTS idx_satker_induk          ON satuan_kerja (induk_id);
CREATE INDEX IF NOT EXISTS idx_pengguna_satker       ON pengguna (satker_id);
CREATE INDEX IF NOT EXISTS idx_pengguna_peran        ON pengguna (peran_id);

CREATE INDEX IF NOT EXISTS idx_log_wa_no             ON log_interaksi_whatsapp (no_whatsapp);
CREATE INDEX IF NOT EXISTS idx_log_wa_waktu          ON log_interaksi_whatsapp (waktu_masuk);
CREATE INDEX IF NOT EXISTS idx_log_pencarian_log     ON log_pencarian (log_id);
CREATE INDEX IF NOT EXISTS idx_hasil_ocr_log         ON hasil_ocr_ai (log_id);
CREATE INDEX IF NOT EXISTS idx_respons_ai_log        ON respons_ai (log_id);

CREATE INDEX IF NOT EXISTS idx_audit_pengguna        ON audit_log (pengguna_id);
CREATE INDEX IF NOT EXISTS idx_audit_waktu           ON audit_log (waktu);
CREATE INDEX IF NOT EXISTS idx_notifikasi_target_user ON notifikasi_alert (target_pengguna_id, status_baca);

-- ---------- Trigger generik updated_at ----------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_pengguna_updated_at ON pengguna;
CREATE TRIGGER trg_pengguna_updated_at BEFORE UPDATE ON pengguna
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_laporan_updated_at ON laporan_polisi;
CREATE TRIGGER trg_laporan_updated_at BEFORE UPDATE ON laporan_polisi
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_kendaraan_updated_at ON kendaraan;
CREATE TRIGGER trg_kendaraan_updated_at BEFORE UPDATE ON kendaraan
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------- Trigger: catat otomatis riwayat_status_kasus saat status berubah ----------
CREATE OR REPLACE FUNCTION log_riwayat_status_kasus()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'UPDATE' AND NEW.status_kasus IS DISTINCT FROM OLD.status_kasus) THEN
        INSERT INTO riwayat_status_kasus (lp_id, status_sebelumnya, status_baru, keterangan)
        VALUES (NEW.id, OLD.status_kasus::text, NEW.status_kasus::text, 'Perubahan otomatis via update laporan_polisi');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_laporan_status_history ON laporan_polisi;
CREATE TRIGGER trg_laporan_status_history AFTER UPDATE ON laporan_polisi
    FOR EACH ROW EXECUTE FUNCTION log_riwayat_status_kasus();

-- ---------- Views agregat dashboard (Catatan Desain: view, bukan tabel fisik) ----------

-- Total kasus & kendaraan ditemukan per satker
CREATE OR REPLACE VIEW v_dashboard_ringkasan AS
SELECT
    lp.satker_id,
    COUNT(DISTINCT lp.id)                                              AS total_kasus,
    COUNT(DISTINCT lp.id) FILTER (WHERE lp.status_kasus = 'Selesai')   AS kasus_selesai,
    COUNT(DISTINCT lp.id) FILTER (WHERE lp.status_kasus != 'Selesai')  AS kasus_belum_terungkap,
    COUNT(DISTINCT k.id) FILTER (WHERE k.status_kendaraan IN ('Ditemukan','Diamankan','Dikembalikan')) AS kendaraan_ditemukan,
    COUNT(DISTINCT k.id) FILTER (WHERE k.status_kendaraan = 'Terlapor Hilang') AS kendaraan_belum_ditemukan
FROM laporan_polisi lp
LEFT JOIN kendaraan k ON k.lp_id = lp.id
GROUP BY lp.satker_id;

-- Grafik tren kasus per bulan (6 bulan terakhir dihitung di aplikasi)
CREATE OR REPLACE VIEW v_dashboard_tren_bulanan AS
SELECT
    satker_id,
    date_trunc('month', tanggal_lp)::date AS bulan,
    COUNT(*) AS jumlah_kasus
FROM laporan_polisi
GROUP BY satker_id, date_trunc('month', tanggal_lp);

-- Kategori kasus (jenis_perkara)
CREATE OR REPLACE VIEW v_dashboard_kategori_kasus AS
SELECT satker_id, jenis_perkara, COUNT(*) AS jumlah
FROM laporan_polisi
GROUP BY satker_id, jenis_perkara;

-- Sebaran lokasi TKP untuk peta (belum & sudah terungkap)
CREATE OR REPLACE VIEW v_peta_sebaran AS
SELECT
    lp.id AS lp_id,
    lp.no_lp,
    lp.satker_id,
    sk.nama_satker,
    lp.tanggal_kejadian,
    lp.tkp_alamat,
    lp.tkp_latitude,
    lp.tkp_longitude,
    lp.status_kasus,
    CASE WHEN lp.status_kasus = 'Selesai' THEN true ELSE false END AS sudah_terungkap
FROM laporan_polisi lp
JOIN satuan_kerja sk ON sk.id = lp.satker_id
WHERE lp.tkp_latitude IS NOT NULL AND lp.tkp_longitude IS NOT NULL;
