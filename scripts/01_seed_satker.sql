-- =====================================================================
-- Seed: Data Induk Satuan Kerja (Polda Kepri + 7 Polres/Polresta + Polsek jajaran)
-- Sumber: dokumentasi database (Polda+Polres) + daftar Polsek jajaran Polda Kepri
-- =====================================================================

INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLDA-KEPRI', 'Polda Kepulauan Riau', 'Polda', NULL, 'Kepulauan Riau')
ON CONFLICT (kode_satker) DO NOTHING;

INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLRESTA-BRL', 'Polresta Barelang', 'Polres', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLDA-KEPRI'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;

INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLRES-TPI', 'Polresta Tanjungpinang', 'Polres', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLDA-KEPRI'), 'Tanjungpinang')
ON CONFLICT (kode_satker) DO NOTHING;

INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLRES-KRM', 'Polres Karimun', 'Polres', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLDA-KEPRI'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;

INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLRES-BTN', 'Polres Bintan', 'Polres', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLDA-KEPRI'), 'Bintan')
ON CONFLICT (kode_satker) DO NOTHING;

INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLRES-LGA', 'Polres Lingga', 'Polres', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLDA-KEPRI'), 'Lingga')
ON CONFLICT (kode_satker) DO NOTHING;

INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLRES-NTN', 'Polres Natuna', 'Polres', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLDA-KEPRI'), 'Natuna')
ON CONFLICT (kode_satker) DO NOTHING;

INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLRES-ANB', 'Polres Kepulauan Anambas', 'Polres', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLDA-KEPRI'), 'Kepulauan Anambas')
ON CONFLICT (kode_satker) DO NOTHING;

-- Polsek jajaran Polresta Barelang
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-01', 'Polsek Batam Kota', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-02', 'Polsek Lubuk Baja', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-03', 'Polsek Batu Ampar', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-04', 'Polsek Bengkong', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-05', 'Polsek Sei Beduk', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-06', 'Polsek Nongsa', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-07', 'Polsek Sekupang', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-08', 'Polsek Sagulung', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-09', 'Polsek Batu Aji', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-10', 'Polsek Galang', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-11', 'Polsek Bulang', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-12', 'Polsek Belakang Padang', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-13', 'Polsek Kawasan Pelabuhan Batam (KKP)', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BRL-14', 'Polsek Kawasan Bandara Hang Nadim Batam', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRESTA-BRL'), 'Batam')
ON CONFLICT (kode_satker) DO NOTHING;

-- Polsek jajaran Polresta Tanjungpinang
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-TPI-01', 'Polsek Tanjungpinang Kota', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-TPI'), 'Tanjungpinang')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-TPI-02', 'Polsek Tanjungpinang Timur', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-TPI'), 'Tanjungpinang')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-TPI-03', 'Polsek Bukit Bestari', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-TPI'), 'Tanjungpinang')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-TPI-04', 'Polsek Tanjungpinang Barat', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-TPI'), 'Tanjungpinang')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-TPI-05', 'Polsek Kawasan Bandara Raja Haji Fisabilillah (RHF)', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-TPI'), 'Tanjungpinang')
ON CONFLICT (kode_satker) DO NOTHING;

-- Polsek jajaran Polres Karimun
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-KRM-01', 'Polsek Balai Karimun', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-KRM'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-KRM-02', 'Polsek Tebing', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-KRM'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-KRM-03', 'Polsek Meral', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-KRM'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-KRM-04', 'Polsek Kundur', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-KRM'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-KRM-05', 'Polsek Moro', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-KRM'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-KRM-06', 'Polsek Kuba (Kundur Utara)', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-KRM'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-KRM-07', 'Polsek Buru', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-KRM'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-KRM-08', 'Polsek KKP (Kawasan Kepolisian Pelabuhan) Karimun', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-KRM'), 'Karimun')
ON CONFLICT (kode_satker) DO NOTHING;

-- Polsek jajaran Polres Bintan
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BTN-01', 'Polsek Bintan Utara', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-BTN'), 'Bintan')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BTN-02', 'Polsek Bintan Timur', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-BTN'), 'Bintan')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BTN-03', 'Polsek Gunung Kijang', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-BTN'), 'Bintan')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BTN-04', 'Polsek Teluk Bintan', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-BTN'), 'Bintan')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-BTN-05', 'Polsek Tambelan', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-BTN'), 'Bintan')
ON CONFLICT (kode_satker) DO NOTHING;

-- Polsek jajaran Polres Lingga
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-LGA-01', 'Polsek Dabo Singkep', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-LGA'), 'Lingga')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-LGA-02', 'Polsek Singkep Barat', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-LGA'), 'Lingga')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-LGA-03', 'Polsek Daik Lingga', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-LGA'), 'Lingga')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-LGA-04', 'Polsek Senayang', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-LGA'), 'Lingga')
ON CONFLICT (kode_satker) DO NOTHING;

-- Polsek jajaran Polres Natuna
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-NTN-01', 'Polsek Bunguran Timur', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-NTN'), 'Natuna')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-NTN-02', 'Polsek Bunguran Barat', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-NTN'), 'Natuna')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-NTN-03', 'Polsek Serasan', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-NTN'), 'Natuna')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-NTN-04', 'Polsek Midai', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-NTN'), 'Natuna')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-NTN-05', 'Polsek Pulau Laut', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-NTN'), 'Natuna')
ON CONFLICT (kode_satker) DO NOTHING;

-- Polsek jajaran Polres Kepulauan Anambas
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-ANB-01', 'Polsek Siantan', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-ANB'), 'Kepulauan Anambas')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-ANB-02', 'Polsek Palmatak', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-ANB'), 'Kepulauan Anambas')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-ANB-03', 'Polsek Jemaja', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-ANB'), 'Kepulauan Anambas')
ON CONFLICT (kode_satker) DO NOTHING;
INSERT INTO satuan_kerja (kode_satker, nama_satker, jenis_satker, induk_id, wilayah) VALUES
    ('POLSEK-ANB-04', 'Polsek Jemaja Timur', 'Polsek', (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLRES-ANB'), 'Kepulauan Anambas')
ON CONFLICT (kode_satker) DO NOTHING;

