package models

import "time"

type LaporanPolisi struct {
	ID              int64     `json:"id"`
	NoLP            string    `json:"no_lp"`
	TanggalLP       string    `json:"tanggal_lp"` // DATE -> YYYY-MM-DD
	JenisPerkara    string    `json:"jenis_perkara"`
	SatkerID        int64     `json:"satker_id"`
	SatkerNama      string    `json:"satker_nama,omitempty"`
	PenyidikID      *int64    `json:"penyidik_id,omitempty"`
	PenyidikNama    string    `json:"penyidik_nama,omitempty"`
	TanggalKejadian *string   `json:"tanggal_kejadian,omitempty"`
	TKPAlamat       string    `json:"tkp_alamat,omitempty"`
	TKPLatitude     *float64  `json:"tkp_latitude,omitempty"`
	TKPLongitude    *float64  `json:"tkp_longitude,omitempty"`
	StatusKasus     string    `json:"status_kasus"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relasi opsional yang di-embed saat GET detail
	Kendaraan         []Kendaraan            `json:"kendaraan,omitempty"`
	PihakTerkait      []PihakTerkait         `json:"pihak_terkait,omitempty"`
	BarangBukti       []BarangBukti          `json:"barang_bukti,omitempty"`
	Dokumentasi       []DokumentasiPendukung `json:"dokumentasi,omitempty"`
	RiwayatStatus     []RiwayatStatusKasus   `json:"riwayat_status,omitempty"`
	RiwayatPenyidikan []RiwayatPenyidikan    `json:"riwayat_penyidikan,omitempty"`
}

type PihakTerkait struct {
	ID         int64  `json:"id"`
	LPID       int64  `json:"lp_id"`
	JenisPihak string `json:"jenis_pihak"` // Pelapor | Terlapor | Saksi
	Nama       string `json:"nama"`
	NIK        string `json:"nik,omitempty"`        // hanya diisi saat request (plaintext), tidak pernah dikembalikan utuh
	NIKMasked  string `json:"nik_masked,omitempty"` // ditampilkan di response, sudah disamarkan
	Alamat     string `json:"alamat,omitempty"`
	NoTelp     string `json:"no_telp,omitempty"`
}

type Kendaraan struct {
	ID               int64     `json:"id"`
	LPID             int64     `json:"lp_id"`
	NoPolisi         string    `json:"no_polisi,omitempty"`
	NoRangkaVIN      string    `json:"no_rangka_vin,omitempty"`
	NoMesin          string    `json:"no_mesin,omitempty"`
	MerkTipe         string    `json:"merk_tipe,omitempty"`
	Warna            string    `json:"warna,omitempty"`
	Tahun            *int      `json:"tahun,omitempty"`
	JenisKendaraan   string    `json:"jenis_kendaraan"`
	StatusKendaraan  string    `json:"status_kendaraan"`
	FotoKendaraanURL string    `json:"foto_kendaraan_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Konteks tambahan saat hasil pencarian (join ke laporan_polisi)
	NoLP        string `json:"no_lp,omitempty"`
	StatusKasus string `json:"status_kasus,omitempty"`
	SatkerNama  string `json:"satker_nama,omitempty"`
}

type RiwayatStatusKasus struct {
	ID               int64     `json:"id"`
	LPID             int64     `json:"lp_id"`
	StatusSebelumnya string    `json:"status_sebelumnya,omitempty"`
	StatusBaru       string    `json:"status_baru"`
	TanggalPerubahan time.Time `json:"tanggal_perubahan"`
	Keterangan       string    `json:"keterangan,omitempty"`
	DiubahOleh       *int64    `json:"diubah_oleh,omitempty"`
	DiubahOlehNama   string    `json:"diubah_oleh_nama,omitempty"`
}

type BarangBukti struct {
	ID                int64   `json:"id"`
	LPID              int64   `json:"lp_id"`
	JenisBB           string  `json:"jenis_bb,omitempty"`
	NoRegistrasiBB    string  `json:"no_registrasi_bb,omitempty"`
	Deskripsi         string  `json:"deskripsi,omitempty"`
	LokasiPenyimpanan string  `json:"lokasi_penyimpanan,omitempty"`
	FotoBBURL         string  `json:"foto_bb_url,omitempty"`
	TanggalDiamankan  *string `json:"tanggal_diamankan,omitempty"`
}

type DokumentasiPendukung struct {
	ID           int64     `json:"id"`
	LPID         int64     `json:"lp_id"`
	JenisDokumen string    `json:"jenis_dokumen"`
	FileURL      string    `json:"file_url"`
	DiunggahOleh *int64    `json:"diunggah_oleh,omitempty"`
	WaktuUnggah  time.Time `json:"waktu_unggah"`
}

type RiwayatPenyidikan struct {
	ID         int64  `json:"id"`
	LPID       int64  `json:"lp_id"`
	Tanggal    string `json:"tanggal"`
	Kegiatan   string `json:"kegiatan,omitempty"`
	PenyidikID *int64 `json:"penyidik_id,omitempty"`
	Catatan    string `json:"catatan,omitempty"`
}
