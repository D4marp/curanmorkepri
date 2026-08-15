package middleware

import (
	"database/sql"
	"net/http"

	"curanmor-ai/internal/httpx"
)

// RequireModule memeriksa matriks peran_modul_akses: 'ditolak' -> 403 untuk
// semua method; 'terbatas' -> 403 untuk method yang mengubah data (hanya
// baca yang diizinkan); 'penuh' -> semua method diizinkan.
// Harus dipasang SETELAH AuthRequired (butuh Claims di context).
func RequireModule(database *sql.DB, kodeModul string) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := httpx.ClaimsFromContext(r.Context())
			if !ok {
				httpx.Unauthorized(w, "Sesi tidak ditemukan")
				return
			}
			var level string
			err := database.QueryRow(`
				SELECT pma.level_akses
				FROM peran_modul_akses pma
				JOIN modul_akses m ON m.id = pma.modul_id
				WHERE pma.peran_id = $1 AND m.kode_modul = $2
			`, claims.PeranID, kodeModul).Scan(&level)
			if err == sql.ErrNoRows {
				httpx.Forbidden(w, "Anda tidak memiliki akses ke modul ini")
				return
			}
			if err != nil {
				httpx.InternalError(w, err)
				return
			}
			if level == "ditolak" {
				httpx.Forbidden(w, "Peran Anda tidak memiliki akses ke modul ini")
				return
			}
			isMutating := r.Method == http.MethodPost || r.Method == http.MethodPut ||
				r.Method == http.MethodPatch || r.Method == http.MethodDelete
			if level == "terbatas" && isMutating {
				httpx.Forbidden(w, "Peran Anda hanya memiliki akses baca (terbatas) pada modul ini")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
