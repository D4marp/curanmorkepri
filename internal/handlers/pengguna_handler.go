package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"curanmor-ai/internal/auth"
	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/models"
	"curanmor-ai/internal/repo"
)

type PenggunaHandler struct{ DB *sql.DB }

func inScope(scope []int64, satkerID int64) bool {
	if scope == nil {
		return true
	}
	for _, s := range scope {
		if s == satkerID {
			return true
		}
	}
	return false
}

// authorizeSatkerScope memastikan satkerID berada dalam cakupan satuan kerja
// pengguna yang login saat ini (Admin Polda: semua; Admin Polres: satker
// sendiri + jajaran Polsek; Admin Polsek: satker sendiri saja). Dipakai di
// semua handler yang mengakses/mengubah/menghapus data milik satker tertentu
// lewat endpoint per-ID, agar tidak bisa diakses lintas-tenant hanya dengan
// menebak ID (celah otorisasi kritis untuk sistem data kasus kepolisian ini).
func authorizeSatkerScope(db *sql.DB, r *http.Request, satkerID int64) (bool, error) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	scope, err := repo.ResolveSatkerScope(db, claims.SatkerID)
	if err != nil {
		return false, err
	}
	return inScope(scope, satkerID), nil
}

// GET /api/v1/pengguna
func (h *PenggunaHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	list, err := repo.ListPengguna(h.DB, scope)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, list)
}

// GET /api/v1/pengguna/{id}
func (h *PenggunaHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := repo.GetPenggunaByID(h.DB, pathID(r))
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Pengguna tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeSatkerScope(h.DB, r, p.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengakses pengguna di satuan kerja tersebut")
		return
	}
	httpx.OK(w, p)
}

type createPenggunaRequest struct {
	NRP         string  `json:"nrp"`
	NamaLengkap string  `json:"nama_lengkap"`
	Pangkat     string  `json:"pangkat"`
	Jabatan     string  `json:"jabatan"`
	PeranID     int64   `json:"peran_id"`
	SatkerID    int64   `json:"satker_id"`
	NoWhatsapp  *string `json:"no_whatsapp"`
	Password    string  `json:"password"`
}

// POST /api/v1/pengguna  (modul kelola_pengguna)
func (h *PenggunaHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := httpx.ClaimsFromContext(r.Context())
	var req createPenggunaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if req.NRP == "" || req.NamaLengkap == "" || req.PeranID == 0 || req.SatkerID == 0 {
		httpx.BadRequest(w, "nrp, nama_lengkap, peran_id, satker_id wajib diisi")
		return
	}
	if len(req.Password) < 10 {
		httpx.BadRequest(w, "Password minimal 10 karakter")
		return
	}
	scope, err := repo.ResolveSatkerScope(h.DB, claims.SatkerID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if !inScope(scope, req.SatkerID) {
		httpx.Forbidden(w, "Anda tidak berwenang membuat pengguna di satuan kerja tersebut")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	creatorID := parseID(claims.Subject)
	p := models.Pengguna{
		NRP: req.NRP, NamaLengkap: req.NamaLengkap, Pangkat: req.Pangkat, Jabatan: req.Jabatan,
		PeranID: req.PeranID, SatkerID: req.SatkerID, NoWhatsapp: req.NoWhatsapp,
		PasswordHash: hash, DibuatOleh: &creatorID, StatusAktif: true,
	}
	id, err := repo.CreatePengguna(h.DB, &p)
	if err != nil {
		httpx.Conflict(w, "Gagal membuat pengguna (NRP/no. WhatsApp mungkin sudah dipakai): "+err.Error())
		return
	}
	// GetPenggunaByID tidak pernah memilih kolom password_hash, jadi hash
	// tidak akan pernah bocor lewat response ini (lihat juga tag json:"-").
	created, err := repo.GetPenggunaByID(h.DB, id)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.Created(w, created)
}

type updatePenggunaRequest struct {
	NamaLengkap string  `json:"nama_lengkap"`
	Pangkat     string  `json:"pangkat"`
	Jabatan     string  `json:"jabatan"`
	PeranID     int64   `json:"peran_id"`
	SatkerID    int64   `json:"satker_id"`
	NoWhatsapp  *string `json:"no_whatsapp"`
	StatusAktif bool    `json:"status_aktif"`
}

// PUT /api/v1/pengguna/{id}
func (h *PenggunaHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req updatePenggunaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	id := pathID(r)
	existing, err := repo.GetPenggunaByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Pengguna tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	// Cakupan satker pengguna yang diubah HARUS dicek terhadap satker asal
	// (existing.SatkerID) — tanpa ini, admin lintas-tenant dapat
	// mengubah/mengambil-alih akun pengguna di satuan kerja lain hanya
	// dengan menebak ID, sekalipun req.SatkerID kebetulan berada dalam
	// cakupannya sendiri.
	if ok, err := authorizeSatkerScope(h.DB, r, existing.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mengubah pengguna di satuan kerja tersebut")
		return
	}
	if ok, err := authorizeSatkerScope(h.DB, r, req.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang memindahkan pengguna ke satuan kerja tersebut")
		return
	}
	p := models.Pengguna{
		ID: id, NamaLengkap: req.NamaLengkap, Pangkat: req.Pangkat, Jabatan: req.Jabatan,
		PeranID: req.PeranID, SatkerID: req.SatkerID, NoWhatsapp: req.NoWhatsapp, StatusAktif: req.StatusAktif,
	}
	if err := repo.UpdatePengguna(h.DB, &p); err != nil {
		httpx.InternalError(w, err)
		return
	}
	updated, err := repo.GetPenggunaByID(h.DB, p.ID)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, updated)
}

// DELETE /api/v1/pengguna/{id}
func (h *PenggunaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	existing, err := repo.GetPenggunaByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Pengguna tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeSatkerScope(h.DB, r, existing.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang menghapus pengguna di satuan kerja tersebut")
		return
	}
	if err := repo.DeletePengguna(h.DB, id); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Pengguna dihapus"})
}

type resetPasswordRequest struct {
	PasswordBaru string `json:"password_baru"`
}

// POST /api/v1/pengguna/{id}/reset-password  (admin reset password anggota)
func (h *PenggunaHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Body request tidak valid")
		return
	}
	if len(req.PasswordBaru) < 10 {
		httpx.BadRequest(w, "Password baru minimal 10 karakter")
		return
	}
	id := pathID(r)
	existing, err := repo.GetPenggunaByID(h.DB, id)
	if err == sql.ErrNoRows {
		httpx.NotFound(w, "Pengguna tidak ditemukan")
		return
	}
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if ok, err := authorizeSatkerScope(h.DB, r, existing.SatkerID); err != nil {
		httpx.InternalError(w, err)
		return
	} else if !ok {
		httpx.Forbidden(w, "Anda tidak berwenang mereset password pengguna di satuan kerja tersebut")
		return
	}
	hash, err := auth.HashPassword(req.PasswordBaru)
	if err != nil {
		httpx.InternalError(w, err)
		return
	}
	if err := repo.UpdatePenggunaPassword(h.DB, id, hash); err != nil {
		httpx.InternalError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"message": "Password pengguna berhasil direset"})
}
