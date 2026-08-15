package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"curanmor-ai/internal/httpx"
)

// UploadHandler menangani unggah berkas (foto kendaraan, foto barang bukti,
// dokumen STNK/BPKB, dsb). Endpoint generik: klien unggah file dahulu lewat
// endpoint ini, mendapat URL, lalu menyertakan URL tsb pada payload JSON
// entitas terkait (kendaraan.foto_kendaraan_url, dsb).
type UploadHandler struct {
	UploadDir  string
	MaxSizeMB  int64
	PublicPath string // prefix URL publik, mis. "/uploads"
}

var allowedExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".pdf": true, ".ogg": true, ".mp3": true, ".m4a": true, ".opus": true,
}

// POST /api/v1/uploads  (multipart/form-data, field "file")
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxSizeMB*1024*1024)
	if err := r.ParseMultipartForm(h.MaxSizeMB * 1024 * 1024); err != nil {
		httpx.BadRequest(w, "Berkas terlalu besar atau form tidak valid (maks "+itoa64(h.MaxSizeMB)+"MB)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.BadRequest(w, "Field 'file' wajib diisi")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExt[ext] {
		httpx.BadRequest(w, "Tipe berkas tidak diizinkan. Diizinkan: jpg, png, webp, pdf, ogg, mp3, m4a, opus")
		return
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		httpx.InternalError(w, err)
		return
	}
	filename := hex.EncodeToString(randBytes) + ext
	destPath := filepath.Join(h.UploadDir, filename)

	if err := os.MkdirAll(h.UploadDir, 0o750); err != nil {
		httpx.InternalError(w, err)
		return
	}
	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		httpx.InternalError(w, err)
		return
	}

	httpx.Created(w, map[string]string{"url": h.PublicPath + "/" + filename})
}
