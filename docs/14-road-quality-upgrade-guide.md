# Panduan: Upgrade Kualitas Jalan Banten & DKI Jakarta (Full Jawa, Semua Kelas Jalan)

**Document:** 14
**Status:** Siap deploy (kode & SQL sudah disiapkan, belum dieksekusi di mesin ini)
**Scope:** Rebuild graf routing pgRouting dengan (1) semua kelas jalan drivable, (2) cost berbasis kecepatan per kelas jalan, (3) perbaikan konektivitas via komponen terhubung.

---

## 1. Latar belakang & keputusan

Goal: meningkatkan kualitas routing jalan di provinsi Banten dan DKI Jakarta.
Keputusan user (2026-08-15):

| Aspek | Keputusan |
|---|---|
| Kelas jalan | Lengkapi jalan lokal (residential, unclassified, road, service, living_street) |
| Waktu tempuh | Realistis — kecepatan berbeda per kelas jalan, bukan konstan 40 km/jam |
| Konektivitas | Perbaiki — giant component sekarang hanya ~66% vertex |
| Cakupan data | **Full Jawa semua kelas jalan** (rute antar-provinsi tetap berfungsi) |

Catatan penting: **import TIDAK dieksekusi di mesin ini** (RAM hanya 14 GB, free 3 GB;
pass node 141 juta butuh berjam-jam). Semua persiapan kode/SQL sudah selesai dan
terverifikasi compile (`go build` & `go vet` PASS). Panduan ini untuk menjalankan
build di resource yang lebih besar.

---

## 2. Diagnosa kondisi saat ini (before)

- `java_ways` hanya berisi jalan arteri: motorway(_link), trunk(_link),
  primary(_link), secondary(_link), tertiary(_link) = **129.154 edge**.
- Jalan lokal **tidak ada** di graf: `unclassified`, `residential`, `road`,
  `service`, `living_street` — padahal mayoritas jaringan urban Banten/DKI adalah
  jalan lokal. Routing di permukiman hanya lompat antar arteri.
- Durasi API = `jarak_km / 40` (konstan) — tidak membedakan tol vs gang.
- Giant component = ~66% vertex; titik yang snap ke komponen kecil → 404.
- DB sebelum drop = 22 GB, ~99%-nya data mentah `planet_osm_*` (sisa osm2pgsql),
  bukan data routing. Setelah drop: **267 MB**.

---

## 3. Perubahan yang sudah disiapkan (belum dieksekusi)

### 3.1 `scripts/import_roads/build_java_ways.sql`
Filter diperluas dari arteri saja menjadi semua kelas drivable:

```sql
WHERE highway IN (
    'motorway', 'motorway_link', 'trunk', 'trunk_link',
    'primary', 'primary_link', 'secondary', 'secondary_link',
    'tertiary', 'tertiary_link',
    'unclassified', 'residential', 'road',
    'service', 'living_street'
);
```

### 3.2 `scripts/import_roads/build_topology.sql`
- `cost` kini **waktu tempuh (menit)** = `(panjang_km / speed_kmh) * 60`.
  pgr_dijkstra mengoptimalkan rute **tercepat**, bukan terpendek jarak.
- Kolom baru `speed_kmh` dan `dist_km` (panjang geodesik km).
- `reverse_cost` = -cost untuk one-way (arah dihormati pgr_dijkstra directed).
- Tabel `java_ways_components` dibangun ulang via `pgr_connectedComponents`,
  plus ringkasan top-5 komponen (ukuran + % dari semua vertex).

Kecepatan asumsi per kelas jalan (km/jam, campuran urban/rural Indonesia):

| highway | speed |
|---|---|
| motorway | 80 |
| motorway_link | 40 |
| trunk | 70 |
| trunk_link | 40 |
| primary | 50 |
| primary_link | 30 |
| secondary | 40 |
| secondary_link | 25 |
| tertiary | 35 |
| tertiary_link | 20 |
| unclassified | 30 |
| residential | 20 |
| service | 15 |
| living_street | 10 |
| road / lain | 20 |

### 3.3 `scripts/import_roads/run.sh` (pipeline 5 langkah)
1. Unduh PBF Geofabrik jika belum ada.
2. `osm2pgsql` pgsql output dengan **`--flat-nodes` + `--drop`** → cache node
   di file (bukan di DB), tabel middle dihapus setelah import. Ini mencegah
   DB membengkak seperti 22 GB dulu.
3. `psql -f build_java_ways.sql` — ekstrak semua kelas drivable → `java_ways`.
4. Drop tabel `planet_osm_*` (data mentah tidak dipakai runtime, hemat ~21 GB).
5. `psql -f build_topology.sql` — topologi + cost kecepatan + komponen.

### 3.4 Kode Go (backend)
- `internal/repository/distance_repo.go`
  - Struct `DistanceRoute` + field baru `DurationMinutes`.
  - Query SELECT: `SUM(w.dist_km)` untuk jarak, `SUM(rt.cost)` untuk durasi
    (cost sekarang menit). Kuery `pgr_dijkstra` tidak berubah (masih memakai
    `java_ways_components` giant-component snap).
- `internal/service/distance_service.go`
  - Konstanta `avgRoadSpeedKmH = 40` dan rumus `jarak/40` **dihapus**.
  - Durasi diambil langsung dari `route.DurationMinutes`.

Verifikasi: `go build ./...` dan `go vet ./...` PASS.

---

## 4. Syarat resource untuk menjalankan build (deploy di mesin besar)

### Minimum yang disarankan
| Resource | Kebutuhan | Catatan |
|---|---|---|
| RAM | **16 GB+** (disarankan 32 GB) | osm2pgsql `--cache 1024` + flat-nodes; `pgr_createTopology` pada jutaan edge butuh memori besar. Di mesin 14 GB proses ini terlalu ketat. |
| Disk | **40 GB free** (sementara), final ~2-4 GB | PBF 895 MB; `planet_osm_*` output sementara ±8 GB (polygon 6,4 GB); `java_ways` semua kelas diperkirakan 1-3 GB + index. Setelah step 4, DB menyusut drastis. |
| CPU | multi-core (4+) | Pass node 141 juta adalah bottleneck; disarankan disk SSD/NVMe. |
| OS | Ubuntu/Debian + PostGIS 15 + pgRouting | DB container memakai image `postgis/postgis:15-3.3` + `postgresql-15-pgrouting` (lihat `postgres/`). |

### Estimasi ukuran final DB
- `java_ways` (semua kelas, full Jawa): ~1,5-2,5 juta edge, ±2-4 GB dengan index.
- `java_ways_vertices_pgr` + `java_ways_components`: ± beberapa ratus MB.
- Batas administrasi (4 level) + kodepos: ±40 MB (tidak berubah).
- Total final: **±2-4 GB** (vs 22 GB sebelum drop raw OSM).

---

## 5. Langkah eksekusi (di mesin target)

```bash
# 0) Persiapkan PBF (895 MB, Geofabrik java-latest)
mkdir -p /home/dy/osm
#   jika belum ada: wget https://download.geofabrik.de/asia/indonesia/java-latest.osm.pbf

# 1) Jalankan pipeline lengkap (menggantikan java_ways lama)
cd scripts/import_roads
./run.sh
#   - import osm2pgsql pgsql output (--flat-nodes + --drop)
#   - build_java_ways.sql (semua kelas jalan)
#   - drop planet_osm_*
#   - build_topology.sql (topologi + cost menit + komponen)

# 2) Verifikasi hasil
psql ... -c "SELECT highway, count(*) FROM java_ways GROUP BY 1 ORDER BY 2 DESC;"
#   Harapannya: residential & unclassified jadi mayoritas, total 1,5-2,5 juta edge.

# 3) Rebuild & restart backend
go build ./...
docker compose build api && docker compose up -d api
```

Catatan: `run.sh` membaca env `OSM_DIR/PBF/DB_*`; default sudah disesuaikan.
Untuk menjalankan step manual (tanpa `run.sh`), ikuti urutan perintah di §3.3.

---

## 6. Verifikasi kualitas (setelah build)

1. **Distribusi kelas jalan**: `SELECT highway, count(*) FROM java_ways GROUP BY 1;`
   — residential/unclassified harus mendominasi (sebelumnya tidak ada sama sekali).
2. **Konektivitas**: bagian akhir `build_topology.sql` mencetak top-5 komponen
   dengan persentase. Giant component diharapkan jauh di atas 66% lama — target
   realistis 85%+ karena jalan lokal menyambungkan pocket-pockot jalan.
3. **Rute nyata Banten & DKI** (via API atau langsung pgr_dijkstra):
   - Jakarta Selatan → Tangerang (Banten)
   - Serang → Jakarta
   - Rute dalam permukiman (mis. kelurahan di Bekasi → Ciputat) yang dulu
     gagal/lompat jauh karena tidak ada jalan lokal.
4. **Durasi masuk akal**: cek `duration_minutes` — tol (motorway 80 km/h) harus
   jauh lebih cepat per km daripada residential (20 km/h).

---

## 7. Gotcha yang ditemukan (agar tidak terulang)

| Gotcha | Penjelasan |
|---|---|
| `osm2pgsql -h` = **help**, bukan host | Flag host yang benar adalah `-H`. Salah pakai `-h localhost` membuat program mencetak usage lalu keluar diam-diam (exit 0) — import tidak terjadi. |
| Paket Debian `osm2pgsql 2.2.0` tidak punya modul Lua `osm2pgsql` | Flex output (`-O flex -S routing.lua`) gagal dengan `module 'osm2pgsql' not found`. **Terpaksa pakai pgsql output + SQL filter** (`build_java_ways.sql`). routing.lua tetap disimpan sebagai referensi untuk mesin yang binary-nya mendukung flex. |
| JANGAN `pgr_nodeNetwork` untuk dataset besar | O(n²) `ST_Crosses`; 129k edge > 40 menit, rawan OOM. OSM sudah ter-noding alami → langsung `pgr_createTopology`. |
| Tabel `planet_osm_*` bikin DB 22 GB | Data mentah osm2pgsql tidak dipakai runtime. Drop setelah ekstrak `java_ways` (sudah masuk `run.sh` step 4). |
| Snap ke vertex terisolasi | `pgr_dijkstra` gagal diam-diam (404). Solusi: snap hanya ke giant component via `java_ways_components` (sudah ada di `distance_repo.go`). |
| Cost km vs menit | Aplikasi Go sudah disesuaikan: jarak dari `dist_km`, durasi dari `cost` (menit). Jangan mengembalikan cost ke format km tanpa menyelaraskan kedua sisi. |

---

## 8. Ringkasan file yang diubah

| File | Perubahan |
|---|---|
| `scripts/import_roads/build_java_ways.sql` | Filter kelas jalan diperluas ke semua drivable |
| `scripts/import_roads/build_topology.sql` | Cost = menit berbasis kecepatan per kelas; kolom `speed_kmh`, `dist_km`; rebuild `java_ways_components`; ringkasan top-5 komponen |
| `scripts/import_roads/run.sh` | `--flat-nodes` + `--drop`; drop `planet_osm_*`; pipeline 5 langkah |
| `internal/repository/distance_repo.go` | Field `DurationMinutes`; SELECT `dist_km` + `cost` |
| `internal/service/distance_service.go` | Hapus `avgRoadSpeedKmH`; durasi dari repo |

---

*Status deploy: belum dieksekusi di mesin ini (resource terbatas). Semua kode
terverifikasi compile; langkah eksekusi siap dijalankan di resource yang lebih besar.*