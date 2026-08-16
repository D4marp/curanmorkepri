package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"curanmor-ai/internal/cryptox"
	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

// BulkImportHandler menyediakan import massal laporan+kendaraan(+pelapor)
// lewat berkas Excel (.xlsx) — dipakai saat entri data lama/backlog perlu
// dipindahkan sekaligus, alih-alih satu-satu lewat formulir web. Template
// yang diunduh dari GET .../template HARUS dipakai sebagai dasar (kolom,
// urutan, & sheet "Petunjuk" berisi kode satker & nilai enum yang valid)
// supaya parsing di POST .../bulk-import bisa diandalkan.
type BulkImportHandler struct {
	DB     *sql.DB
	Crypto *cryptox.Cipher
	MaxMB  int64
}

const bulkSheetName = "Data"

var bulkHeaders = []string{
	"no_lp", "tanggal_lp", "jenis_perkara", "kode_satker", "tanggal_kejadian",
	"tkp_alamat", "tkp_latitude", "tkp_longitude", "status_kasus",
	"no_polisi", "no_rangka_vin", "no_mesin", "merk_tipe", "warna", "tahun",
	"jenis_kendaraan", "status_kendaraan",
	"pelapor_nama", "pelapor_nik", "pelapor_no_telp", "pelapor_alamat",
}

const bulkExampleMarker = "CONTOH_HAPUS_BARIS_INI"

// GET /api/v1/laporan/bulk-import/template — unduh template .xlsx kosong,
// lengkap dengan dropdown validasi enum & sheet referensi kode satuan kerja
// (diambil live dari database supaya selalu sinkron, bukan daftar statis
// yang bisa basi).
func (h *BulkImportHandler) DownloadTemplate(w http.ResponseWriter, r *http.Request) {
	satkerList, err := repo.ListSatuanKerja(h.DB)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	dataSheet := bulkSheetName
	f.SetSheetName("Sheet1", dataSheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1E1E1E"}, Pattern: 1},
	})
	for i, hd := range bulkHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(dataSheet, cell, hd)
		f.SetCellStyle(dataSheet, cell, cell, headerStyle)
	}
	f.SetColWidth(dataSheet, "A", "U", 20)
	f.SetRowHeight(dataSheet, 1, 22)

	example := []interface{}{
		bulkExampleMarker, "2026-08-15", "Pencurian Ranmor", satkerKodeOrPlaceholder(satkerList),
		"2026-08-15", "Jl. Contoh No. 1, Batam", 1.1305, 104.0553, "LP",
		"BP1234AB", "MH1JF1234567890123", "JF1234567890", "Honda Beat", "Merah", 2022,
		"Roda 2", "Terlapor Hilang",
		"Budi Santoso (contoh)", "3271081234567890", "081234567890", "Jl. Kenari No. 45, Batam",
	}
	rowNum := 2
	cellStart, _ := excelize.CoordinatesToCellName(1, rowNum)
	f.SetSheetRow(dataSheet, cellStart, &example)

	// Dropdown validasi supaya operator tidak salah ketik nilai enum.
	addListValidation(f, dataSheet, "C2:C1000", []string{"Pencurian Ranmor", "Penggelapan", "Penadahan", "Lainnya"})
	addListValidation(f, dataSheet, "I2:I1000", []string{"LP", "SP2HP", "DPO", "Selesai"})
	addListValidation(f, dataSheet, "P2:P1000", []string{"Roda 2", "Roda 4", "Lainnya"})
	addListValidation(f, dataSheet, "Q2:Q1000", []string{"Terlapor Hilang", "Ditemukan", "Diamankan", "Dikembalikan"})

	// Sheet referensi: kode satuan kerja (untuk kolom kode_satker) & petunjuk.
	refSheet := "Petunjuk"
	f.NewSheet(refSheet)
	f.SetColWidth(refSheet, "A", "A", 60)
	instr := []string{
		"PETUNJUK IMPORT MASSAL LAPORAN — CURANMOR AI",
		"",
		"1. Isi data pada sheet \"Data\", satu baris = satu laporan (+ 1 kendaraan opsional + 1 pelapor opsional).",
		"2. HAPUS baris contoh (baris 2, kolom no_lp berisi " + bulkExampleMarker + ") sebelum upload.",
		"3. Kolom wajib: no_lp, tanggal_lp, kode_satker. Sisanya opsional.",
		"4. Format tanggal: YYYY-MM-DD (contoh: 2026-08-15).",
		"5. kode_satker HARUS sama persis dengan daftar di bawah — lihat kolom \"Kode Satker\".",
		"6. Kolom kendaraan (no_polisi, dst) dikosongkan semua jika laporan belum ada data kendaraan.",
		"7. Kolom pelapor (pelapor_nama, dst) dikosongkan semua jika belum ada data pelapor.",
		"8. no_lp harus unik — baris dengan no_lp yang sudah ada di sistem akan GAGAL (dilaporkan di hasil upload).",
		"9. Upload berkas ini (setelah diisi) lewat menu Import Massal, field \"file\".",
		"",
		"=== DAFTAR KODE SATUAN KERJA (gunakan persis seperti ini) ===",
	}
	for i, line := range instr {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		f.SetCellValue(refSheet, cell, line)
	}
	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	f.SetCellStyle(refSheet, "A1", "A1", titleStyle)

	headerRow := len(instr) + 1
	hCell, _ := excelize.CoordinatesToCellName(1, headerRow)
	hCell2, _ := excelize.CoordinatesToCellName(2, headerRow)
	hCell3, _ := excelize.CoordinatesToCellName(3, headerRow)
	f.SetCellValue(refSheet, hCell, "Kode Satker")
	f.SetCellValue(refSheet, hCell2, "Nama Satuan Kerja")
	f.SetCellValue(refSheet, hCell3, "Jenis")
	f.SetCellStyle(refSheet, hCell, hCell3, headerStyle)
	for i, s := range satkerList {
		row := headerRow + 1 + i
		c1, _ := excelize.CoordinatesToCellName(1, row)
		c2, _ := excelize.CoordinatesToCellName(2, row)
		c3, _ := excelize.CoordinatesToCellName(3, row)
		f.SetCellValue(refSheet, c1, s.KodeSatker)
		f.SetCellValue(refSheet, c2, s.NamaSatker)
		f.SetCellValue(refSheet, c3, s.JenisSatker)
	}
	f.SetColWidth(refSheet, "B", "B", 40)

	f.SetActiveSheet(0)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="template-import-laporan-curanmor.xlsx"`)
	if err := f.Write(w); err != nil {
		log.Printf("[bulk-import] gagal menulis file xlsx ke response: %v", err)
	}
}

func addListValidation(f *excelize.File, sheet, cellRange string, options []string) {
	dv := excelize.NewDataValidation(true)
	dv.SetSqref(cellRange)
	_ = dv.SetDropList(options)
	_ = f.AddDataValidation(sheet, dv)
}

func satkerKodeOrPlaceholder(list []models.SatuanKerja) string {
	if len(list) == 0 {
		return "POLDA"
	}
	return list[0].KodeSatker
}

// BulkImportRowResult merangkum hasil satu baris untuk ditampilkan ke
// operator setelah upload — operator perlu tahu PERSIS baris mana yang
// gagal & kenapa, karena satu berkas bisa berisi puluhan/ratusan baris.
type BulkImportRowResult struct {
	Baris           int    `json:"baris"`
	NoLP            string `json:"no_lp"`
	Status          string `json:"status"` // "berhasil" | "gagal" | "dilewati"
	LPID            int64  `json:"lp_id,omitempty"`
	KendaraanDibuat bool   `json:"kendaraan_dibuat,omitempty"`
	PelaporDibuat   bool   `json:"pelapor_dibuat,omitempty"`
	Pesan           string `json:"pesan,omitempty"`
}

// POST /api/v1/laporan/bulk-import  (multipart/form-data, field "file")
func (h *BulkImportHandler) Import(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())

	maxBytes := h.MaxMB * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		httpx.BadRequest(w, "Berkas terlalu besar atau form tidak valid (maks "+itoa64(h.MaxMB)+"MB)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.BadRequest(w, "Field 'file' wajib diisi (unggah berkas .xlsx)")
		return
	}
	defer file.Close()
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		httpx.BadRequest(w, "Hanya berkas .xlsx yang didukung — unduh & pakai template resmi")
		return
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		httpx.BadRequest(w, "Gagal membaca berkas Excel: "+err.Error())
		return
	}
	defer f.Close()

	rows, err := f.GetRows(bulkSheetName)
	if err != nil {
		httpx.BadRequest(w, `Sheet "Data" tidak ditemukan — pakai template resmi, jangan ubah nama sheet`)
		return
	}
	if len(rows) < 2 {
		httpx.BadRequest(w, "Tidak ada baris data untuk diimpor")
		return
	}

	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	penyidikID := parseID(claims.Subject)

	var results []BulkImportRowResult
	created, failed, skipped := 0, 0, 0

	for i, row := range rows[1:] { // rows[0] = header
		rowNum := i + 2
		col := func(idx int) string {
			if idx >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[idx])
		}

		noLP := col(0)
		if noLP == "" || noLP == bulkExampleMarker {
			skipped++
			results = append(results, BulkImportRowResult{Baris: rowNum, NoLP: noLP, Status: "dilewati", Pesan: "baris kosong atau baris contoh"})
			continue
		}

		res := BulkImportRowResult{Baris: rowNum, NoLP: noLP}

		tanggalLP := col(1)
		jenisPerkara := col(2)
		kodeSatker := col(3)
		tanggalKejadian := col(4)
		tkpAlamat := col(5)
		tkpLat := col(6)
		tkpLng := col(7)
		statusKasus := col(8)
		noPolisi := col(9)
		noRangka := col(10)
		noMesin := col(11)
		merkTipe := col(12)
		warna := col(13)
		tahunStr := col(14)
		jenisKendaraan := col(15)
		statusKendaraan := col(16)
		pelaporNama := col(17)
		pelaporNIK := col(18)
		pelaporTelp := col(19)
		pelaporAlamat := col(20)

		if tanggalLP == "" || kodeSatker == "" {
			res.Status, res.Pesan = "gagal", "tanggal_lp dan kode_satker wajib diisi"
			failed++
			results = append(results, res)
			continue
		}
		if _, err := time.Parse("2006-01-02", tanggalLP); err != nil {
			res.Status, res.Pesan = "gagal", "tanggal_lp harus format YYYY-MM-DD"
			failed++
			results = append(results, res)
			continue
		}

		satker, err := repo.GetSatuanKerjaByKode(h.DB, kodeSatker)
		if err == sql.ErrNoRows {
			res.Status, res.Pesan = "gagal", "kode_satker \""+kodeSatker+"\" tidak ditemukan — lihat sheet Petunjuk"
			failed++
			results = append(results, res)
			continue
		} else if err != nil {
			res.Status, res.Pesan = "gagal", "gagal cek satuan kerja: "+err.Error()
			failed++
			results = append(results, res)
			continue
		}
		if !inScope(scope, satker.ID) {
			res.Status, res.Pesan = "gagal", "Anda tidak berwenang membuat laporan untuk satuan kerja \""+satker.NamaSatker+"\""
			failed++
			results = append(results, res)
			continue
		}

		if jenisPerkara == "" {
			jenisPerkara = "Pencurian Ranmor"
		}
		if statusKasus == "" {
			statusKasus = "LP"
		}

		lp := models.LaporanPolisi{
			NoLP: noLP, TanggalLP: tanggalLP, JenisPerkara: jenisPerkara,
			SatkerID: satker.ID, PenyidikID: &penyidikID,
			TKPAlamat: tkpAlamat, StatusKasus: statusKasus,
		}
		if tanggalKejadian != "" {
			lp.TanggalKejadian = &tanggalKejadian
		}
		if v, err := strconv.ParseFloat(tkpLat, 64); err == nil {
			lp.TKPLatitude = &v
		}
		if v, err := strconv.ParseFloat(tkpLng, 64); err == nil {
			lp.TKPLongitude = &v
		}

		lpID, err := repo.CreateLaporan(h.DB, &lp)
		if err != nil {
			res.Status, res.Pesan = "gagal", "gagal simpan laporan (no_lp mungkin sudah dipakai): "+err.Error()
			failed++
			results = append(results, res)
			continue
		}
		res.LPID = lpID

		hasKendaraan := noPolisi != "" || noRangka != "" || noMesin != ""
		if hasKendaraan {
			if jenisKendaraan == "" {
				jenisKendaraan = "Roda 2"
			}
			if statusKendaraan == "" {
				statusKendaraan = "Terlapor Hilang"
			}
			k := models.Kendaraan{
				LPID: lpID, NoPolisi: noPolisi, NoRangkaVIN: noRangka, NoMesin: noMesin,
				MerkTipe: merkTipe, Warna: warna, JenisKendaraan: jenisKendaraan, StatusKendaraan: statusKendaraan,
			}
			if tahun, err := strconv.Atoi(tahunStr); err == nil {
				k.Tahun = &tahun
			}
			if _, err := repo.CreateKendaraan(h.DB, &k); err != nil {
				res.Pesan += " | kendaraan gagal disimpan: " + err.Error()
			} else {
				res.KendaraanDibuat = true
			}
		}

		if pelaporNama != "" {
			nikEnc, err := h.Crypto.Encrypt(pelaporNIK)
			if err != nil {
				res.Pesan += " | pelapor gagal dienkripsi: " + err.Error()
			} else if _, err := repo.CreatePihak(h.DB, lpID, "Pelapor", pelaporNama, nikEnc, pelaporAlamat, pelaporTelp); err != nil {
				res.Pesan += " | pelapor gagal disimpan: " + err.Error()
			} else {
				res.PelaporDibuat = true
			}
		}

		res.Status = "berhasil"
		created++
		results = append(results, res)
	}

	httpx.OK(w, map[string]interface{}{
		"total_baris": len(rows) - 1,
		"berhasil":    created,
		"gagal":       failed,
		"dilewati":    skipped,
		"hasil":       results,
	})
}
