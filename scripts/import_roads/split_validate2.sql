-- split_validate2.sql — bandingkan komponen sebelum vs sesudah split (subset Bintaro).
-- Mekanisme split sudah terbukti di split_validate.sql; di sini hanya menambahkan
-- kolom cost yang diminta pgr_connectedComponents v3.8 (wajib ada di SELECT).

ALTER TABLE jw_bintaro_split ADD COLUMN IF NOT EXISTS cost DOUBLE PRECISION DEFAULT 1;

\echo '=== KOMPONEN SEBELUM SPLIT (subset Bintaro) ==='
DROP TABLE IF EXISTS comp_before_bintaro;
CREATE TABLE comp_before_bintaro AS
SELECT seq, component, node AS vertex_id
FROM pgr_connectedComponents('SELECT id, source, target, cost FROM jw_bintaro');
CREATE INDEX idx_comp_before_comp ON comp_before_bintaro (component);
WITH c AS (SELECT component, count(*) AS cnt FROM comp_before_bintaro GROUP BY component)
SELECT count(*) AS n_components, max(cnt) AS giant_size,
       count(*) FILTER (WHERE cnt <= 5) AS tiny_le5
FROM c;

\echo '=== KOMPONEN SESUDAH SPLIT (subset Bintaro) ==='
DROP TABLE IF EXISTS comp_after_bintaro;
CREATE TABLE comp_after_bintaro AS
SELECT seq, component, node AS vertex_id
FROM pgr_connectedComponents('SELECT id, source, target, cost FROM jw_bintaro_split');
CREATE INDEX idx_comp_after_comp ON comp_after_bintaro (component);
WITH c AS (SELECT component, count(*) AS cnt FROM comp_after_bintaro GROUP BY component)
SELECT count(*) AS n_components, max(cnt) AS giant_size,
       count(*) FILTER (WHERE cnt <= 5) AS tiny_le5
FROM c;
