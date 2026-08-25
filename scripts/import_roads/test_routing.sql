-- test_routing.sql — uji rute Bintaro → SCBD via pgRouting.
-- Prasyarat: java_ways sudah punya topologi (build_topology.sql selesai).
-- Jalankan: psql -h 172.80.1.2 -U postgres -d geomapping_id -f test_routing.sql

\pset pager off

\echo '=== 1) Snap Bintaro (Jalan Cucur Timur X) ke vertex terdekat ==='
SELECT v.id AS bintaro_vid,
       ROUND(ST_Y(v.the_geom)::numeric, 6) AS lat,
       ROUND(ST_X(v.the_geom)::numeric, 6) AS lon,
       ROUND(ST_Distance(v.the_geom::geography,
             ST_SetSRID(ST_MakePoint(106.742761, -6.263799), 4326)::geography)::numeric, 1) AS snap_m
FROM java_ways_vertices_pgr v
ORDER BY v.the_geom <-> ST_SetSRID(ST_MakePoint(106.742761, -6.263799), 4326)
LIMIT 1 \gset
\echo '  -> bintaro_vid = :bintaro_vid'

\echo '=== 2) Snap SCBD (jalan bernama Sudirman Central Business District) ==='
WITH scbd_way AS (
    SELECT id, geom FROM java_ways
    WHERE name ILIKE '%Sudirman Central Business District%'
    ORDER BY ST_Length(geom::geography) DESC
    LIMIT 1
)
SELECT v.id AS scbd_vid,
       ROUND(ST_Y(v.the_geom)::numeric, 6) AS lat,
       ROUND(ST_X(v.the_geom)::numeric, 6) AS lon,
       ROUND(ST_Distance(v.the_geom::geography, w.geom::geography)::numeric, 1) AS snap_m
FROM java_ways_vertices_pgr v, scbd_way w
ORDER BY ST_Distance(v.the_geom, w.geom)
LIMIT 1 \gset
\echo '  -> scbd_vid = :scbd_vid'

\echo '=== 3) Cek komponen terhubung ==='
SELECT COUNT(DISTINCT component) = 1 AS same_component
FROM java_ways_components
WHERE vertex_id IN (:bintaro_vid, :scbd_vid);

\echo '=== 4) Dijkstra (directed) Bintaro -> SCBD ==='
SELECT COUNT(*) AS n_edges,
       COALESCE(ROUND(SUM(w.dist_km)::numeric, 2), 0) AS jarak_km,
       COALESCE(ROUND(SUM(r.cost)::numeric, 1), 0)    AS waktu_menit
FROM pgr_dijkstra(
    'SELECT id, source, target, cost, reverse_cost FROM java_ways',
    :bintaro_vid, :scbd_vid, directed := true
) r
JOIN java_ways w ON w.id = r.edge;

\echo '=== 5) Rute (urutan jalan) ==='
SELECT ROW_NUMBER() OVER () AS no, w.name AS jalan, w.highway,
       ROUND(w.dist_km::numeric, 3) AS km, ROUND(r.cost::numeric, 1) AS menit
FROM pgr_dijkstra(
    'SELECT id, source, target, cost, reverse_cost FROM java_ways',
    :bintaro_vid, :scbd_vid, directed := true
) r
JOIN java_ways w ON w.id = r.edge
WHERE w.name IS NOT NULL AND w.name <> ''
ORDER BY r.seq;
