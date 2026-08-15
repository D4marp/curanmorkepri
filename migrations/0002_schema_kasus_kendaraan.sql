-- =====================================================================
-- Migration 0002: Domain B - Kasus & Kendaraan
-- Laporan Polisi sebagai single source of truth
-- =====================================================================

DO $$ BEGIN
    CREATE TYPE jenis_perkara_enum AS ENUM ('Pencurian Ranmor', 'Penggelapan', 'Penadahan', 'Lainnya');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE status_kasus_enum AS ENUM ('LP', 'SP2HP', 'DPO', 'Selesai');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE jenis_pihak_enum AS ENUM ('Pelapor', 'Terlapor', 'Saksi');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE jenis_kendaraan_enum AS ENUM ('Roda 2', 'Roda 4', 'Lainnya');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE status_kendaraan_enum AS ENUM ('Terlapor Hilang', 'Ditemukan', 'Diamankan', 'Dikembalikan');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE jenis_dokumen_enum AS ENUM ('STNK', 'BPKB', 'Foto TKP', 'Foto Rangka', 'Foto Mesin', 'Lainnya');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ---------- laporan_polisi ----------
CREATE TABLE IF NOT EXISTS laporan_polisi (
    id              BIGSERIAL PRIMARY KEY,
    no_lp           VARCHAR(30) NOT NULL UNIQUE, -- format LP/xxxx/xx/xxxx/PKT
    tanggal_lp      DATE NOT NULL,
    jenis_perkara   jenis_perkara_enum NOT NULL DEFAULT 'Pencurian Ranmor',
    satker_id       BIGINT NOT NULL REFERENCES satuan_kerja(id),
    penyidik_id     BIGINT REFERENCES pengguna(id),
    tanggal_kejadian DATE,
    tkp_alamat      TEXT,
    tkp_latitude    DECIMAL(9,6),
    tkp_longitude   DECIMAL(9,6),
    status_kasus    status_kasus_enum NOT NULL DEFAULT 'LP',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE laporan_polisi IS 'Single source of truth kasus (LP)';

-- ---------- pihak_terkait ----------
CREATE TABLE IF NOT EXISTS pihak_terkait (
    id          BIGSERIAL PRIMARY KEY,
    lp_id       BIGINT NOT NULL REFERENCES laporan_polisi(id) ON DELETE CASCADE,
    jenis_pihak jenis_pihak_enum NOT NULL,
    nama        VARCHAR(150) NOT NULL,
    nik_enc     BYTEA, -- NIK dienkripsi AES-256-GCM di level aplikasi (lihat internal/cryptox)
    alamat      TEXT,
    no_telp     VARCHAR(20)
);
COMMENT ON TABLE pihak_terkait IS 'Pelapor, terlapor, saksi';
COMMENT ON COLUMN pihak_terkait.nik_enc IS 'NIK terenkripsi at-rest (AES-256-GCM, kunci dari APP_ENCRYPTION_KEY)';

-- ---------- kendaraan ----------
CREATE TABLE IF NOT EXISTS kendaraan (
    id                  BIGSERIAL PRIMARY KEY,
    lp_id               BIGINT NOT NULL REFERENCES laporan_polisi(id) ON DELETE CASCADE,
    no_polisi           VARCHAR(15),
    no_rangka_vin       VARCHAR(30),
    no_mesin            VARCHAR(30),
    merk_tipe           VARCHAR(50),
    warna               VARCHAR(30),
    tahun               SMALLINT CHECK (tahun IS NULL OR (tahun BETWEEN 1900 AND 2100)),
    jenis_kendaraan     jenis_kendaraan_enum NOT NULL DEFAULT 'Roda 2',
    status_kendaraan    status_kendaraan_enum NOT NULL DEFAULT 'Terlapor Hilang',
    foto_kendaraan_url  TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE kendaraan IS 'Objek pencarian utama CURANMOR AI';

-- ---------- riwayat_status_kasus ----------
CREATE TABLE IF NOT EXISTS riwayat_status_kasus (
    id                  BIGSERIAL PRIMARY KEY,
    lp_id               BIGINT NOT NULL REFERENCES laporan_polisi(id) ON DELETE CASCADE,
    status_sebelumnya   VARCHAR(30),
    status_baru         VARCHAR(30) NOT NULL,
    tanggal_perubahan   TIMESTAMPTZ NOT NULL DEFAULT now(),
    keterangan          TEXT,
    diubah_oleh         BIGINT REFERENCES pengguna(id)
);
COMMENT ON TABLE riwayat_status_kasus IS 'Jejak perubahan status LP -> SP2HP -> DPO -> Selesai';

-- ---------- barang_bukti ----------
CREATE TABLE IF NOT EXISTS barang_bukti (
    id                  BIGSERIAL PRIMARY KEY,
    lp_id               BIGINT NOT NULL REFERENCES laporan_polisi(id) ON DELETE CASCADE,
    jenis_bb            VARCHAR(100),
    no_registrasi_bb    VARCHAR(30) UNIQUE,
    deskripsi           TEXT,
    lokasi_penyimpanan  VARCHAR(150),
    foto_bb_url         TEXT,
    tanggal_diamankan   DATE
);
COMMENT ON TABLE barang_bukti IS 'Barang bukti terkait LP';

-- ---------- dokumentasi_pendukung ----------
CREATE TABLE IF NOT EXISTS dokumentasi_pendukung (
    id              BIGSERIAL PRIMARY KEY,
    lp_id           BIGINT NOT NULL REFERENCES laporan_polisi(id) ON DELETE CASCADE,
    jenis_dokumen   jenis_dokumen_enum NOT NULL,
    file_url        TEXT NOT NULL,
    diunggah_oleh   BIGINT REFERENCES pengguna(id),
    waktu_unggah    TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE dokumentasi_pendukung IS 'Berkas pendukung LP';

-- ---------- riwayat_penyidikan ----------
CREATE TABLE IF NOT EXISTS riwayat_penyidikan (
    id          BIGSERIAL PRIMARY KEY,
    lp_id       BIGINT NOT NULL REFERENCES laporan_polisi(id) ON DELETE CASCADE,
    tanggal     DATE NOT NULL,
    kegiatan    VARCHAR(200),
    penyidik_id BIGINT REFERENCES pengguna(id),
    catatan     TEXT
);
COMMENT ON TABLE riwayat_penyidikan IS 'Jurnal kegiatan penyidik';
