package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
)

// ReverseRepository handles reverse geocoding and search queries.
type ReverseRepository struct {
	pool poolQuerier
}

func NewReverseRepository(pool poolQuerier) *ReverseRepository {
	return &ReverseRepository{pool: pool}
}

// ReverseGeocode finds the most specific wilayah containing the given point,
// returning the full hierarchy. Requires imported geometry data.
//
// The boundaries source (cahyadsn/wilayah_boundaries) only ships geometry for
// provinsi, kabupaten and kecamatan, so the lookup falls back from the most
// specific level to the least. Each query shares the same 10-column scan
// shape; missing higher levels are NULL (scanned as zero values).
func (r *ReverseRepository) ReverseGeocode(ctx context.Context, lat, lng float64) (*models.ReverseResult, error) {
	var res models.ReverseResult
	res.Input = models.Centroid{Lat: lat, Lng: lng}

	point := "ST_SetSRID(ST_MakePoint($1, $2), 4326)"
	queries := []string{
		`SELECT kl.kode, kl.nama, kc.kode, kc.nama, kb.kode, kb.nama, p.kode, p.nama, kl.lat, kl.lng
		FROM kelurahan kl
		JOIN kecamatan kc ON kl.kecamatan_kode = kc.kode
		JOIN kabupaten kb ON kc.kabupaten_kode = kb.kode
		JOIN provinsi p ON kb.provinsi_kode = p.kode
		WHERE ST_Contains(kl.geometry, ` + point + `) LIMIT 1`,
		`SELECT '', '', kc.kode, kc.nama, kb.kode, kb.nama, p.kode, p.nama, kc.lat, kc.lng
		FROM kecamatan kc
		JOIN kabupaten kb ON kc.kabupaten_kode = kb.kode
		JOIN provinsi p ON kb.provinsi_kode = p.kode
		WHERE ST_Contains(kc.geometry, ` + point + `) LIMIT 1`,
		`SELECT '', '', '', '', kb.kode, kb.nama, p.kode, p.nama, kb.lat, kb.lng
		FROM kabupaten kb
		JOIN provinsi p ON kb.provinsi_kode = p.kode
		WHERE ST_Contains(kb.geometry, ` + point + `) LIMIT 1`,
		`SELECT '', '', '', '', '', '', p.kode, p.nama, p.lat, p.lng
		FROM provinsi p
		WHERE ST_Contains(p.geometry, ` + point + `) LIMIT 1`,
	}

	for _, q := range queries {
		err := r.pool.QueryRow(ctx, q, lng, lat).Scan(
			&res.Kelurahan.Kode, &res.Kelurahan.Nama,
			&res.Kecamatan.Kode, &res.Kecamatan.Nama,
			&res.Kabupaten.Kode, &res.Kabupaten.Nama,
			&res.Provinsi.Kode, &res.Provinsi.Nama,
			&res.CentroidLat, &res.CentroidLng,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reverse geocode: %w", err)
		}
		res.BuildReverse()
		return &res, nil
	}
	return nil, ErrNotFound
}

// KodeposByKelurahan returns all postal codes for a kelurahan kode.
func (r *ReverseRepository) KodeposByKelurahan(ctx context.Context, kelurahanKode string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT kode FROM kodepos WHERE kelurahan_kode = $1 ORDER BY kode", kelurahanKode)
	if err != nil {
		return nil, fmt.Errorf("kodepos by kelurahan: %w", err)
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
	return codes, rows.Err()
}

// Search finds wilayah by name across all levels using trigram similarity.
func (r *ReverseRepository) Search(ctx context.Context, q, tipe string, limit int) ([]models.SearchResult, error) {
	if q == "" {
		return nil, nil
	}

	queries := map[models.WilayahType]string{
		models.TypeProvinsi: `
			SELECT kode, nama FROM provinsi WHERE nama ILIKE '%' || $1 || '%'
			ORDER BY similarity(nama, $1) DESC, kode LIMIT $2`,
		models.TypeKabupaten: `
			SELECT k.kode, k.nama, p.kode, p.nama FROM kabupaten k
			JOIN provinsi p ON k.provinsi_kode = p.kode
			WHERE k.nama ILIKE '%' || $1 || '%'
			ORDER BY similarity(k.nama, $1) DESC, k.kode LIMIT $2`,
		models.TypeKecamatan: `
			SELECT k.kode, k.nama, kb.kode, kb.nama, p.nama
			FROM kecamatan k
			JOIN kabupaten kb ON k.kabupaten_kode = kb.kode
			JOIN provinsi p ON kb.provinsi_kode = p.kode
			WHERE k.nama ILIKE '%' || $1 || '%'
			ORDER BY similarity(k.nama, $1) DESC, k.kode LIMIT $2`,
		models.TypeKelurahan: `
			SELECT k.kode, k.nama, kc.kode, kc.nama, p.nama
			FROM kelurahan k
			JOIN kecamatan kc ON k.kecamatan_kode = kc.kode
			JOIN kabupaten kb ON kc.kabupaten_kode = kb.kode
			JOIN provinsi p ON kb.provinsi_kode = p.kode
			WHERE k.nama ILIKE '%' || $1 || '%'
			ORDER BY similarity(k.nama, $1) DESC, k.kode LIMIT $2`,
	}

	var results []models.SearchResult
	for t, query := range queries {
		if tipe != "" && string(t) != tipe {
			continue
		}
		rows, err := r.pool.Query(ctx, query, q, limit)
		if err != nil {
			return nil, fmt.Errorf("search %s: %w", t, err)
		}
		for rows.Next() {
			var s models.SearchResult
			switch t {
			case models.TypeProvinsi:
				err = rows.Scan(&s.Kode, &s.Nama)
			case models.TypeKabupaten:
				err = rows.Scan(&s.Kode, &s.Nama, &s.ParentKode, &s.ParentNama)
			case models.TypeKecamatan:
				err = rows.Scan(&s.Kode, &s.Nama, &s.ParentKode, &s.ParentNama, &s.Province)
			case models.TypeKelurahan:
				err = rows.Scan(&s.Kode, &s.Nama, &s.ParentKode, &s.ParentNama, &s.Province)
			}
			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan %s: %w", t, err)
			}
			s.Type = t
			s.BuildSearch()
			results = append(results, s)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	return results, nil
}
