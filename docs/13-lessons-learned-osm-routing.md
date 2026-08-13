# Lesson Learned: Self-Hosted Road Network Routing (pgRouting)

**Document:** 13
**Status:** Implemented
**Scope:** Distance routing via `pgr_dijkstra` atas data jalan OSM (ponsel: Pulau Java).

---

## Overview

Dokumen ini merangkum pelajaran lapangan dari membangun routing jarak jalan
road-network secara self-hosted memakai pgRouting di dalam PostGIS yang sudah
dimiliki project. Tujuannya jadi referensi cepat ketika ingin **men-deploy ulang**
atau memperluas dataset (misalnya full Indonesia) agar tidak mengulang kesalahan
yang sudah pernah ditemukan.

---

## 1. JANGAN pakai `pgr_nodeNetwork` untuk dataset besar

### Masalah
`pgr_nodeNetwork` memecah garis yang saling berpotongan secara fisik lewat
self-join `ST_Crosses` (kompleksitas O(n²)). Pada ~129.154 ruas jalan (`java_ways`)
proses ini **makan waktu > 40 menit** dan rawan OOM (lihat log `bw45q5s9b`,
`b8iluagx7`: `steps={"noded": false}` selama puluhan menit).

### Solusi
Data OSM dari osm2pgsql **sudah ter-noding secara alami**: dua jalan yang
bersimpangan berbagi node yang sama (dipertahankan sebagai vertex identik).
Karena itu cukup jalankan `pgr_createTopology` **langsung** tanpa tahap noding:

```sql
SELECT pgr_createTopology('java_ways', 0.0001, 'geom', 'id');
```

Hasil: 129.154 edges diproses dan topologi terbentuk dalam hitungan menit (log
`bpo90l784`). Skrip lengkap ada di `scripts/import_roads/build_topology.sql`.

> **Key takeaway:** Untuk data dari osm2pgsql, hasilnya *sudah* berupa graph.
> `pgr_nodeNetwork` hanya diperlukan jika sumber data adalah garis yang belum
> tersambung di persimpangan. Selalu cek dulu, jangan otomatis asumsi harus noding.

---

## 2. Snap ke nearest vertex bisa bikin `NOT_FOUND` — pakai component-aware snap

### Masalah
Query awal men-snap origin/destination ke vertex terdekat di seluruh
`java_ways_vertices_pgr`. Ternyata graf jalan **tidak fully connected**: ada
ribuan komponen terisolasi (pocket jalan kecil, `tertiary_link` rusak, dsb.).
Jika salah satu titik snap mendarat di komponen kecil yang terisolasi,
`pgr_dijkstra` tidak bisa menemukan rute → HTTP `404 NOT_FOUND` padahal dua titik
itu jelas berada di jalan raya besar (kasus nyata: Bandung ter-snap ke vertex
`17391` milik komponen terisolasi sendiri).

### Solusi
1. Buat tabel komponen terkoneksi yang persisten:
   ```sql
   CREATE TABLE java_ways_components AS
   SELECT seq, component, node AS vertex_id
   FROM pgr_connectedComponents(
       'SELECT id, source, target, cost, reverse_cost FROM java_ways'
   );
   -- indeks pada vertex_id dan component
   ```
2. Batasi snapping hanya ke **giant component** (komponen terbesar). ID giant
   component dihitung dinamis dalam query via CTE `giant` (group by component,
   order by count desc, limit 1), lalu `JOIN java_ways_components` menyaring snap.

Dengan begitu kedua endpoint dijamin berada dalam satu komponen yang sama dan
selalu tersambung. Implementasi ada di `internal/repository/distance_repo.go`
(CTE `giant` + `snap`).

> **Key takeaway:** Grafik jalan OSM hampir selalu punya banyak komponen
> terisolasi. Selalu snap ke komponen raksasa (atau minimal pastikan dua endpoint
> satu komponen), kalau tidak routing bakal gagal diam-diam jika titik kebetulan
> menempel di komponen kecil.

---

## 3. Urutan build & skrip impor

Pipeline data jalan (dari `planet_osm_line`):

```
osm2pgsql
  → planet_osm_line
  → scripts/import_roads/build_java_ways.sql      (filter jenis jalan → java_ways)
  → scripts/import_roads/build_topology.sql       (pgr_createTopology + cost + indeks)
  → (manual one-off) java_ways_components         (pgr_connectedComponents)
```

`build_java_ways.sql` memfilter jenis jalan. Status AKTUAL di database
(cek langsung, `SELECT highway, count(*) FROM java_ways GROUP BY 1`): hanya
`motorway(_link)`, `trunk(_link)`, `primary(_link)`, `secondary(_link)`,
`tertiary(_link)` — **tanpa** `service`, `living_street`, `residential`,
`unclassified`, maupun `road`. Yang `residential`/`unclassified`/`road` memang
di-comment untuk referensi pengembangan berikutnya; kerjakan saat cakupan jalan
yang lebih lengkap (jalan residen) dibutuhkan.

Format cost:
```sql
cost         = ST_Length(geom::geography) / 1000.0   -- km, geodesik
reverse_cost = cost * (-1)  jika oneway YES;  else cost
```

---

## 4. Rebuild/restart backend (Docker)

Backend berjalan di Docker Compose. Setelah ubah kode Go:

```bash
go build ./...                                   # cek kompilasi dulu
docker compose build api                         # rebuild image
docker compose up -d api                         # restart dengan image baru
```

Pastikan query SQL-nya juga diverifikasi langsung ke DB sebelum rebuild, supaya
error SQL tidak baru ketahuan di runtime.

---

## 5. Akses database (docker)

```bash
PGPASSWORD=secret psql -h localhost -p 5432 -U postgres -d geomapping_id
```

- host: `localhost`, port: `5432`, db: `geomapping_id`, user/pass: `postgres/secret`.

---

## 6. Menjalankan / menguji API distance

1. **Auth dulu** (token ada di level teratas, bukan `data.token`):
   ```bash
   curl -s -X POST http://localhost:8080/api/v1/auth \
     -H 'Content-Type: application/json' \
     -d '{"public_key":"pk_test_12345","private_key":"sk_test_67890"}'
   # → {"token":"..."}
   ```
2. **Hitung jarak**:
   ```bash
   curl -s -X POST http://localhost:8080/api/v1/distance \
     -H 'Authorization: Bearer <TOKEN>' \
     -H 'Content-Type: application/json' \
     -d '{"origin":{"lat":-6.1754,"lng":106.8272},"destination":{"lat":-6.9127,"lng":107.6090}}'
   # → {"distance_km":158.85,"duration_minutes":..,"geometry":"{...}"}
   ```

### Validasi input (ditambahkan setelah temuan titik 0,0)
Handler kini menolak request yang tidak lengkap:
- `origin` atau `destination` hilang → HTTP 400 `"origin and destination are required"`.
- Koordinat di luar bounding box Indonesia (lat `-11..6`, lng `95..141`) → HTTP 400
  `"coordinates must be within Indonesia bounds"`.

Tanpa validasi ini, field yang kosong ter-bind sebagai `(0,0)` dan diam-diam
melakukan routing ke koordinat 0,0 — membingungkan. Reimplementasi ada di
`internal/handler/distance_handler.go` (bind lewat `*Centroid` agar field hilang
terdeteksi sebagai `nil`, bukan nilai nol).

> **Key takeaway:** Field JSON yang hilang pada struct non-pointer akan jadi nilai
> nol (`0`), bukan error. Kalau nol itu koordinat valid secara teknis, pakai pointer
> (`*Centroid`) supaya bisa membedakan "tidak dikirim" vs "0". Lalu validasi rentang
> geografis agar input tidak masuk akal ditolak dengan pesan yang jelas.

---

## 7. Gotcha yang pernah bikin tersandung

| Gotcha | Penjelasan |
|--------|-----------|
| `pgr_nodeNetwork` lambat | O(n²) `ST_Crosses`; 129k edge > 40 menit. Skip kalau data sudah ter-noding. |
| Snap ke vertex terisolasi | Penyebab `NOT_FOUND`; saring ke giant component. |
| Token auth di level atas | `response.token`, bukan `response.data.token`. |
| Field hilang = 0 | Struct Go non-pointer → field hilang jadi nol; gunakan pointer untuk deteksi. |
| `reverse_cost` negatif | Wajib untuk one-way agar `pgr_dijkstra` (directed) menghormati arah. |
| Indeks GIST/source/target | Pastikan ada sebelum pgr_dijkstra agar query tidak full-scan. |
| `ST_LineMerge` → `MultiLineString` | Geometri rute (`ST_LineMerge(ST_Collect(w.geom))`) sering jadi `MultiLineString` (teruji: Jakarta→Bandung = 15 garis). Frontend wajib flatening rekursif, bukan asumsi `LineString` datar, dan membalik `[lng,lat]` GeoJSON → `[lat,lng]`. |

---

## 8. Saran untuk deploy ulang / skala lebih besar

1. **Full Indonesia**: aktifkan filter jalan tambahan (`residential`,
   `unclassified`, `road`) di `build_java_ways.sql`, jalankan ulang
   `build_java_ways.sql` + `build_topology.sql`, dan regenerasi
   `java_ways_components`.
2. **Performa di produksi**: `pgr_dijkstra` per-request terhadap 100k+ edge tetap
   cepat (< detik) berkat indeks source/target. Untuk trafik tinggi, pertimbangkan
   caching hasil route (misal Redis) atau precomputed distance matrix untuk
   pasangan titik yang sering.
3. **Satu komponen raksasa = terkendala**: jika target hanya Jakarta-Bandung dsb,
   partial data cukup. Komponen raksasa saat ini ~66% dari semua vertex.
4. Pastikan toleransi `pgr_createTopology` disesuaikan dengan proyeksi/ketelitian
   data (0.0001° ≈ 11 m untuk snap).

---

[← Back to Index](./README.md)