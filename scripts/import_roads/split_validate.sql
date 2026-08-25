-- split_validate.sql — validasi mekanisme split post-hoc di subset Bintaro.
-- Tujuan: buktikan bahwa memecah java_ways di node persimpangan OSM asli
-- (node yang dipakai >=2 way drivable) menaikkan jumlah edge, menurunkan
-- jumlah komponen, dan menumbuhkan komponen raksasa pada subset tersebut.
-- Setelah ini terbukti, mekanisme yang sama dijalankan untuk seluruh Java.

-- Bbox Bintaro (sama dengan diag_before).
\echo '=== A. SUBSET java_ways di bbox Bintaro ==='
DROP TABLE IF EXISTS jw_bintaro;
CREATE TABLE jw_bintaro AS
SELECT * FROM java_ways
WHERE geom && ST_MakeEnvelope(106.68, -6.27, 106.85, -6.22, 4326);
ALTER TABLE jw_bintaro ADD COLUMN id BIGSERIAL PRIMARY KEY;
ALTER TABLE jw_bintaro ADD COLUMN source INTEGER;
ALTER TABLE jw_bintaro ADD COLUMN target INTEGER;
SELECT count(*) AS subset_edges FROM jw_bintaro;
SELECT count(DISTINCT osm_id) AS distinct_osm_ids FROM jw_bintaro;
-- pastikan osm_id selalu terhubung ke planet_osm_ways
SELECT count(*) AS orphan_osm_ids
FROM jw_bintaro j LEFT JOIN planet_osm_ways p ON p.id = j.osm_id
WHERE p.id IS NULL;

\echo '=== B. NODE PERSIMPANGAN pada subset ==='
DROP TABLE IF EXISTS jn_bintaro;
CREATE TABLE jn_bintaro AS
WITH pairs AS (
    SELECT pw.id AS way_id, n.nid
    FROM planet_osm_ways pw
    CROSS JOIN LATERAL unnest(pw.nodes) AS n(nid)
    WHERE pw.id IN (SELECT DISTINCT osm_id FROM jw_bintaro)
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
ALTER TABLE jn_bintaro ADD PRIMARY KEY (nid);
ALTER TABLE jn_bintaro ADD COLUMN geom geometry(Point,4326);
UPDATE jn_bintaro j
SET geom = ST_SetSRID(ST_MakePoint(p.lon / 1e7::float8, p.lat / 1e7::float8), 4326)
FROM planet_osm_nodes p
WHERE p.id = j.nid;
CREATE INDEX jn_bintaro_geom ON jn_bintaro USING GIST (geom);
SELECT count(*) AS junction_nodes_subset FROM jn_bintaro;

\echo '=== C. TOPOLOGI + KOMPONEN SEBELUM SPLIT (subset) ==='
SELECT pgr_createTopology('jw_bintaro', 0.0001, 'geom', 'id');
DROP TABLE IF EXISTS comp_before_bintaro;
CREATE TABLE comp_before_bintaro AS
SELECT seq, component, node AS vertex_id
FROM pgr_connectedComponents('SELECT id, source, target FROM jw_bintaro');
CREATE INDEX idx_comp_before_comp ON comp_before_bintaro (component);
WITH c AS (SELECT component, count(*) AS cnt FROM comp_before_bintaro GROUP BY component)
SELECT count(*) AS n_components, max(cnt) AS giant_size,
       count(*) FILTER (WHERE cnt <= 5) AS tiny_le5
FROM c;

\echo '=== D. SPLIT di node persimpangan ==='
DROP TABLE IF EXISTS jw_bintaro_split;
CREATE TABLE jw_bintaro_split (
    osm_id bigint,
    name text,
    highway text,
    oneway text,
    geom geometry(LineString, 4326)
);
INSERT INTO jw_bintaro_split (osm_id, name, highway, oneway, geom)
SELECT jw.osm_id, jw.name, jw.highway, jw.oneway, dmp.geom
FROM jw_bintaro jw
CROSS JOIN LATERAL (
    SELECT (ST_Dump(COALESCE(ST_Split(jw.geom, pts.pgeom), jw.geom))).geom AS geom
    FROM (
        SELECT ST_Collect(jn.geom) AS pgeom
        FROM planet_osm_ways pw
        CROSS JOIN LATERAL unnest(pw.nodes) AS n(nid)
        JOIN jn_bintaro jn ON jn.nid = n.nid
        WHERE pw.id = jw.osm_id
    ) pts
) dmp;
ALTER TABLE jw_bintaro_split ADD COLUMN id BIGSERIAL PRIMARY KEY;
ALTER TABLE jw_bintaro_split ADD COLUMN source INTEGER;
ALTER TABLE jw_bintaro_split ADD COLUMN target INTEGER;
SELECT count(*) AS split_edges FROM jw_bintaro_split;

\echo '=== E. TOPOLOGI + KOMPONEN SESUDAH SPLIT (subset) ==='
SELECT pgr_createTopology('jw_bintaro_split', 0.0001, 'geom', 'id');
DROP TABLE IF EXISTS comp_after_bintaro;
CREATE TABLE comp_after_bintaro AS
SELECT seq, component, node AS vertex_id
FROM pgr_connectedComponents('SELECT id, source, target FROM jw_bintaro_split');
CREATE INDEX idx_comp_after_comp ON comp_after_bintaro (component);
WITH c AS (SELECT component, count(*) AS cnt FROM comp_after_bintaro GROUP BY component)
SELECT count(*) AS n_components, max(cnt) AS giant_size,
       count(*) FILTER (WHERE cnt <= 5) AS tiny_le5
FROM c;

\echo '=== F. BUKTI PERSIMPANGAN MENJADI VERTEX ==='
-- jumlah vertex ber-derajat >= 3 (persimpangan sejati) sebelum vs sesudah
SELECT 'before' AS stage, count(*) AS vertices_deg_ge3 FROM (
    SELECT v FROM (SELECT source AS v FROM jw_bintaro UNION ALL SELECT target AS v FROM jw_bintaro) t
    GROUP BY v HAVING count(*) >= 3
) x
UNION ALL
SELECT 'after' AS stage, count(*) AS vertices_deg_ge3 FROM (
    SELECT v FROM (SELECT source AS v FROM jw_bintaro_split UNION ALL SELECT target AS v FROM jw_bintaro_split) t
    GROUP BY v HAVING count(*) >= 3
) x;

-- contoh satu pasangan way yang sebelumnya bersilangan (ST_Crosses),
-- sekarang harus berbagi vertex setelah split
\echo '=== G. CONTOH PASANGAN BERSILANGAN ==='
WITH pair AS (
    SELECT a.id AS aid, b.id AS bid
    FROM jw_bintaro a
    JOIN jw_bintaro b ON a.id < b.id AND a.geom && b.geom
    WHERE ST_Crosses(a.geom, b.geom)
    LIMIT 1
)
SELECT p.aid, p.bid,
       ST_AsText(ST_Intersection(a.geom, b.geom)) AS crossing_point,
       (SELECT count(*) FROM (
            SELECT source AS v FROM jw_bintaro_split WHERE osm_id = a.osm_id
            UNION
            SELECT target AS v FROM jw_bintaro_split WHERE osm_id = a.osm_id
            INTERSECT
            SELECT source AS v FROM jw_bintaro_split WHERE osm_id = b.osm_id
            UNION
            SELECT target AS v FROM jw_bintaro_split WHERE osm_id = b.osm_id
       ) sh) AS shared_vertices
FROM pair p
JOIN jw_bintaro a ON a.id = p.aid
JOIN jw_bintaro b ON b.id = p.bid;
