package repo

import (
	"database/sql"

	"curanmor-ai/internal/models"
)

func ListAPIKeys(db *sql.DB) ([]models.WaAPIKey, error) {
	rows, err := db.Query(`SELECT id, nama_layanan, key_prefix, status_aktif, dibuat_oleh, last_used_at, created_at FROM wa_api_key ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.WaAPIKey
	for rows.Next() {
		var k models.WaAPIKey
		if err := rows.Scan(&k.ID, &k.NamaLayanan, &k.KeyPrefix, &k.StatusAktif, &k.DibuatOleh, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func CreateAPIKey(db *sql.DB, namaLayanan, keyHash, keyPrefix string, dibuatOleh int64) (int64, error) {
	var id int64
	err := db.QueryRow(
		`INSERT INTO wa_api_key (nama_layanan, key_hash, key_prefix, dibuat_oleh) VALUES ($1,$2,$3,$4) RETURNING id`,
		namaLayanan, keyHash, keyPrefix, dibuatOleh,
	).Scan(&id)
	return id, err
}

func RevokeAPIKey(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE wa_api_key SET status_aktif = false WHERE id = $1`, id)
	return err
}
