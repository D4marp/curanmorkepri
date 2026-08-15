// Package db mengatur koneksi PostgreSQL dan migrasi skema.
package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/lib/pq"

	"curanmor-ai/internal/auth"
)

// Connect membuka connection pool ke PostgreSQL dengan retry, karena pada
// docker-compose kontainer API bisa start lebih cepat dari kontainer DB siap
// menerima koneksi.
func Connect(databaseURL string) (*sql.DB, error) {
	var dbConn *sql.DB
	var err error

	for attempt := 1; attempt <= 15; attempt++ {
		dbConn, err = sql.Open("postgres", databaseURL)
		if err == nil {
			err = dbConn.Ping()
		}
		if err == nil {
			break
		}
		log.Printf("[db] percobaan koneksi #%d gagal: %v (retry dalam 2s)", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal konek database setelah beberapa percobaan: %w", err)
	}

	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(10)
	dbConn.SetConnMaxLifetime(30 * time.Minute)
	return dbConn, nil
}

// migrationsTable memastikan tabel pelacak migrasi tersedia.
const migrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    filename    TEXT PRIMARY KEY,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);`

// RunMigrations menjalankan seluruh file .sql pada direktori migrationsDir
// secara berurutan (nama file diurutkan ascending), idempotent berkat tabel
// schema_migrations. Setiap file dijalankan dalam satu transaksi.
func RunMigrations(dbConn *sql.DB, migrationsDir string) error {
	if _, err := dbConn.Exec(migrationsTable); err != nil {
		return fmt.Errorf("gagal membuat tabel schema_migrations: %w", err)
	}

	files, err := listSQLFiles(migrationsDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		var already int
		if err := dbConn.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE filename = $1`, filepath.Base(f)).Scan(&already); err != nil {
			return fmt.Errorf("gagal cek status migrasi %s: %w", f, err)
		}
		if already > 0 {
			continue
		}

		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("gagal baca file migrasi %s: %w", f, err)
		}

		tx, err := dbConn.Begin()
		if err != nil {
			return fmt.Errorf("gagal mulai transaksi migrasi %s: %w", f, err)
		}
		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("gagal eksekusi migrasi %s: %w", f, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (filename) VALUES ($1)`, filepath.Base(f)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("gagal catat migrasi %s: %w", f, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("gagal commit migrasi %s: %w", f, err)
		}
		log.Printf("[db] migrasi diterapkan: %s", filepath.Base(f))
	}
	return nil
}

// RunSeedIfEmpty menjalankan file seed (folder terpisah, tidak dilacak di
// schema_migrations) hanya jika tabel satuan_kerja masih kosong — supaya
// aman dijalankan berulang kali tanpa menimpa data produksi.
func RunSeedIfEmpty(dbConn *sql.DB, seedDir string) error {
	var count int
	if err := dbConn.QueryRow(`SELECT COUNT(*) FROM satuan_kerja`).Scan(&count); err != nil {
		return fmt.Errorf("gagal cek data satuan_kerja: %w", err)
	}
	if count > 0 {
		log.Printf("[db] data satuan_kerja sudah ada (%d baris), lewati seeding", count)
		return nil
	}

	files, err := listSQLFiles(seedDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("gagal baca file seed %s: %w", f, err)
		}
		if _, err := dbConn.Exec(string(content)); err != nil {
			return fmt.Errorf("gagal eksekusi seed %s: %w", f, err)
		}
		log.Printf("[db] seed diterapkan: %s", filepath.Base(f))
	}

	if err := applySeedAdminOverride(dbConn); err != nil {
		return err
	}
	return nil
}

// applySeedAdminOverride mengganti NRP & password akun Super Admin pertama
// (dibuat oleh scripts/02_seed_rbac.sql dengan NRP tetap "00000000") memakai
// SEED_ADMIN_NRP / SEED_ADMIN_PASSWORD dari environment, bila di-set. Tanpa
// ini operator tidak pernah tahu kredensial awal selain dari kode sumber —
// dengan override, kredensial pertama ditentukan lewat .env deployment
// masing-masing, bukan nilai default yang sama di semua instalasi.
func applySeedAdminOverride(dbConn *sql.DB) error {
	nrp := os.Getenv("SEED_ADMIN_NRP")
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if nrp == "" && password == "" {
		return nil
	}
	if nrp == "" || password == "" {
		return fmt.Errorf("SEED_ADMIN_NRP dan SEED_ADMIN_PASSWORD harus diisi berdua atau dikosongkan berdua")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("gagal hash SEED_ADMIN_PASSWORD: %w", err)
	}

	res, err := dbConn.Exec(`
		UPDATE pengguna SET nrp = $1, password_hash = $2
		WHERE nrp = '00000000'
	`, nrp, hash)
	if err != nil {
		return fmt.Errorf("gagal terapkan SEED_ADMIN_NRP/SEED_ADMIN_PASSWORD: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("[db] kredensial Super Admin awal diganti sesuai SEED_ADMIN_NRP dari .env")
	}
	return nil
}

func listSQLFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".sql" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membaca direktori %s: %w", dir, err)
	}
	sort.Strings(files)
	return files, nil
}
