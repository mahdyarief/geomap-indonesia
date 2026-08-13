# Image postgres custom: postgis/postgis:15-3.3 (Debian-based) + pgRouting.
# pgRouting dipakai untuk perhitungan jarak via jalan (pgr_dijkstra/pgr_astar)
# pada backend, sehingga tidak butuh service routing terpisah.
FROM postgis/postgis:15-3.3

# Routing untuk PostgreSQL 15. Image ini Debian (bookworm), paket tersedia via
# apt image (bedakan dengan apt Ubuntu host yang tidak punya paket ini).
RUN apt-get update \
    && apt-get install -y --no-install-recommends postgresql-15-pgrouting postgresql-15-pgrouting-scripts \
    && rm -rf /var/lib/apt/lists/*

# Init script: membuat ekstensi pgrouting otomatis saat volume data pertama
# kali diinisialisasi (fresh init). Untuk volume yang sudah ada, CREATE
# EXTENSION dijalankan manual / via migrasi.
COPY init-pgrouting.sh /docker-entrypoint-initdb.d/10-init-pgrouting.sh
RUN chmod +x /docker-entrypoint-initdb.d/10-init-pgrouting.sh