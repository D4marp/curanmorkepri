// Package models berisi struct data domain, dipetakan langsung dari skema
// pada migrations/*.sql (lihat dokumentasi database).
package models

import "time"

type SatuanKerja struct {
	ID          int64     `json:"id"`
	KodeSatker  string    `json:"kode_satker"`
	NamaSatker  string    `json:"nama_satker"`
	JenisSatker string    `json:"jenis_satker"` // Polda | Polres | Polsek
	IndukID     *int64    `json:"induk_id,omitempty"`
	Wilayah     string    `json:"wilayah,omitempty"`
	Alamat      string    `json:"alamat,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Peran struct {
	ID        int64  `json:"id"`
	NamaPeran string `json:"nama_peran"`
	Deskripsi string `json:"deskripsi,omitempty"`
}

type ModulAkses struct {
	ID        int64  `json:"id"`
	KodeModul string `json:"kode_modul"`
	NamaModul string `json:"nama_modul"`
}

type PeranModulAkses struct {
	ID         int64  `json:"id"`
	PeranID    int64  `json:"peran_id"`
	ModulID    int64  `json:"modul_id"`
	LevelAkses string `json:"level_akses"` // penuh | terbatas | ditolak
}

type Pengguna struct {
	ID          int64     `json:"id"`
	NRP         string    `json:"nrp"`
	NamaLengkap string    `json:"nama_lengkap"`
	Pangkat     string    `json:"pangkat,omitempty"`
	Jabatan     string    `json:"jabatan,omitempty"`
	PeranID     int64     `json:"peran_id"`
	PeranNama   string    `json:"peran_nama,omitempty"` // hasil JOIN, read-only
	SatkerID    int64     `json:"satker_id"`
	SatkerNama  string    `json:"satker_nama,omitempty"`  // hasil JOIN, read-only
	JenisSatker string    `json:"jenis_satker,omitempty"` // hasil JOIN, read-only
	NoWhatsapp  *string   `json:"no_whatsapp,omitempty"`
	StatusAktif bool      `json:"status_aktif"`
	DibuatOleh  *int64    `json:"dibuat_oleh,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// PasswordHash sengaja TIDAK di-JSON-kan (tag "-") agar tidak pernah
	// bocor lewat response API.
	PasswordHash string `json:"-"`
}

type WhitelistWhatsapp struct {
	ID                int64      `json:"id"`
	NoWhatsapp        string     `json:"no_whatsapp"`
	PenggunaID        int64      `json:"pengguna_id"`
	Status            string     `json:"status"` // aktif | nonaktif | diblokir
	TanggalRegistrasi time.Time  `json:"tanggal_registrasi"`
	DiregistrasiOleh  *int64     `json:"diregistrasi_oleh,omitempty"`
	TanggalVerifikasi *time.Time `json:"tanggal_verifikasi,omitempty"`
}

type WaAPIKey struct {
	ID          int64      `json:"id"`
	NamaLayanan string     `json:"nama_layanan"`
	KeyPrefix   string     `json:"key_prefix"`
	StatusAktif bool       `json:"status_aktif"`
	DibuatOleh  *int64     `json:"dibuat_oleh,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
