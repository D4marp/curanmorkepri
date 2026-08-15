package middleware

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"curanmor-ai/internal/httpx"
)

// AuditTrail mencatat setiap request yang mengubah data (POST/PUT/PATCH/DELETE)
// dari pengguna yang sudah terotentikasi ke tabel audit_log. Dipasang setelah
// AuthRequired agar Claims tersedia di context. Kegagalan menulis audit log
// TIDAK menggagalkan request (dicatat ke log server saja) supaya modul audit
// tidak menjadi single point of failure bagi operasional sistem.
func AuditTrail(database *sql.DB) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			isMutating := r.Method == http.MethodPost || r.Method == http.MethodPut ||
				r.Method == http.MethodPatch || r.Method == http.MethodDelete
			if !isMutating {
				return
			}
			claims, ok := httpx.ClaimsFromContext(r.Context())
			if !ok {
				return
			}
			modul := inferModuleFromPath(r.URL.Path)
			aktivitas := truncate(r.Method+" "+r.URL.Path, 200)
			_, err := database.Exec(
				`INSERT INTO audit_log (pengguna_id, aktivitas, modul, ip_address, perangkat) VALUES ($1, $2, $3, $4, $5)`,
				claims.Subject, aktivitas, modul, httpx.ClientIP(r), truncate(r.UserAgent(), 100),
			)
			if err != nil {
				log.Printf("[audit] gagal menulis audit_log: %v", err)
			}
		})
	}
}

func inferModuleFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// contoh: api/v1/kendaraan/{id} -> "kendaraan"
	for i, p := range parts {
		if p == "v1" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// RecordAuditManual dipakai untuk mencatat aktivitas yang perlu deskripsi
// spesifik (mis. login sukses/gagal, sebelum Claims tersedia di context).
func RecordAuditManual(database *sql.DB, penggunaID *int64, aktivitas, modul, ip, perangkat string) {
	_, err := database.Exec(
		`INSERT INTO audit_log (pengguna_id, aktivitas, modul, ip_address, perangkat) VALUES ($1, $2, $3, $4, $5)`,
		penggunaID, truncate(aktivitas, 200), modul, ip, truncate(perangkat, 100),
	)
	if err != nil {
		log.Printf("[audit] gagal menulis audit_log (manual): %v", err)
	}
}

// truncate memotong s ke maksimal n rune agar muat di kolom VARCHAR(n) —
// User-Agent browser modern rutin melebihi 100 karakter.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
