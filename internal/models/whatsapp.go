package models

import "time"

type LogInteraksiWhatsapp struct {
	ID            int64     `json:"id"`
	NoWhatsapp    string    `json:"no_whatsapp"`
	PenggunaID    *int64    `json:"pengguna_id,omitempty"`
	JenisInput    string    `json:"jenis_input"`
	IsiPesanMasuk string    `json:"isi_pesan_masuk,omitempty"`
	StatusAkses   string    `json:"status_akses"` // diizinkan | ditolak
	WaktuMasuk    time.Time `json:"waktu_masuk"`
}

type HasilOcrAI struct {
	ID                 int64                  `json:"id"`
	LogID              int64                  `json:"log_id"`
	JenisProses        string                 `json:"jenis_proses"`
	TeksMentah         string                 `json:"teks_mentah,omitempty"`
	SkorKeyakinan      *float64               `json:"skor_keyakinan,omitempty"`
	HasilEkstraksiJSON map[string]interface{} `json:"hasil_ekstraksi_json,omitempty"`
}

type LogPencarian struct {
	ID             int64  `json:"id"`
	LogID          int64  `json:"log_id"`
	TipePencarian  string `json:"tipe_pencarian"`
	KataKunci      string `json:"kata_kunci"`
	Ditemukan      bool   `json:"ditemukan"`
	KendaraanID    *int64 `json:"kendaraan_id,omitempty"`
	WaktuResponsMs *int   `json:"waktu_respons_ms,omitempty"`
}

type ResponsAI struct {
	ID               int64     `json:"id"`
	LogID            int64     `json:"log_id"`
	IsiRespons       string    `json:"isi_respons"`
	WaktuKirim       time.Time `json:"waktu_kirim"`
	StatusPengiriman string    `json:"status_pengiriman"`
}
