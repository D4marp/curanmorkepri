-- =====================================================================
-- CURANMOR AI — Ditreskrimum Polda Kepulauan Riau
-- Migration 0001: Domain A - Master & Akses
-- =====================================================================
-- Catatan desain: skema mengikuti dokumen "Susunan Database CURANMOR AI"
-- (23 Jul 2026). Satu penyesuaian ditambahkan di luar dokumen: kolom
-- induk_id + nilai 'Polsek' pada jenis_satker, untuk mendukung struktur
-- multi-tenant 3 tingkat (Polda > Polres > Polsek) yang diminta pada
-- diskusi WhatsApp (Admin Polda / Admin Polres / Admin Polsek).
-- =====================================================================

CREATE SCHEMA IF NOT EXISTS public;

-- ---------- ENUM TYPES ----------
DO $$ BEGIN
    CREATE TYPE jenis_satker_enum AS ENUM ('Polda', 'Polres', 'Polsek');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE level_akses_enum AS ENUM ('penuh', 'terbatas', 'ditolak');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE whitelist_status_enum AS ENUM ('aktif', 'nonaktif', 'diblokir');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ---------- satuan_kerja ----------
CREATE TABLE IF NOT EXISTS satuan_kerja (
    id              BIGSERIAL PRIMARY KEY,
    kode_satker     VARCHAR(20)  NOT NULL UNIQUE,
    nama_satker     VARCHAR(100) NOT NULL,
    jenis_satker    jenis_satker_enum NOT NULL,
    induk_id        BIGINT REFERENCES satuan_kerja(id) ON DELETE SET NULL, -- ekstensi hirarki (lihat catatan di atas)
    wilayah         VARCHAR(100),
    alamat          TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE satuan_kerja IS 'Polda & Polres/Polsek jajaran Kepulauan Riau';
COMMENT ON COLUMN satuan_kerja.induk_id IS 'Ekstensi desain: parent hierarchy Polda->Polres->Polsek untuk data-scoping multi-tenant';

-- ---------- peran ----------
CREATE TABLE IF NOT EXISTS peran (
    id          BIGSERIAL PRIMARY KEY,
    nama_peran  VARCHAR(50) NOT NULL UNIQUE,
    deskripsi   TEXT
);
COMMENT ON TABLE peran IS 'Daftar peran RBAC (6 peran sistem)';

-- ---------- modul_akses ----------
CREATE TABLE IF NOT EXISTS modul_akses (
    id          BIGSERIAL PRIMARY KEY,
    kode_modul  VARCHAR(50) NOT NULL UNIQUE,
    nama_modul  VARCHAR(100) NOT NULL
);
COMMENT ON TABLE modul_akses IS 'Modul fungsional sistem';

-- ---------- peran_modul_akses ----------
CREATE TABLE IF NOT EXISTS peran_modul_akses (
    id          BIGSERIAL PRIMARY KEY,
    peran_id    BIGINT NOT NULL REFERENCES peran(id) ON DELETE CASCADE,
    modul_id    BIGINT NOT NULL REFERENCES modul_akses(id) ON DELETE CASCADE,
    level_akses level_akses_enum NOT NULL,
    UNIQUE (peran_id, modul_id)
);
COMMENT ON TABLE peran_modul_akses IS 'Matriks hak akses (RBAC)';

-- ---------- pengguna ----------
CREATE TABLE IF NOT EXISTS pengguna (
    id              BIGSERIAL PRIMARY KEY,
    nrp             VARCHAR(20)  NOT NULL UNIQUE,
    nama_lengkap    VARCHAR(150) NOT NULL,
    pangkat         VARCHAR(50),
    jabatan         VARCHAR(100),
    peran_id        BIGINT NOT NULL REFERENCES peran(id),
    satker_id       BIGINT NOT NULL REFERENCES satuan_kerja(id),
    no_whatsapp     VARCHAR(20) UNIQUE,
    password_hash   TEXT NOT NULL, -- ekstensi wajib untuk login (bcrypt); tidak eksplisit di dokumen sumber
    status_aktif    BOOLEAN NOT NULL DEFAULT true,
    dibuat_oleh     BIGINT REFERENCES pengguna(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE pengguna IS 'Petugas & administrator sistem';
COMMENT ON COLUMN pengguna.password_hash IS 'Ekstensi wajib: hash bcrypt untuk otentikasi login dashboard';

-- ---------- whitelist_whatsapp ----------
CREATE TABLE IF NOT EXISTS whitelist_whatsapp (
    id                  BIGSERIAL PRIMARY KEY,
    no_whatsapp         VARCHAR(20) NOT NULL UNIQUE,
    pengguna_id         BIGINT NOT NULL REFERENCES pengguna(id) ON DELETE CASCADE,
    status              whitelist_status_enum NOT NULL DEFAULT 'aktif',
    tanggal_registrasi  TIMESTAMPTZ NOT NULL DEFAULT now(),
    diregistrasi_oleh   BIGINT REFERENCES pengguna(id),
    tanggal_verifikasi  TIMESTAMPTZ
);
COMMENT ON TABLE whitelist_whatsapp IS 'Keamanan akses chatbot WA berbasis nomor terdaftar';

-- ---------- wa_api_key (ekstensi: kredensial layanan untuk integrasi chatbot) ----------
CREATE TABLE IF NOT EXISTS wa_api_key (
    id              BIGSERIAL PRIMARY KEY,
    nama_layanan    VARCHAR(100) NOT NULL,
    key_hash        TEXT NOT NULL UNIQUE, -- SHA-256 hash dari API key (key asli hanya ditampilkan sekali saat dibuat)
    key_prefix      VARCHAR(12) NOT NULL, -- 8 karakter pertama, untuk identifikasi di UI/log tanpa membuka hash
    status_aktif    BOOLEAN NOT NULL DEFAULT true,
    dibuat_oleh     BIGINT REFERENCES pengguna(id),
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE wa_api_key IS 'Ekstensi: kredensial service-to-service untuk bot WhatsApp AI mengakses /api/v1/wa/*';
