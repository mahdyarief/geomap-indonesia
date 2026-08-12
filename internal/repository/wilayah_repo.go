package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// WilayahRepository handles all wilayah-related queries.
type WilayahRepository struct {
	pool *pgxpool.Pool
}

func NewWilayahRepository(pool *pgxpool.Pool) *WilayahRepository {
	return &WilayahRepository{pool: pool}
}

// DetectLevel returns the administrative level based on the length of a
// (dot-less) kode: 2=provinsi, 4=kabupaten, 6=kecamatan, 10=kelurahan.
func DetectLevel(kode string) models.WilayahType {
	clean := strings.ReplaceAll(kode, ".", "")
	switch len(clean) {
	case 2:
		return models.TypeProvinsi
	case 4:
		return models.TypeKabupaten
	case 6:
		return models.TypeKecamatan
	case 10:
		return models.TypeKelurahan
	default:
		return ""
	}
}

type listRow struct {
	Kode string
	Nama string
}

// List returns wilayah entries filtered by type, parent and search term,
// with pagination. Returns the items and the total number of matches.
func (r *WilayahRepository) List(ctx context.Context, tipe models.WilayahType, parentID, search string, page, limit int) ([]models.WilayahListItem, int, error) {
	table, parentCol, err := tableFor(tipe)
	if err != nil {
		return nil, 0, err
	}

	var conds []string
	var args []any

	if parentID != "" && parentCol != "" {
		args = append(args, parentID)
		conds = append(conds, fmt.Sprintf("%s = $%d", parentCol, len(args)))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		conds = append(conds, fmt.Sprintf("nama ILIKE $%d", len(args)))
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", table, where)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count %s: %w", table, err)
	}

	offset := (page - 1) * limit
	args = append(args, limit, offset)
	query := fmt.Sprintf(
		"SELECT kode, nama FROM %s%s ORDER BY kode LIMIT $%d OFFSET $%d",
		table, where, len(args)-1, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list %s: %w", table, err)
	}
	defer rows.Close()

	items := make([]models.WilayahListItem, 0, limit)
	for rows.Next() {
		var row listRow
		if err := rows.Scan(&row.Kode, &row.Nama); err != nil {
			return nil, 0, fmt.Errorf("scan %s: %w", table, err)
		}
		items = append(items, models.WilayahListItem{
			Kode: row.Kode,
			Nama: row.Nama,
			Type: tipe,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Detail returns the full detail of a wilayah by kode, including parent info
// and optional metadata (luas, penduduk, logo, etc.).
func (r *WilayahRepository) Detail(ctx context.Context, kode string) (*models.WilayahDetail, error) {
	tipe := DetectLevel(kode)
	if tipe == "" {
		return nil, fmt.Errorf("invalid kode: %s", kode)
	}

	var d models.WilayahDetail
	d.Kode = strings.ReplaceAll(kode, ".", "")
	d.Type = tipe

	switch tipe {
	case models.TypeProvinsi:
		row := r.pool.QueryRow(ctx, `
			SELECT nama, lat, lng, luas, penduduk_total, penduduk_pria, penduduk_wanita,
			       logo_url, zona_waktu, elevasi
			FROM provinsi WHERE kode = $1`, d.Kode)
		err := row.Scan(
			&d.Nama, &d.CentroidLat, &d.CentroidLng, &d.Luas,
			&d.PendudukTotal, &d.PendudukPria, &d.PendudukWanita,
			&d.LogoURL, &d.ZonaWaktu, &d.Elevasi,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("detail provinsi: %w", err)
		}

	case models.TypeKabupaten:
		row := r.pool.QueryRow(ctx, `
			SELECT k.nama, k.lat, k.lng, k.luas, k.penduduk_total, k.penduduk_pria, k.penduduk_wanita,
			       k.logo_url, k.zona_waktu, k.elevasi, p.kode, p.nama
			FROM kabupaten k
			JOIN provinsi p ON k.provinsi_kode = p.kode
			WHERE k.kode = $1`, d.Kode)
		err := row.Scan(
			&d.Nama, &d.CentroidLat, &d.CentroidLng, &d.Luas,
			&d.PendudukTotal, &d.PendudukPria, &d.PendudukWanita,
			&d.LogoURL, &d.ZonaWaktu, &d.Elevasi,
			&d.ParentKode, &d.ParentNama,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("detail kabupaten: %w", err)
		}

	case models.TypeKecamatan:
		row := r.pool.QueryRow(ctx, `
			SELECT k.nama, k.lat, k.lng, kb.kode, kb.nama
			FROM kecamatan k
			JOIN kabupaten kb ON k.kabupaten_kode = kb.kode
			WHERE k.kode = $1`, d.Kode)
		err := row.Scan(&d.Nama, &d.CentroidLat, &d.CentroidLng, &d.ParentKode, &d.ParentNama)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("detail kecamatan: %w", err)
		}

	case models.TypeKelurahan:
		row := r.pool.QueryRow(ctx, `
			SELECT kl.nama, kl.lat, kl.lng, kc.kode, kc.nama
			FROM kelurahan kl
			JOIN kecamatan kc ON kl.kecamatan_kode = kc.kode
			WHERE kl.kode = $1`, d.Kode)
		err := row.Scan(&d.Nama, &d.CentroidLat, &d.CentroidLng, &d.ParentKode, &d.ParentNama)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("detail kelurahan: %w", err)
		}
	}

	d.BuildDetail()
	return &d, nil
}

// Children returns the direct children of a wilayah, e.g. kabupaten for a
// provinsi kode. Returns nil if the kode is already at the deepest level.
func (r *WilayahRepository) Children(ctx context.Context, kode string) ([]models.WilayahListItem, error) {
	tipe := DetectLevel(kode)
	if tipe == "" {
		return nil, fmt.Errorf("invalid kode: %s", kode)
	}

	var query string
	switch tipe {
	case models.TypeProvinsi:
		query = "SELECT kode, nama FROM kabupaten WHERE provinsi_kode = $1 ORDER BY kode"
	case models.TypeKabupaten:
		query = "SELECT kode, nama FROM kecamatan WHERE kabupaten_kode = $1 ORDER BY kode"
	case models.TypeKecamatan:
		query = "SELECT kode, nama FROM kelurahan WHERE kecamatan_kode = $1 ORDER BY kode"
	default:
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, query, kode)
	if err != nil {
		return nil, fmt.Errorf("children %s: %w", tipe, err)
	}
	defer rows.Close()

	childType := nextType(tipe)
	items := make([]models.WilayahListItem, 0, 32)
	for rows.Next() {
		var it models.WilayahListItem
		if err := rows.Scan(&it.Kode, &it.Nama); err != nil {
			return nil, err
		}
		it.Type = childType
		items = append(items, it)
	}
	return items, rows.Err()
}

func tableFor(tipe models.WilayahType) (table, parentCol string, err error) {
	switch tipe {
	case models.TypeProvinsi:
		return "provinsi", "", nil
	case models.TypeKabupaten:
		return "kabupaten", "provinsi_kode", nil
	case models.TypeKecamatan:
		return "kecamatan", "kabupaten_kode", nil
	case models.TypeKelurahan:
		return "kelurahan", "kecamatan_kode", nil
	default:
		return "", "", fmt.Errorf("invalid type: %s", tipe)
	}
}

func nextType(t models.WilayahType) models.WilayahType {
	switch t {
	case models.TypeProvinsi:
		return models.TypeKabupaten
	case models.TypeKabupaten:
		return models.TypeKecamatan
	case models.TypeKecamatan:
		return models.TypeKelurahan
	default:
		return ""
	}
}
