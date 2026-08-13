-- build_topology.sql — bangun topologi routing dari tabel `java_ways`
-- (hasil build_java_ways.sql, sumber: planet_osm_line).
--
-- CATATAN: data OSM sudah ter-noding secara alami (dua jalan yang bersimpangan
-- berbagi node yang sama, dipertahankan osm2pgsql sebagai vertex identik), jadi
-- kita LANGSUNG pgr_createTopology pada `java_ways` — tanpa pgr_nodeNetwork.
-- pgr_nodeNetwork secara fisik memecah garis via self-join ST_Crosses O(n²)
-- dan terbukti sangat lambat / rentan OOM untuk dataset besar (129k ruas > 40
-- menit). Menghasilkan java_ways berkolom source/target + cost (km), plus tabel
-- vertex java_ways_vertices_pgr, siap dipakai pgr_dijkstra.

-- 1) Pastikan ada kolom id unik + indeks geometri (untuk pgr_createTopology).
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS id BIGSERIAL PRIMARY KEY;
CREATE INDEX IF NOT EXISTS idx_java_ways_geom ON java_ways USING GIST (geom);

-- 2) Pastikan kolom source/target ada (pgr_createTopology mengisi keduanya).
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS source INTEGER;
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS target INTEGER;

-- 3) Buat topologi langsung pada java_ways (toleransi 0.0001° ~11 m untuk snap).
--    Membuat tabel java_ways_vertices_pgr dan mengisi source/target.
SELECT pgr_createTopology('java_ways', 0.0001, 'geom', 'id');

-- 3) Kost biaya = panjang dalam km (geodesik via SRID 4326 geography);
--    reverse_cost negatif untuk jalan one-way (satu arah).
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS cost DOUBLE PRECISION;
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS reverse_cost DOUBLE PRECISION;

UPDATE java_ways
SET cost = ST_Length(geom::geography) / 1000.0,
    reverse_cost = CASE
        WHEN oneway = 'yes' OR oneway = '1' OR oneway = 'true' THEN ST_Length(geom::geography) / 1000.0 * (-1)
        ELSE ST_Length(geom::geography) / 1000.0
    END;

-- 4) Indeks untuk pgr_dijkstra (lookup source/target) & snapping vertex.
CREATE INDEX IF NOT EXISTS idx_java_ways_source ON java_ways (source);
CREATE INDEX IF NOT EXISTS idx_java_ways_target ON java_ways (target);
CREATE INDEX IF NOT EXISTS idx_java_ways_vertices_the_geom ON java_ways_vertices_pgr USING GIST (the_geom);

-- 5) Ringkasan.
SELECT 'java_ways edges', count(*) FROM java_ways;