package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

// WaHandler mengekspos endpoint yang dipanggil oleh SISTEM CHATBOT WHATSAPP
// AI (eksternal, di luar cakupan proyek ini — lihat README) untuk membaca
// & mencatat data lewat API, TANPA pernah menyentuh database secara
// langsung, sesuai arahan: "anggota di lapangan aksesnya lewat API yang
// terkoneksi ke knowledge base AI... bagian kita cuma buat web input &
// pencarian data lewat dashboard". Grup ini diotentikasi dengan API key
// (X-API-Key), bukan sesi JWT pengguna.
type WaHandler struct{ DB *sql.DB }

type whitelistCheckRequest struct {
	NoWhatsapp string `json:"no_whatsapp"`
}

// POST /api/v1/wa/whitelist/check
// Dipanggil chatbot SEBELUM memproses pesan apa pun dari suatu nomor.
func (h *WaHandler) WhitelistCheck(w http.ResponseWriter, r *http.Request) {
	var req whitelistCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	result, err := repo.CheckWhitelist(h.DB, req.NoWhatsapp)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, result)
}

type logInteraksiRequest struct {
	NoWhatsapp    string `json:"no_whatsapp"`
	JenisInput    string `json:"jenis_input"`
	IsiPesanMasuk string `json:"isi_pesan_masuk"`
}

// POST /api/v1/wa/log-interaksi
// Mencatat pesan masuk (teks/foto STNK/foto rangka/foto mesin/voice note).
// Otomatis mengisi status_akses berdasarkan hasil cek whitelist saat ini.
func (h *WaHandler) LogInteraksi(w http.ResponseWriter, r *http.Request) {
	var req logInteraksiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.NoWhatsapp == "" || req.JenisInput == "" {
		httpx.BadRequest(w, "no_whatsapp dan jenis_input wajib diisi")
		return
	}
	wl, err := repo.CheckWhitelist(h.DB, req.NoWhatsapp)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	statusAkses := "ditolak"
	var penggunaID *int64
	if wl.Diizinkan {
		statusAkses = "diizinkan"
		penggunaID = wl.PenggunaID
	}
	l := models.LogInteraksiWhatsapp{
		NoWhatsapp: req.NoWhatsapp, PenggunaID: penggunaID, JenisInput: req.JenisInput,
		IsiPesanMasuk: req.IsiPesanMasuk, StatusAkses: statusAkses,
	}
	id, err := repo.CreateLogInteraksi(h.DB, &l)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	created, err := repo.GetLogInteraksiByID(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

type hasilOcrRequest struct {
	LogID              int64                  `json:"log_id"`
	JenisProses        string                 `json:"jenis_proses"`
	TeksMentah         string                 `json:"teks_mentah"`
	SkorKeyakinan      *float64               `json:"skor_keyakinan"`
	HasilEkstraksiJSON map[string]interface{} `json:"hasil_ekstraksi_json"`
}

// POST /api/v1/wa/hasil-ocr
// Mencatat output OCR/speech-to-text/visual matching dari AI eksternal.
func (h *WaHandler) HasilOcr(w http.ResponseWriter, r *http.Request) {
	var req hasilOcrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.LogID == 0 || req.JenisProses == "" {
		httpx.BadRequest(w, "log_id dan jenis_proses wajib diisi")
		return
	}
	hasil := models.HasilOcrAI{
		LogID: req.LogID, JenisProses: req.JenisProses, TeksMentah: req.TeksMentah,
		SkorKeyakinan: req.SkorKeyakinan, HasilEkstraksiJSON: req.HasilEkstraksiJSON,
	}
	id, err := repo.CreateHasilOcr(h.DB, &hasil)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	hasil.ID = id
	httpx.Created(w, hasil)
}

type waSearchRequest struct {
	NoWhatsapp    string `json:"no_whatsapp"`
	LogID         *int64 `json:"log_id"`         // opsional; bila kosong, endpoint ini membuat log_interaksi sendiri
	TipePencarian string `json:"tipe_pencarian"` // "No. Polisi" | "No. Rangka" | "No. Mesin"
	KataKunci     string `json:"kata_kunci"`
}

// POST /api/v1/wa/search — JALUR UTAMA pencarian CURANMOR AI dari WhatsApp.
// Merespons cepat (indeks pada kolom no_polisi/no_rangka_vin/no_mesin,
// lihat migrations/0005) sesuai target "hitungan detik" pada paparan.
// Hasil pencarian di-scope otomatis sesuai satuan kerja pemilik nomor WA
// (bukan akses global), konsisten dengan multi-tenant Admin Polda/Polres/Polsek.
func (h *WaHandler) Search(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req waSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.NoWhatsapp == "" || req.KataKunci == "" {
		httpx.BadRequest(w, "no_whatsapp dan kata_kunci wajib diisi")
		return
	}
	if req.TipePencarian == "" {
		req.TipePencarian = "No. Polisi"
	}

	wl, err := repo.CheckWhitelist(h.DB, req.NoWhatsapp)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if !wl.Diizinkan {
		httpx.Forbidden(w, "Nomor WhatsApp tidak terdaftar/tidak aktif: "+wl.Alasan)
		return
	}

	logID := req.LogID
	if logID == nil {
		l := models.LogInteraksiWhatsapp{
			NoWhatsapp: req.NoWhatsapp, PenggunaID: wl.PenggunaID, JenisInput: "Teks",
			IsiPesanMasuk: req.KataKunci, StatusAkses: "diizinkan",
		}
		newID, err := repo.CreateLogInteraksi(h.DB, &l)
		if err != nil {
			httpx.InternalError(w, err)
			return
		}
		logID = &newID
	}

	var scope []int64
	if wl.PenggunaID != nil {
		p, err := repo.GetPenggunaByID(h.DB, *wl.PenggunaID)
		if err != nil {
			httpx.InternalError(w, err)
			return
		}
		scope, err = repo.ResolveSatkerScope(h.DB, p.SatkerID)
		if err != nil {
			httpx.InternalError(w, err)
			return
		}
	}

	results, err := repo.SearchKendaraan(h.DB, req.TipePencarian, req.KataKunci, scope)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}

	elapsedMs := int(time.Since(start).Milliseconds())
	ditemukan := len(results) > 0
	var kendaraanID *int64
	if ditemukan {
		kendaraanID = &results[0].ID
	}
	logPencarian := models.LogPencarian{
		LogID: *logID, TipePencarian: req.TipePencarian, KataKunci: req.KataKunci,
		Ditemukan: ditemukan, KendaraanID: kendaraanID, WaktuResponsMs: &elapsedMs,
	}
	if _, err := repo.CreateLogPencarian(h.DB, &logPencarian); err != nil {
		httpx.InternalError(w, err)
		return
	}

	httpx.OK(w, map[string]interface{}{
		"log_id": *logID, "ditemukan": ditemukan, "waktu_respons_ms": elapsedMs, "hasil": results,
	})
}

type responsAiRequest struct {
	LogID            int64  `json:"log_id"`
	IsiRespons       string `json:"isi_respons"`
	StatusPengiriman string `json:"status_pengiriman"`
}

// POST /api/v1/wa/respons
// Mencatat balasan yang dikirim AI ke WhatsApp petugas (jejak audit interaksi).
func (h *WaHandler) CreateRespons(w http.ResponseWriter, r *http.Request) {
	var req responsAiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.LogID == 0 || req.IsiRespons == "" {
		httpx.BadRequest(w, "log_id dan isi_respons wajib diisi")
		return
	}
	if req.StatusPengiriman == "" {
		req.StatusPengiriman = "terkirim"
	}
	resp := models.ResponsAI{LogID: req.LogID, IsiRespons: req.IsiRespons, StatusPengiriman: req.StatusPengiriman}
	id, err := repo.CreateResponsAI(h.DB, &resp)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	created, err := repo.GetResponsAIByID(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

// GET /api/v1/wa/log-interaksi?no_whatsapp=&satker_id=&page=&page_size=
// Dipanggil dari sisi WA-service untuk audit internal mereka. Tanpa
// "satker_id" mengembalikan interaksi lintas semua satker (kredensial
// bridge WA bersifat global, bukan per-tenant) — beri "satker_id" untuk
// membatasi hasil ke satker tersebut + jajarannya.
func (h *WaHandler) ListLogInteraksi(w http.ResponseWriter, r *http.Request) {
	f := repo.LogInteraksiFilter{
		NoWhatsapp: r.URL.Query().Get("no_whatsapp"),
		Page:       queryInt(r, "page", 1),
		PageSize:   queryInt(r, "page_size", 50),
	}
	if satkerID := int64(queryInt(r, "satker_id", 0)); satkerID != 0 {
		scope, err := repo.ResolveSatkerScope(h.DB, satkerID)
		if err != nil {
			httpx.InternalError(w, err)
			return
		}
		if scope == nil {
			scope = []int64{satkerID} // Polda diminta eksplisit -> tetap tampilkan hanya dirinya, bukan "semua"
		}
		f.SatkerScope = scope
	}
	list, total, err := repo.ListLogInteraksi(h.DB, f)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OKMeta(w, list, map[string]interface{}{"total": total, "page": f.Page, "page_size": f.PageSize})
}
