// Command api adalah entrypoint backend CURANMOR AI — Ditreskrimum Polda
// Kepulauan Riau. Menjalankan migrasi database, lalu menyajikan REST API
// (JSON, /api/v1/*), dashboard statis (HTML/CSS/JS di ./web), dokumentasi
// Swagger UI (/docs), dan berkas unggahan (/uploads).
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"curanmor-ai/internal/config"
	"curanmor-ai/internal/cryptox"
	"curanmor-ai/internal/db"
	"curanmor-ai/internal/handlers"
	"curanmor-ai/internal/httpx"
	"curanmor-ai/internal/middleware"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[fatal] %v", err)
	}
	defer database.Close()

	migrationsDir := envOrDefault("MIGRATIONS_DIR", "/app/migrations")
	seedDir := envOrDefault("SEED_DIR", "/app/scripts")
	if err := db.RunMigrations(database, migrationsDir); err != nil {
		log.Fatalf("[fatal] migrasi gagal: %v", err)
	}
	if err := db.RunSeedIfEmpty(database, seedDir); err != nil {
		log.Fatalf("[fatal] seeding gagal: %v", err)
	}

	crypto, err := cryptox.New(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("[fatal] gagal inisialisasi enkripsi: %v", err)
	}

	if err := os.MkdirAll(cfg.UploadDir, 0o750); err != nil {
		log.Fatalf("[fatal] gagal membuat direktori upload: %v", err)
	}

	r := httpx.NewRouter()

	// ---------- Handler instances ----------
	authH := &handlers.AuthHandler{DB: database, Cfg: cfg}
	satkerH := &handlers.SatkerHandler{DB: database}
	rbacH := &handlers.RBACHandler{DB: database}
	penggunaH := &handlers.PenggunaHandler{DB: database}
	whitelistH := &handlers.WhitelistHandler{DB: database}
	apiKeyH := &handlers.APIKeyHandler{DB: database}
	laporanH := &handlers.LaporanHandler{DB: database, Crypto: crypto}
	kendaraanH := &handlers.KendaraanHandler{DB: database}
	pihakH := &handlers.PihakHandler{DB: database, Crypto: crypto}
	evidenceH := &handlers.EvidenceHandler{DB: database}
	dashboardH := &handlers.DashboardHandler{DB: database}
	monitoringH := &handlers.MonitoringHandler{DB: database}
	waH := &handlers.WaHandler{DB: database}
	uploadH := &handlers.UploadHandler{UploadDir: cfg.UploadDir, MaxSizeMB: cfg.MaxUploadSizeMB, PublicPath: "/uploads"}

	loginLimiter := middleware.NewRateLimiter(10, 5*time.Minute)

	// ---------- Middleware umum (semua route) ----------
	authMw := middleware.AuthRequired(cfg.JWTSecret)
	auditMw := middleware.AuditTrail(database)

	// ---------- Publik (tanpa sesi) ----------
	pub := r.Group("")
	loginHandler := loginLimiter.Middleware()(http.HandlerFunc(authH.Login))
	pub.Handle("POST /api/v1/auth/login", loginHandler)

	// ---------- Sesi wajib, tanpa gate modul spesifik ----------
	authOnly := r.Group("", authMw)
	authOnly.Post("/api/v1/auth/logout", authH.Logout)
	authOnly.Get("/api/v1/auth/me", authH.Me)
	authOnly.Post("/api/v1/auth/change-password", authH.ChangePassword)
	authOnly.Get("/api/v1/satuan-kerja", satkerH.List)
	authOnly.Get("/api/v1/satuan-kerja/{id}", satkerH.Get)
	authOnly.Get("/api/v1/peran", rbacH.ListPeran)
	authOnly.Get("/api/v1/modul-akses", rbacH.ListModul)
	authOnly.Get("/api/v1/rbac/matriks", rbacH.Matriks)
	authOnly.Post("/api/v1/uploads", uploadH.Upload)
	authOnly.Get("/api/v1/notifikasi", monitoringH.ListNotifikasi)
	authOnly.Patch("/api/v1/notifikasi/{id}/baca", monitoringH.MarkNotifikasiRead)

	// ---------- dashboard_monitoring ----------
	dash := r.Group("", authMw, auditMw, middleware.RequireModule(database, "dashboard_monitoring"))
	dash.Get("/api/v1/dashboard/ringkasan", dashboardH.Ringkasan)
	dash.Get("/api/v1/dashboard/peta-sebaran", dashboardH.PetaSebaran)

	// ---------- pencarian_data ----------
	cari := r.Group("", authMw, auditMw, middleware.RequireModule(database, "pencarian_data"))
	cari.Get("/api/v1/laporan", laporanH.Search)
	cari.Get("/api/v1/laporan/{id}", laporanH.Get)
	cari.Get("/api/v1/kendaraan/search", kendaraanH.Search)

	// ---------- kelola_data ----------
	data := r.Group("", authMw, auditMw, middleware.RequireModule(database, "kelola_data"))
	data.Post("/api/v1/laporan", laporanH.Create)
	data.Put("/api/v1/laporan/{id}", laporanH.Update)
	data.Patch("/api/v1/laporan/{id}/status", laporanH.UpdateStatus)
	data.Delete("/api/v1/laporan/{id}", laporanH.Delete)

	data.Post("/api/v1/laporan/{id}/kendaraan", kendaraanH.Create)
	data.Put("/api/v1/kendaraan/{id}", kendaraanH.Update)
	data.Patch("/api/v1/kendaraan/{id}/status", kendaraanH.UpdateStatus)
	data.Delete("/api/v1/kendaraan/{id}", kendaraanH.Delete)

	data.Post("/api/v1/laporan/{id}/pihak-terkait", pihakH.Create)
	data.Put("/api/v1/pihak-terkait/{id}", pihakH.Update)
	data.Delete("/api/v1/pihak-terkait/{id}", pihakH.Delete)

	data.Post("/api/v1/laporan/{id}/barang-bukti", evidenceH.CreateBarangBukti)
	data.Delete("/api/v1/barang-bukti/{id}", evidenceH.DeleteBarangBukti)

	data.Post("/api/v1/laporan/{id}/dokumentasi", evidenceH.CreateDokumentasi)
	data.Delete("/api/v1/dokumentasi/{id}", evidenceH.DeleteDokumentasi)

	data.Post("/api/v1/laporan/{id}/riwayat-penyidikan", evidenceH.CreateRiwayatPenyidikan)
	data.Delete("/api/v1/riwayat-penyidikan/{id}", evidenceH.DeleteRiwayatPenyidikan)

	// ---------- laporan_analitik ----------
	analitik := r.Group("", authMw, auditMw, middleware.RequireModule(database, "laporan_analitik"))
	analitik.Get("/api/v1/dashboard/tren-bulanan", dashboardH.TrenBulanan)
	analitik.Get("/api/v1/dashboard/kategori-kasus", dashboardH.KategoriKasus)
	analitik.Get("/api/v1/laporan-periodik", monitoringH.ListLaporanPeriodik)
	analitik.Post("/api/v1/laporan-periodik", monitoringH.CreateLaporanPeriodik)
	analitik.Get("/api/v1/audit-log", monitoringH.ListAuditLog)

	// ---------- kelola_pengguna ----------
	pengguna := r.Group("", authMw, auditMw, middleware.RequireModule(database, "kelola_pengguna"))
	pengguna.Get("/api/v1/pengguna", penggunaH.List)
	pengguna.Get("/api/v1/pengguna/{id}", penggunaH.Get)
	pengguna.Post("/api/v1/pengguna", penggunaH.Create)
	pengguna.Put("/api/v1/pengguna/{id}", penggunaH.Update)
	pengguna.Delete("/api/v1/pengguna/{id}", penggunaH.Delete)
	pengguna.Post("/api/v1/pengguna/{id}/reset-password", penggunaH.ResetPassword)
	pengguna.Get("/api/v1/whitelist-whatsapp", whitelistH.List)
	pengguna.Post("/api/v1/whitelist-whatsapp", whitelistH.Create)
	pengguna.Patch("/api/v1/whitelist-whatsapp/{id}/status", whitelistH.UpdateStatus)
	pengguna.Delete("/api/v1/whitelist-whatsapp/{id}", whitelistH.Delete)

	// ---------- pengaturan_sistem ----------
	sistem := r.Group("", authMw, auditMw, middleware.RequireModule(database, "pengaturan_sistem"))
	sistem.Post("/api/v1/satuan-kerja", satkerH.Create)
	sistem.Put("/api/v1/satuan-kerja/{id}", satkerH.Update)
	sistem.Delete("/api/v1/satuan-kerja/{id}", satkerH.Delete)
	sistem.Put("/api/v1/rbac/matriks", rbacH.UpdateMatriks)
	sistem.Get("/api/v1/api-keys", apiKeyH.List)
	sistem.Post("/api/v1/api-keys", apiKeyH.Create)
	sistem.Delete("/api/v1/api-keys/{id}", apiKeyH.Revoke)

	// ---------- WA-AI service API (API key, bukan sesi pengguna) ----------
	wa := r.Group("/api/v1/wa", middleware.APIKeyRequired(database))
	wa.Post("/whitelist/check", waH.WhitelistCheck)
	wa.Post("/log-interaksi", waH.LogInteraksi)
	wa.Get("/log-interaksi", waH.ListLogInteraksi)
	wa.Post("/hasil-ocr", waH.HasilOcr)
	wa.Post("/search", waH.Search)
	wa.Post("/respons", waH.CreateRespons)

	// ---------- Berkas unggahan (dilindungi cookie sesi) ----------
	uploadsFS := http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir)))
	uploadsGroup := r.Group("", authMw)
	uploadsGroup.Handle("/uploads/", uploadsFS)

	// ---------- Dokumentasi API (Swagger UI) ----------
	docsDir := envOrDefault("DOCS_DIR", "/app/docs")
	pub.Handle("/openapi.yaml", http.FileServer(http.Dir(docsDir)))
	pub.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir(filepath.Join(docsDir, "swagger-ui")))))

	// ---------- Frontend statis (dashboard HTML/CSS/JS) ----------
	webDir := envOrDefault("WEB_DIR", "/app/web")
	pub.Handle("/", http.FileServer(http.Dir(webDir)))

	// ---------- Bungkus dengan middleware global ----------
	var handler http.Handler = r.ServeMux()
	handler = middleware.CORS(cfg.CORSOrigins)(handler)
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.AccessLog(handler)
	handler = middleware.Recover(handler)

	addr := ":" + cfg.HTTPPort
	log.Printf("[startup] CURANMOR AI API berjalan di %s (env=%s)", addr, cfg.AppEnv)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
