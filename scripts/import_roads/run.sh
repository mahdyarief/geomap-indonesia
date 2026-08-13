#!/usr/bin/env bash
# run.sh — pipeline impor jaringan jalan Jawa untuk routing (pgRouting).
#  1. Unduh java-latest.osm.pbf dari Geofabrik (jika belum ada).
#  2. Import via osm2pgsql (output pgsql default, tanpa Lua) → planet_osm_line.
#  3. Ekstrak jalan drivable → tabel java_ways (geometri SRID 4326).
#  4. Bangun topologi routing (pgr_nodeNetwork + pgr_createTopology + cost).
#
# Koneksi DB dibaca dari env (default sama dengan docker-compose).
set -euo pipefail

OSM_DIR="${OSM_DIR:-/home/dy/osm}"
PBF="${OSM_DIR}/java-latest.osm.pbf"
PBF_URL="${PBF_URL:-https://download.geofabrik.de/asia/indonesia/java-latest.osm.pbf}"

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-secret}"
DB_NAME="${DB_NAME:-geomapping_id}"

cd "$(dirname "$0")"

export PGPASSWORD="$DB_PASSWORD"
PG_ARGS="-h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"

echo "==> [1/3] Memastikan file OSM Jawa tersedia"
mkdir -p "$OSM_DIR"
if [ ! -s "$PBF" ]; then
    echo "    Mengunduh $PBF_URL ..."
    wget --progress=dot:giga -O "$PBF" "$PBF_URL"
else
    echo "    File sudah ada ($(du -h "$PBF" | cut -f1))"
fi

echo "==> [1b/4] Pastikan ekstensi hstore tersedia (dipakai --hstore)"
psql $PG_ARGS -v ON_ERROR_STOP=1 -tAc "CREATE EXTENSION IF NOT EXISTS hstore;" >/dev/null

echo "==> [2/4] Import seluruh data OSM ke Postgres (output pgsql default)"
# Catatan: paket Ubuntu osm2pgsql 2.2.0 tidak menyertakan modul Lua `osm2pgsql`
# untuk output flex, jadi kita pakai output pgsql default lalu memfilter jalan
# drivable dari planet_osm_line lewat SQL (build_java_ways.sql).
osm2pgsql \
    --create \
    --slim \
    --hstore \
    --database "$DB_NAME" \
    --host "$DB_HOST" \
    --port "$DB_PORT" \
    --username "$DB_USER" \
    "$PBF"

echo "==> [3/4] Ekstrak jalan drivable ke tabel java_ways"
psql $PG_ARGS -v ON_ERROR_STOP=1 -f build_java_ways.sql

echo "==> [4/4] Bangun topologi routing"
psql $PG_ARGS -v ON_ERROR_STOP=1 -f build_topology.sql

echo "==> Selesai. Graf routing siap dipakai: java_ways_noded"