package models

import "time"

type NotifikasiAlert struct {
	ID               int64     `json:"id"`
	JenisAlert       string    `json:"jenis_alert"`
	Deskripsi        string    `json:"deskripsi"`
	TargetPeranID    *int64    `json:"target_peran_id,omitempty"`
	TargetPenggunaID *int64    `json:"target_pengguna_id,omitempty"`
	StatusBaca       bool      `json:"status_baca"`
	CreatedAt        time.Time `json:"created_at"`
}

type LaporanPeriodik struct {
	ID             int64     `json:"id"`
	JenisLaporan   string    `json:"jenis_laporan"` // Harian | Mingguan | Bulanan
	PeriodeMulai   string    `json:"periode_mulai"`
	PeriodeSelesai string    `json:"periode_selesai"`
	SatkerID       *int64    `json:"satker_id,omitempty"`
	FileURL        string    `json:"file_url,omitempty"`
	DibuatOtomatis bool      `json:"dibuat_otomatis"`
	CreatedAt      time.Time `json:"created_at"`
}

type AuditLog struct {
	ID           int64     `json:"id"`
	PenggunaID   *int64    `json:"pengguna_id,omitempty"`
	PenggunaNama string    `json:"pengguna_nama,omitempty"`
	Aktivitas    string    `json:"aktivitas"`
	Modul        string    `json:"modul,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	Perangkat    string    `json:"perangkat,omitempty"`
	Waktu        time.Time `json:"waktu"`
}

type DashboardRingkasan struct {
	TotalKasus              int `json:"total_kasus"`
	KasusSelesai            int `json:"kasus_selesai"`
	KasusBelumTerungkap     int `json:"kasus_belum_terungkap"`
	KendaraanDitemukan      int `json:"kendaraan_ditemukan"`
	KendaraanBelumDitemukan int `json:"kendaraan_belum_ditemukan"`
}

type TrenBulanan struct {
	Bulan       string `json:"bulan"`
	JumlahKasus int    `json:"jumlah_kasus"`
}

type KategoriKasus struct {
	JenisPerkara string `json:"jenis_perkara"`
	Jumlah       int    `json:"jumlah"`
}

type PetaSebaranPoint struct {
	LPID            int64   `json:"lp_id"`
	NoLP            string  `json:"no_lp"`
	SatkerID        int64   `json:"satker_id"`
	NamaSatker      string  `json:"nama_satker"`
	TanggalKejadian *string `json:"tanggal_kejadian,omitempty"`
	TKPAlamat       string  `json:"tkp_alamat,omitempty"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	StatusKasus     string  `json:"status_kasus"`
	SudahTerungkap  bool    `json:"sudah_terungkap"`
}
