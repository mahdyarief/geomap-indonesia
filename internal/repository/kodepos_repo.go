package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
)

// KodeposRepository handles postal code queries.
type KodeposRepository struct {
	pool *pgxpool.Pool
}

func NewKodeposRepository(pool *pgxpool.Pool) *KodeposRepository {
	return &KodeposRepository{pool: pool}
}

// Lookup returns the wilayah associated with a postal code.
func (r *KodeposRepository) Lookup(ctx context.Context, kodepos string) (*models.KodeposLookup, error) {
	var res models.KodeposLookup
	row := r.pool.QueryRow(ctx, `
		SELECT kp.kode, kl.kode, kl.nama, kc.nama, kb.nama, p.nama
		FROM kodepos kp
		JOIN kelurahan kl ON kp.kelurahan_kode = kl.kode
		JOIN kecamatan kc ON kl.kecamatan_kode = kc.kode
		JOIN kabupaten kb ON kc.kabupaten_kode = kb.kode
		JOIN provinsi p ON kb.provinsi_kode = p.kode
		WHERE kp.kode = $1`, kodepos)

	err := row.Scan(
		&res.Kodepos,
		&res.Wilayah.Kode, &res.Wilayah.Nama,
		&res.Wilayah.Kecamatan, &res.Wilayah.Kabupaten, &res.Wilayah.Provinsi,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("kodepos lookup: %w", err)
	}
	return &res, nil
}

// ByWilayah returns all postal codes for a wilayah kode (kelurahan level).
func (r *KodeposRepository) ByWilayah(ctx context.Context, wilayahKode string) (*models.KodeposByWilayah, error) {
	tipe := DetectLevel(wilayahKode)
	// The kodepos table is keyed by kelurahan. For higher levels, resolve the
	// kode to its kelurahan children first.
	table, col, err := tableFor(tipe)
	if err != nil {
		return nil, err
	}
	// tableFor returns the parent FK column; here we need the level's own
	// kode column to match wilayahKode against the correct hierarchy level.
	col = levelKodeCol(tipe)
	query := fmt.Sprintf(`
		SELECT kp.kode
		FROM kodepos kp
		JOIN kelurahan kl ON kp.kelurahan_kode = kl.kode
		JOIN kecamatan kc ON kl.kecamatan_kode = kc.kode
		JOIN kabupaten kb ON kc.kabupaten_kode = kb.kode
		JOIN provinsi p ON kb.provinsi_kode = p.kode
		WHERE %s = $1 ORDER BY kp.kode`, col)

	rows, err := r.pool.Query(ctx, query, wilayahKode)
	if err != nil {
		return nil, fmt.Errorf("kodepos by wilayah (%s): %w", table, err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return nil, ErrNotFound
	}
	return &models.KodeposByWilayah{Kodepos: codes}, nil
}

// levelKodeCol maps a wilayah type to its own kode column alias as used in the
// ByWilayah join, so wilayahKode is matched against the correct hierarchy level.
func levelKodeCol(tipe models.WilayahType) string {
	switch tipe {
	case models.TypeProvinsi:
		return "p.kode"
	case models.TypeKabupaten:
		return "kb.kode"
	case models.TypeKecamatan:
		return "kc.kode"
	case models.TypeKelurahan:
		return "kl.kode"
	}
	return ""
}
