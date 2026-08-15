-- =====================================================================
-- Migration 0004: Domain D - Monitoring & Audit
-- =====================================================================

DO $$ BEGIN
    CREATE TYPE jenis_alert_enum AS ENUM ('Lonjakan Kasus', 'Aktivitas Mencurigakan', 'Kendaraan Ditemukan');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE jenis_laporan_enum AS ENUM ('Harian', 'Mingguan', 'Bulanan');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ---------- notifikasi_alert ----------
CREATE TABLE IF NOT EXISTS notifikasi_alert (
    id              BIGSERIAL PRIMARY KEY,
    jenis_alert     jenis_alert_enum NOT NULL,
    deskripsi       TEXT NOT NULL,
    target_peran_id BIGINT REFERENCES peran(id),
    target_pengguna_id BIGINT REFERENCES pengguna(id),
    status_baca     BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE notifikasi_alert IS 'Peringatan otomatis dashboard';

-- ---------- laporan_periodik ----------
CREATE TABLE IF NOT EXISTS laporan_periodik (
    id              BIGSERIAL PRIMARY KEY,
    jenis_laporan   jenis_laporan_enum NOT NULL,
    periode_mulai   DATE NOT NULL,
    periode_selesai DATE NOT NULL,
    satker_id       BIGINT REFERENCES satuan_kerja(id), -- kosong = seluruh Polda
    file_url        TEXT,
    dibuat_otomatis BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_periode CHECK (periode_selesai >= periode_mulai)
);
COMMENT ON TABLE laporan_periodik IS 'Laporan otomatis siap cetak';

-- ---------- audit_log ----------
CREATE TABLE IF NOT EXISTS audit_log (
    id              BIGSERIAL PRIMARY KEY,
    pengguna_id     BIGINT REFERENCES pengguna(id),
    aktivitas       VARCHAR(200) NOT NULL,
    modul           VARCHAR(50),
    ip_address      VARCHAR(45),
    perangkat       VARCHAR(100),
    waktu           TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE audit_log IS 'Jejak audit seluruh aktivitas pengguna';
