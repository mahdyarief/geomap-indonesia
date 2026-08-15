package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DistanceRoute is the raw road-network result from pgRouting.
type DistanceRoute struct {
	SourceVertex    int64
	TargetVertex    int64
	DistanceKm      float64
	DurationMinutes float64
	Geometry        json.RawMessage
}

// ErrNoRoute is returned when pgRouting cannot connect origin and destination.
var ErrNoRoute = errors.New("no route")

// DistanceRepository handles road-network distance queries via pgRouting.
type DistanceRepository struct {
	pool *pgxpool.Pool
}

// NewDistanceRepository creates a DistanceRepository.
func NewDistanceRepository(pool *pgxpool.Pool) *DistanceRepository {
	return &DistanceRepository{pool: pool}
}

// Route returns the fastest road path between two coordinates (snapping to the
// nearest vertices in java_ways). Requires the graph built by
// scripts/import_roads. Edge cost is travel time in minutes (length/speed);
// dist_km is the geodesic length in kilometres.
func (r *DistanceRepository) Route(ctx context.Context, srcLat, srcLng, dstLat, dstLng float64) (*DistanceRoute, error) {
	const query = `
		WITH giant AS (
			SELECT component
			FROM java_ways_components
			GROUP BY component
			ORDER BY count(*) DESC
			LIMIT 1
		),
		snap AS (
			SELECT
				(SELECT id FROM java_ways_vertices_pgr v
				 JOIN java_ways_components c ON c.vertex_id = v.id
				 WHERE c.component = (SELECT component FROM giant)
				 ORDER BY v.the_geom <-> ST_SetSRID(ST_MakePoint($2, $1), 4326)
				 LIMIT 1) AS src,
				(SELECT id FROM java_ways_vertices_pgr v
				 JOIN java_ways_components c ON c.vertex_id = v.id
				 WHERE c.component = (SELECT component FROM giant)
				 ORDER BY v.the_geom <-> ST_SetSRID(ST_MakePoint($4, $3), 4326)
				 LIMIT 1) AS dst
		),
		route AS (
			SELECT *
			FROM pgr_dijkstra(
				'SELECT id, source, target, cost, reverse_cost FROM java_ways',
				(SELECT src FROM snap),
				(SELECT dst FROM snap),
				directed := true
			)
		)
		SELECT
			(SELECT src FROM snap)::bigint,
			(SELECT dst FROM snap)::bigint,
			COALESCE(SUM(w.dist_km), 0)::float8,
			COALESCE(SUM(rt.cost), 0)::float8,
			ST_AsGeoJSON(ST_LineMerge(ST_Collect(w.geom)))
		FROM route rt
		LEFT JOIN java_ways w ON w.id = rt.edge
		WHERE rt.edge <> -1
		  AND (SELECT src FROM snap) IS NOT NULL
		  AND (SELECT dst FROM snap) IS NOT NULL
		GROUP BY (SELECT src FROM snap), (SELECT dst FROM snap);
	`

	var (
		src, dst  int64
		distance  float64
		duration  float64
		geomJSON  []byte
	)
	err := r.pool.QueryRow(ctx, query, srcLat, srcLng, dstLat, dstLng).
		Scan(&src, &dst, &distance, &duration, &geomJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoRoute
	}
	if err != nil {
		return nil, fmt.Errorf("distance route: %w", err)
	}
	res := &DistanceRoute{
		SourceVertex:    src,
		TargetVertex:    dst,
		DistanceKm:      distance,
		DurationMinutes: duration,
	}
	if len(geomJSON) > 0 {
		res.Geometry = json.RawMessage(geomJSON)
	}
	return res, nil
}