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
