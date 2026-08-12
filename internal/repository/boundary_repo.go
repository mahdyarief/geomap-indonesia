package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GeoJSONFeature is a RFC 7946 GeoJSON Feature with wilayah properties.
type GeoJSONFeature struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"`
}

// BoundaryRepository handles boundary (geometry) retrieval.
type BoundaryRepository struct {
	pool *pgxpool.Pool
}

func NewBoundaryRepository(pool *pgxpool.Pool) *BoundaryRepository {
	return &BoundaryRepository{pool: pool}
}

// GetGeoJSON returns the boundary of a wilayah as a GeoJSON Feature, or
// ErrNotFound if the kode does not exist or has no geometry.
func (r *BoundaryRepository) GetGeoJSON(ctx context.Context, kode string) (*GeoJSONFeature, error) {
	tipe := DetectLevel(kode)
	table, _, err := tableFor(tipe)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT nama, ST_AsGeoJSON(geometry)
		FROM %s WHERE kode = $1`, table)

	var nama string
	var geom []byte
	err = r.pool.QueryRow(ctx, query, kode).Scan(&nama, &geom)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("boundary %s: %w", table, err)
	}

	feature := &GeoJSONFeature{
		Type: "Feature",
		Properties: map[string]any{
			"kode": kode,
			"nama": nama,
			"type": string(tipe),
		},
	}
	if len(geom) > 0 && string(geom) != "null" {
		feature.Geometry = json.RawMessage(geom)
	} else {
		feature.Geometry = json.RawMessage("null")
	}
	return feature, nil
}
