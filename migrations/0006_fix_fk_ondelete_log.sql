-- Banyak kolom FK ke pengguna(id)/kendaraan(id) di seluruh skema adalah
-- tautan atribusi historis yang nullable (siapa membuat/mengunggah/mengubah/
-- menyidik/mencari), tapi tanpa ON DELETE eksplisit Postgres memakai NO
-- ACTION -> menghapus satu pengguna atau laporan (yang cascade ke
-- kendaraan miliknya) gagal dengan foreign key violation begitu ada
-- riwayat yang merujuknya. Data historis tetap berharga meski entitas yang
-- dirujuk sudah dihapus, jadi SET NULL alih-alih memblokir penghapusan.
-- (whitelist_whatsapp.pengguna_id sengaja TIDAK diubah — itu relasi
-- kepemilikan, bukan atribusi, dan memang sudah ON DELETE CASCADE.)

ALTER TABLE log_interaksi_whatsapp
    DROP CONSTRAINT log_interaksi_whatsapp_pengguna_id_fkey,
    ADD CONSTRAINT log_interaksi_whatsapp_pengguna_id_fkey
        FOREIGN KEY (pengguna_id) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE log_pencarian
    DROP CONSTRAINT log_pencarian_kendaraan_id_fkey,
    ADD CONSTRAINT log_pencarian_kendaraan_id_fkey
        FOREIGN KEY (kendaraan_id) REFERENCES kendaraan(id) ON DELETE SET NULL;

ALTER TABLE pengguna
    DROP CONSTRAINT pengguna_dibuat_oleh_fkey,
    ADD CONSTRAINT pengguna_dibuat_oleh_fkey
        FOREIGN KEY (dibuat_oleh) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE whitelist_whatsapp
    DROP CONSTRAINT whitelist_whatsapp_diregistrasi_oleh_fkey,
    ADD CONSTRAINT whitelist_whatsapp_diregistrasi_oleh_fkey
        FOREIGN KEY (diregistrasi_oleh) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE wa_api_key
    DROP CONSTRAINT wa_api_key_dibuat_oleh_fkey,
    ADD CONSTRAINT wa_api_key_dibuat_oleh_fkey
        FOREIGN KEY (dibuat_oleh) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE laporan_polisi
    DROP CONSTRAINT laporan_polisi_penyidik_id_fkey,
    ADD CONSTRAINT laporan_polisi_penyidik_id_fkey
        FOREIGN KEY (penyidik_id) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE riwayat_status_kasus
    DROP CONSTRAINT riwayat_status_kasus_diubah_oleh_fkey,
    ADD CONSTRAINT riwayat_status_kasus_diubah_oleh_fkey
        FOREIGN KEY (diubah_oleh) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE dokumentasi_pendukung
    DROP CONSTRAINT dokumentasi_pendukung_diunggah_oleh_fkey,
    ADD CONSTRAINT dokumentasi_pendukung_diunggah_oleh_fkey
        FOREIGN KEY (diunggah_oleh) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE riwayat_penyidikan
    DROP CONSTRAINT riwayat_penyidikan_penyidik_id_fkey,
    ADD CONSTRAINT riwayat_penyidikan_penyidik_id_fkey
        FOREIGN KEY (penyidik_id) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE audit_log
    DROP CONSTRAINT audit_log_pengguna_id_fkey,
    ADD CONSTRAINT audit_log_pengguna_id_fkey
        FOREIGN KEY (pengguna_id) REFERENCES pengguna(id) ON DELETE SET NULL;

ALTER TABLE notifikasi_alert
    DROP CONSTRAINT notifikasi_alert_target_pengguna_id_fkey,
    ADD CONSTRAINT notifikasi_alert_target_pengguna_id_fkey
        FOREIGN KEY (target_pengguna_id) REFERENCES pengguna(id) ON DELETE SET NULL;
