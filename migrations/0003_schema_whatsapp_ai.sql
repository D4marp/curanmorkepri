-- =====================================================================
-- Migration 0003: Domain C - Interaksi WhatsApp AI
-- Dicatat oleh layanan chatbot AI eksternal lewat /api/v1/wa/*
-- =====================================================================

DO $$ BEGIN
    CREATE TYPE jenis_input_enum AS ENUM ('Teks', 'Foto STNK', 'Foto No. Mesin', 'Foto No. Rangka', 'Voice Note', 'Foto Kendaraan');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE status_akses_enum AS ENUM ('diizinkan', 'ditolak');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE jenis_proses_enum AS ENUM ('OCR Teks', 'OCR Gambar', 'Speech-to-Text', 'Visual Matching');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE tipe_pencarian_enum AS ENUM ('No. Polisi', 'No. Rangka', 'No. Mesin');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE status_pengiriman_enum AS ENUM ('terkirim', 'gagal');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ---------- log_interaksi_whatsapp ----------
CREATE TABLE IF NOT EXISTS log_interaksi_whatsapp (
    id              BIGSERIAL PRIMARY KEY,
    no_whatsapp     VARCHAR(20) NOT NULL,
    pengguna_id     BIGINT REFERENCES pengguna(id),
    jenis_input     jenis_input_enum NOT NULL,
    isi_pesan_masuk TEXT,
    status_akses    status_akses_enum NOT NULL,
    waktu_masuk     TIMESTAMPTZ NOT NULL DEFAULT now()
);
COMMENT ON TABLE log_interaksi_whatsapp IS 'Pesan masuk dari petugas via WhatsApp Business API';

-- ---------- hasil_ocr_ai ----------
CREATE TABLE IF NOT EXISTS hasil_ocr_ai (
    id                  BIGSERIAL PRIMARY KEY,
    log_id              BIGINT NOT NULL REFERENCES log_interaksi_whatsapp(id) ON DELETE CASCADE,
    jenis_proses        jenis_proses_enum NOT NULL,
    teks_mentah         TEXT,
    skor_keyakinan      DECIMAL(4,3) CHECK (skor_keyakinan IS NULL OR (skor_keyakinan BETWEEN 0 AND 1)),
    hasil_ekstraksi_json JSONB
);
COMMENT ON TABLE hasil_ocr_ai IS 'Output OCR / speech-to-text / visual AI';

-- ---------- log_pencarian ----------
CREATE TABLE IF NOT EXISTS log_pencarian (
    id              BIGSERIAL PRIMARY KEY,
    log_id          BIGINT NOT NULL REFERENCES log_interaksi_whatsapp(id) ON DELETE CASCADE,
    tipe_pencarian  tipe_pencarian_enum NOT NULL,
    kata_kunci      VARCHAR(50) NOT NULL,
    ditemukan       BOOLEAN NOT NULL DEFAULT false,
    kendaraan_id    BIGINT REFERENCES kendaraan(id),
    waktu_respons_ms INTEGER
);
COMMENT ON TABLE log_pencarian IS 'Query pencarian kendaraan & hasilnya';

-- ---------- respons_ai ----------
CREATE TABLE IF NOT EXISTS respons_ai (
    id              BIGSERIAL PRIMARY KEY,
    log_id          BIGINT NOT NULL REFERENCES log_interaksi_whatsapp(id) ON DELETE CASCADE,
    isi_respons     TEXT NOT NULL,
    waktu_kirim     TIMESTAMPTZ NOT NULL DEFAULT now(),
    status_pengiriman status_pengiriman_enum NOT NULL DEFAULT 'terkirim'
);
COMMENT ON TABLE respons_ai IS 'Balasan yang dikirim ke WhatsApp petugas';
