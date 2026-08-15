package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/cryptox"
	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

type PihakHandler struct {
	DB     *sql.DB
	Crypto *cryptox.Cipher
}

type pihakRequest struct {
	JenisPihak string `json:"jenis_pihak"`
	Nama       string `json:"nama"`
	NIK        string `json:"nik"`
	Alamat     string `json:"alamat"`
	NoTelp     string `json:"no_telp"`
}

// POST /api/v1/laporan/{id}/pihak-terkait
func (h *PihakHandler) Create(w http.ResponseWriter, r *http.Request) {
	lpID := pathID(r)
	if ok, err := authorizeLPScope(h.DB, r, lpID); err == sql.ErrNoRows {
		httpx.NotFound(w, "Laporan tidak ditemukan")
		return
	} else if err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menambah pihak terkait pada laporan satuan kerja tersebut")
		return
	}
	var req pihakRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.Nama == "" || req.JenisPihak == "" {
		httpx.BadRequest(w, "nama dan jenis_pihak wajib diisi")
		return
	}
	nikEnc, err := h.Crypto.Encrypt(req.NIK)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	id, err := repo.CreatePihak(h.DB, lpID, req.JenisPihak, req.Nama, nikEnc, req.Alamat, req.NoTelp)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	resp := models.PihakTerkait{ID: id, LPID: lpID, JenisPihak: req.JenisPihak, Nama: req.Nama, Alamat: req.Alamat, NoTelp: req.NoTelp}
	if req.NIK != "" {
		resp.NIKMasked = cryptox.MaskNIK(req.NIK)
	}
	httpx.Created(w, resp)
}

// PUT /api/v1/pihak-terkait/{id}
func (h *PihakHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetPihakByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Pihak terkait tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeLPScope(h.DB, r, existing.LPID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengubah pihak terkait satuan kerja tersebut")
		return
	}

	var req pihakRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	var nikEnc []byte
	if req.NIK != "" {
		nikEnc, err = h.Crypto.Encrypt(req.NIK)
		if err != nil {
			httpx.InternalError(w, err)
			return
		}
	} else {
		nikEnc, err = repo.GetPihakNikEnc(h.DB, id) // pertahankan NIK lama bila tidak diubah
		if err != nil && err != sql.ErrNoRows {
			httpx.InternalError(w, err)
			return
		}
	}
	if err := repo.UpdatePihak(h.DB, id, req.JenisPihak, req.Nama, nikEnc, req.Alamat, req.NoTelp); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Pihak terkait diperbarui"})
}

// DELETE /api/v1/pihak-terkait/{id}
func (h *PihakHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetPihakByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Pihak terkait tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeLPScope(h.DB, r, existing.LPID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menghapus pihak terkait satuan kerja tersebut")
		return
	}
	if err := repo.DeletePihak(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Pihak terkait dihapus"})
}
