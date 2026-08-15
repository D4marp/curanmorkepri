-- =====================================================================
-- Seed: peran (6 peran RBAC), modul_akses (6 modul), peran_modul_akses
-- Sumber: paparan slide 10 "Matriks Hak Akses (RBAC)" (dokumentasi hal. 14)
-- =====================================================================

INSERT INTO peran (nama_peran, deskripsi) VALUES
    ('Super Admin',        'Akses penuh seluruh data & pengaturan sistem. Digunakan oleh Admin Polda / Dirreskrimum / Wadirreskrimum.'),
    ('Admin Polres',       'Kelola data & pengguna di lingkup Polres/Polresta beserta Polsek jajarannya.'),
    ('Penyidik',           'Mengelola kasus, kendaraan, dan riwayat penyidikan yang menjadi tanggung jawabnya.'),
    ('Operator',           'Menginput & memperbarui data laporan/kendaraan di tingkat Polsek/Polres, akses laporan analitik terbatas.'),
    ('Viewer/Staf',        'Melihat data & laporan tanpa hak ubah (read-only), akses kelola data terbatas.'),
    ('Eksternal Terbatas', 'Akses terbatas untuk pihak eksternal/instansi terkait, hanya pencarian data & dashboard ringkas.')
ON CONFLICT (nama_peran) DO NOTHING;

INSERT INTO modul_akses (kode_modul, nama_modul) VALUES
    ('dashboard_monitoring', 'Dashboard Monitoring'),
    ('pencarian_data',       'Pencarian Data'),
    ('kelola_data',          'Kelola Data'),
    ('laporan_analitik',     'Laporan & Analitik'),
    ('kelola_pengguna',      'Kelola Pengguna'),
    ('pengaturan_sistem',    'Pengaturan Sistem')
ON CONFLICT (kode_modul) DO NOTHING;

-- Matriks hak akses (peran x modul -> level_akses)
WITH m AS (
    SELECT id, kode_modul FROM modul_akses
), p AS (
    SELECT id, nama_peran FROM peran
)
INSERT INTO peran_modul_akses (peran_id, modul_id, level_akses)
SELECT p.id, m.id, v.level::level_akses_enum
FROM (VALUES
    ('Super Admin',        'dashboard_monitoring', 'penuh'),
    ('Super Admin',        'pencarian_data',       'penuh'),
    ('Super Admin',        'kelola_data',          'penuh'),
    ('Super Admin',        'laporan_analitik',     'penuh'),
    ('Super Admin',        'kelola_pengguna',      'penuh'),
    ('Super Admin',        'pengaturan_sistem',    'penuh'),

    ('Admin Polres',       'dashboard_monitoring', 'penuh'),
    ('Admin Polres',       'pencarian_data',       'penuh'),
    ('Admin Polres',       'kelola_data',          'penuh'),
    ('Admin Polres',       'laporan_analitik',     'penuh'),
    ('Admin Polres',       'kelola_pengguna',      'ditolak'),
    ('Admin Polres',       'pengaturan_sistem',    'ditolak'),

    ('Penyidik',           'dashboard_monitoring', 'penuh'),
    ('Penyidik',           'pencarian_data',       'penuh'),
    ('Penyidik',           'kelola_data',          'penuh'),
    ('Penyidik',           'laporan_analitik',     'penuh'),
    ('Penyidik',           'kelola_pengguna',      'ditolak'),
    ('Penyidik',           'pengaturan_sistem',    'ditolak'),

    ('Operator',           'dashboard_monitoring', 'terbatas'),
    ('Operator',           'pencarian_data',       'penuh'),
    ('Operator',           'kelola_data',          'penuh'),
    ('Operator',           'laporan_analitik',     'terbatas'),
    ('Operator',           'kelola_pengguna',      'ditolak'),
    ('Operator',           'pengaturan_sistem',    'ditolak'),

    ('Viewer/Staf',        'dashboard_monitoring', 'penuh'),
    ('Viewer/Staf',        'pencarian_data',       'penuh'),
    ('Viewer/Staf',        'kelola_data',          'terbatas'),
    ('Viewer/Staf',        'laporan_analitik',     'penuh'),
    ('Viewer/Staf',        'kelola_pengguna',      'terbatas'),
    ('Viewer/Staf',        'pengaturan_sistem',    'terbatas'),

    ('Eksternal Terbatas', 'dashboard_monitoring', 'penuh'),
    ('Eksternal Terbatas', 'pencarian_data',       'penuh'),
    ('Eksternal Terbatas', 'kelola_data',          'terbatas'),
    ('Eksternal Terbatas', 'laporan_analitik',     'terbatas'),
    ('Eksternal Terbatas', 'kelola_pengguna',      'terbatas'),
    ('Eksternal Terbatas', 'pengaturan_sistem',    'terbatas')
) AS v(nama_peran, kode_modul, level)
JOIN p ON p.nama_peran = v.nama_peran
JOIN m ON m.kode_modul = v.kode_modul
ON CONFLICT (peran_id, modul_id) DO UPDATE SET level_akses = EXCLUDED.level_akses;

-- =====================================================================
-- Default Super Admin (WAJIB ganti password & no_whatsapp setelah deploy!)
-- NRP     : 00000000
-- Password: ChangeMe!12345   (bcrypt hash di bawah dibuat khusus untuk ini)
-- =====================================================================
INSERT INTO pengguna (nrp, nama_lengkap, pangkat, jabatan, peran_id, satker_id, no_whatsapp, password_hash, status_aktif)
SELECT
    '00000000',
    'Administrator Sistem',
    'AKBP',
    'Dirreskrimum',
    (SELECT id FROM peran WHERE nama_peran = 'Super Admin'),
    (SELECT id FROM satuan_kerja WHERE kode_satker = 'POLDA-KEPRI'),
    '0812xxxx1008',
    '$2a$10$cylYgtEfeIpfLGDXqPw5qOw5DhRpgy69ehKySYfqjnHnA99FQ8iAi', -- bcrypt('ChangeMe!12345') -- WAJIB DIGANTI setelah login pertama
    true
WHERE NOT EXISTS (SELECT 1 FROM pengguna WHERE nrp = '00000000');

INSERT INTO whitelist_whatsapp (no_whatsapp, pengguna_id, status)
SELECT '0812xxxx1008', (SELECT id FROM pengguna WHERE nrp = '00000000'), 'aktif'
WHERE NOT EXISTS (SELECT 1 FROM whitelist_whatsapp WHERE no_whatsapp = '0812xxxx1008');
