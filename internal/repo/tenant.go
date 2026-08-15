// Package repo adalah lapisan akses data (SQL murni via database/sql +
// lib/pq), satu file per domain entitas.
package repo

import "database/sql"

// ResolveSatkerScope mengembalikan daftar satker_id yang boleh dilihat oleh
// pengguna berdasarkan posisinya di hirarki satuan_kerja:
//   - Polda            -> nil (tanpa filter, lihat semua satker)
//   - Polres/Polresta   -> satker sendiri + seluruh Polsek jajarannya
//   - Polsek            -> satker sendiri saja
//
// nil berarti "tanpa batas" (harus dicek eksplisit oleh pemanggil), slice
// kosong tidak pernah dikembalikan untuk kasus valid (minimal berisi diri
// sendiri).
func ResolveSatkerScope(db *sql.DB, satkerID int64) ([]int64, error) {
	var jenis string
	if err := db.QueryRow(`SELECT jenis_satker FROM satuan_kerja WHERE id = $1`, satkerID).Scan(&jenis); err != nil {
		return nil, err
	}
	if jenis == "Polda" {
		return nil, nil // Polda: tanpa filter
	}
	rows, err := db.Query(`SELECT id FROM satuan_kerja WHERE id = $1 OR induk_id = $1`, satkerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
