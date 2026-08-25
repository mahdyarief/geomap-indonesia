-- split_java.sql — split penuh java_ways (2,83 jt edge) di node persimpangan OSM.
-- Mekanisme sudah divalidasi di split_validate.sql pada subset Bintaro
-- (37.462 edge -> 62.559, komponen 10.910 -> 71, raksasa 8.135 -> 41.704 vertex).
-- Di sini mekanisme yang sama dijalankan untuk seluruh Java, menghasilkan
-- tabel java_ways_split yang sudah bertopologi, berbiaya, dan terkomponen.

\pset pager off

\echo '=== 1. NODE PERSIMPANGAN seluruh Java (dipakai >=2 way drivable) ==='
DROP TABLE IF EXISTS java_junctions;
CREATE TABLE java_junctions AS
WITH pairs AS (
    SELECT pw.id AS way_id, n.nid
    FROM planet_osm_ways pw
    CROSS JOIN LATERAL unnest(pw.nodes) AS n(nid)
    WHERE pw.id IN (SELECT DISTINCT osm_id FROM java_ways)
      AND pw.tags->>'highway' IN
          ('motorway','motorway_link','trunk','trunk_link',
           'primary','primary_link','secondary','secondary_link',
           'tertiary','tertiary_link','unclassified','residential',
           'service','living_street','road')
)
SELECT nid, count(DISTINCT way_id) AS n_ways
FROM pairs
GROUP BY nid
HAVING count(DISTINCT way_id) > 1;
ALTER TABLE java_junctions ADD PRIMARY KEY (nid);
ALTER TABLE java_junctions ADD COLUMN geom geometry(Point,4326);
UPDATE java_junctions j
SET geom = ST_SetSRID(ST_MakePoint(p.lon / 1e7::float8, p.lat / 1e7::float8), 4326)
FROM planet_osm_nodes p
WHERE p.id = j.nid;
SELECT count(*) AS junction_nodes_java FROM java_junctions;

\echo '=== 2. TITIK POTONG per way (ST_Collect) ==='
DROP TABLE IF EXISTS way_split_points;
CREATE TABLE way_split_points AS
SELECT pw.id AS osm_id, ST_Collect(jn.geom) AS pgeom
FROM planet_osm_ways pw
CROSS JOIN LATERAL unnest(pw.nodes) AS n(nid)
JOIN java_junctions jn ON jn.nid = n.nid
WHERE pw.id IN (SELECT DISTINCT osm_id FROM java_ways)
GROUP BY pw.id;
SELECT count(*) AS ways_with_junctions FROM way_split_points;

\echo '=== 3. SPLIT seluruh java_ways ==='
DROP TABLE IF EXISTS java_ways_split;
CREATE TABLE java_ways_split (
    osm_id bigint,
    name text,
    highway text,
    oneway text,
    geom geometry(LineString, 4326)
);
INSERT INTO java_ways_split (osm_id, name, highway, oneway, geom)
SELECT jw.osm_id, jw.name, jw.highway, jw.oneway, dmp.geom
FROM java_ways jw
LEFT JOIN way_split_points sp ON sp.osm_id = jw.osm_id
CROSS JOIN LATERAL (
    SELECT (ST_Dump(COALESCE(ST_Split(jw.geom, sp.pgeom), jw.geom))).geom AS geom
) dmp;
ALTER TABLE java_ways_split ADD COLUMN id BIGSERIAL PRIMARY KEY;
ALTER TABLE java_ways_split ADD COLUMN source INTEGER;
ALTER TABLE java_ways_split ADD COLUMN target INTEGER;
SELECT count(*) AS split_edges FROM java_ways_split;

\echo '=== 4. TOPOLOGI java_ways_split ==='
SELECT pgr_createTopology('java_ways_split', 0.0001, 'geom', 'id');

\echo '=== 5. COST (menit tempuh) + dist_km ==='
ALTER TABLE java_ways_split ADD COLUMN speed_kmh DOUBLE PRECISION;
ALTER TABLE java_ways_split ADD COLUMN dist_km DOUBLE PRECISION;
ALTER TABLE java_ways_split ADD COLUMN cost DOUBLE PRECISION;
ALTER TABLE java_ways_split ADD COLUMN reverse_cost DOUBLE PRECISION;

UPDATE java_ways_split SET speed_kmh = CASE highway
    WHEN 'motorway'      THEN 80
    WHEN 'motorway_link' THEN 40
    WHEN 'trunk'         THEN 70
    WHEN 'trunk_link'    THEN 40
    WHEN 'primary'       THEN 50
    WHEN 'primary_link'  THEN 30
    WHEN 'secondary'     THEN 40
    WHEN 'secondary_link' THEN 25
    WHEN 'tertiary'      THEN 35
    WHEN 'tertiary_link' THEN 20
    WHEN 'unclassified'  THEN 30
    WHEN 'residential'   THEN 20
    WHEN 'service'       THEN 15
    WHEN 'living_street' THEN 10
    ELSE 20 -- 'road' / lainnya
END;

UPDATE java_ways_split
SET dist_km = ST_Length(geom::geography) / 1000.0,
    cost = (ST_Length(geom::geography) / 1000.0) / speed_kmh * 60.0,
    reverse_cost = CASE
        WHEN oneway = 'yes' OR oneway = '1' OR oneway = 'true' THEN (ST_Length(geom::geography) / 1000.0) / speed_kmh * 60.0 * (-1)
        ELSE (ST_Length(geom::geography) / 1000.0) / speed_kmh * 60.0
    END;

\echo '=== 6. KOMPONEN TERHUBUNG (sesudah split) ==='
DROP TABLE IF EXISTS java_ways_components_split;
CREATE TABLE java_ways_components_split AS
SELECT seq, component, node AS vertex_id
FROM pgr_connectedComponents('SELECT id, source, target, cost FROM java_ways_split');
CREATE INDEX idx_comp_split_vertex ON java_ways_components_split (vertex_id);
CREATE INDEX idx_comp_split_component ON java_ways_components_split (component);

\echo '=== 7. RINGKASAN SESUDAH SPLIT ==='
SELECT 'java_ways_split edges', count(*) FROM java_ways_split;
WITH comp AS (
    SELECT component, count(*) AS cnt FROM java_ways_components_split GROUP BY component
)
SELECT
    (SELECT count(*) FROM comp)                                   AS n_components,
    (SELECT max(cnt) FROM comp)                                   AS giant_size,
    ROUND(100.0 * (SELECT max(cnt) FROM comp)::numeric /
          (SELECT sum(cnt)::numeric FROM comp), 1)                AS giant_pct,
    (SELECT count(*) FROM comp WHERE cnt <= 5)                    AS tiny_le5,
    (SELECT count(*) FROM comp WHERE cnt = 2)                     AS single_edge_components
FROM comp LIMIT 1;

\echo '=== 8. INDEKS untuk pgr_dijkstra ==='
CREATE INDEX idx_split_source ON java_ways_split (source);
CREATE INDEX idx_split_target ON java_ways_split (target);
CREATE INDEX idx_split_geom ON java_ways_split USING GIST (geom);
CREATE INDEX idx_split_vertices_geom ON java_ways_split_vertices_pgr USING GIST (the_geom);

\echo 'SELESAI.'
