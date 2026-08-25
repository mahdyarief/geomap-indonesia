-- build_topology.sql — bangun topologi routing dari tabel `java_ways`
-- (hasil routing.lua / osm2pgsql flex: seluruh kelas jalan drivable).
--
-- CATATAN (2026-08-25): routing.lua semula mengimpor dengan
-- split_at_way_intersections=true, namun opsi itu TIDAK PERNAH ADA di
-- osm2pgsql 2.3.1 dan diabaikan diam-diam oleh parser flex — akibatnya
-- java_ways sangat terfragmentasi (1,48 jt komponen, raksasa hanya 6,5%
-- vertex). Pipeline yang benar: jalankan split_java.sql dulu untuk memecah
-- java_ways di node persimpangan OSM asli menjadi java_ways_split, BARU
-- pgr_createTopology pada java_ways_split. (pgr_nodeNetwork tidak dipakai:
-- self-join ST_Crosses O(n²), sangat lambat untuk dataset sebesar ini.)
-- Menghasilkan java_ways berkolom source/target + cost (waktu tempuh menit) +
-- dist_km, plus tabel vertex java_ways_vertices_pgr, siap dipakai pgr_dijkstra.

-- 1) Pastikan ada kolom id unik + indeks geometri (untuk pgr_createTopology).
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS id BIGSERIAL PRIMARY KEY;
CREATE INDEX IF NOT EXISTS idx_java_ways_geom ON java_ways USING GIST (geom);

-- 2) Pastikan kolom source/target ada (pgr_createTopology mengisi keduanya).
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS source INTEGER;
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS target INTEGER;

-- 3) Buat topologi langsung pada java_ways (toleransi 0.0001° ~11 m untuk snap).
--    Membuat tabel java_ways_vertices_pgr dan mengisi source/target.
SELECT pgr_createTopology('java_ways', 0.0001, 'geom', 'id');

-- 4) Cost = WAKTU TEMPUH (menit), bukan jarak: pgr_dijkstra mengoptimalkan
--    rute tercepat, bukan terpendek jarak. Kecepatan asumsi per kelas jalan
--    (km/jam), campuran urban/rural Indonesia.
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS speed_kmh DOUBLE PRECISION;
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS dist_km DOUBLE PRECISION;
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS cost DOUBLE PRECISION;
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS reverse_cost DOUBLE PRECISION;

UPDATE java_ways SET speed_kmh = CASE highway
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

UPDATE java_ways
SET dist_km = ST_Length(geom::geography) / 1000.0,
    cost = (ST_Length(geom::geography) / 1000.0) / speed_kmh * 60.0,
    reverse_cost = CASE
        WHEN oneway = 'yes' OR oneway = '1' OR oneway = 'true' THEN (ST_Length(geom::geography) / 1000.0) / speed_kmh * 60.0 * (-1)
        ELSE (ST_Length(geom::geography) / 1000.0) / speed_kmh * 60.0
    END;

-- 5) Komponen terhubung (untuk giant-component lookup di API + laporan konektivitas).
DROP TABLE IF EXISTS java_ways_components;
CREATE TABLE java_ways_components AS
SELECT seq, component, node AS vertex_id
FROM pgr_connectedComponents('SELECT id, source, target, cost, reverse_cost FROM java_ways');
CREATE INDEX IF NOT EXISTS idx_java_ways_components_vertex ON java_ways_components (vertex_id);
CREATE INDEX IF NOT EXISTS idx_java_ways_components_component ON java_ways_components (component);

-- 6) Indeks untuk pgr_dijkstra (lookup source/target) & snapping vertex.
CREATE INDEX IF NOT EXISTS idx_java_ways_source ON java_ways (source);
CREATE INDEX IF NOT EXISTS idx_java_ways_target ON java_ways (target);
CREATE INDEX IF NOT EXISTS idx_java_ways_vertices_the_geom ON java_ways_vertices_pgr USING GIST (the_geom);

-- 7) Ringkasan.
SELECT 'java_ways edges', count(*) FROM java_ways;
WITH comp AS (
    SELECT component, count(*) AS cnt FROM java_ways_components GROUP BY component
)
SELECT component, cnt, ROUND(100.0 * cnt::numeric / (SELECT sum(cnt)::numeric FROM comp), 1) AS pct_vertices
FROM comp ORDER BY cnt DESC LIMIT 5;