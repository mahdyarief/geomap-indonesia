-- build_java_ways.sql — ekstrak jalan yang dapat dilalui kendaraan dari tabel
-- `planet_osm_line` (hasil osm2pgsql output pgsql default) ke tabel `java_ways`
-- dengan geometri SRID 4326, siap diproses pgRouting.
--
-- Kolom `way` dari osm2pgsql pgsql default ber-SRID 3857 (Web Mercator), jadi
-- kita transform ke 4326 agar cost geodesik (ST_Length(geom::geography)) benar.

DROP TABLE IF EXISTS java_ways;

CREATE TABLE java_ways AS
SELECT
    osm_id,
    name,
    highway,
    oneway,
    ST_Transform(way, 4326) AS geom
FROM planet_osm_line
WHERE highway IN (
    'motorway',      'motorway_link',
    'trunk',         'trunk_link',
    'primary',       'primary_link',
    'secondary',     'secondary_link',
    'tertiary',      'tertiary_link',
    'unclassified',
    'residential',
    'road',
    'service',
    'living_street'
);

-- Pastikan kolom id + indeks geometri (digunakan build_topology.sql).
ALTER TABLE java_ways ADD COLUMN IF NOT EXISTS id BIGSERIAL PRIMARY KEY;
CREATE INDEX IF NOT EXISTS idx_java_ways_geom ON java_ways USING GIST (geom);

SELECT 'java_ways rows', count(*) FROM java_ways;