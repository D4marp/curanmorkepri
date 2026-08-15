package middleware

import (
	"log"
	"net/http"
	"time"

	"curanmor-ai/internal/httpx"
)

// CORS mengizinkan origin tertentu (dikonfigurasi via CORS_ALLOWED_ORIGINS).
func CORS(allowedOrigins string) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders menambahkan header keamanan standar (mitigasi clickjacking,
// MIME sniffing, dsb) — relevan karena sistem menyimpan data kasus kepolisian.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		// Seluruh aset front-end (Leaflet, Chart.js, Swagger UI) di-vendor lokal
		// di web/assets/vendor & docs/swagger-ui — tidak ada dependensi CDN
		// eksternal saat runtime, sehingga CSP bisa dibatasi ketat ke 'self'.
		// Pengecualian: ubin peta OpenStreetMap (img-src) untuk peta sebaran —
		// tidak bisa di-vendor karena jumlah ubin tak terbatas tergantung
		// wilayah/zoom yang dilihat pengguna.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https://*.tile.openstreetmap.org; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// Recover menangkap panic pada handler agar tidak mematikan server dan
// tidak membocorkan stack trace ke klien.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[panic] %s %s -> %v", r.Method, r.URL.Path, rec)
				httpx.Err(w, http.StatusInternalServerError, "internal_error", "Terjadi kesalahan pada server")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// AccessLog mencatat setiap request masuk (method, path, durasi, IP).
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		log.Printf("[http] %s %s -> %d (%s) ip=%s", r.Method, r.URL.Path, sw.status, time.Since(start), httpx.ClientIP(r))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}
