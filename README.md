# Indonesia Geomapping Service

API service geospasial untuk data wilayah administrasi Indonesia (provinsi → kabupaten → kecamatan → kelurahan/desa → kodepos), lengkap dengan reverse geocoding, pencarian nama wilayah, data pulau, dan boundary geometry (GeoJSON).

**Status:** Implemented (production-ready)  
**Data Source:** Kepmendagri No 300.2.2-2430 Tahun 2025 (via repositori [cahyadsn](https://github.com/cahyadsn))

---

## Fitur

- **Hierarki wilayah lengkap**: provinsi (38) → kabupaten (514) → kecamatan (7.285) → kelurahan/desa (83.762)
- **Reverse geocoding**: koordinat GPS (lat, lng) → hierarki wilayah paling spesifik, dengan fallback 4 level (kelurahan → kecamatan → kabupaten → provinsi)
- **Reverse geocoding massal** (batch, maks. 100 titik per request)
- **Pencarian wilayah** berbasis trigram similarity (pg_trgm) di semua level
- **Kodepos**: lookup per kodepos maupun daftar kodepos per wilayah
- **Boundary GeoJSON** per wilayah untuk visualisasi peta
- **Data pendukung**: penduduk, luas wilayah, elevasi, zona waktu, centroid, dan 17.193 pulau
- **Autentikasi JWT** (API key → token Bearer, berlaku 24 jam) + **rate limiting** per jam

## Tech Stack

| Komponen | Teknologi |
|---|---|
| Bahasa | Go 1.25 |
| HTTP Framework | Gin v1.12 |
| Database | PostgreSQL 15 + PostGIS 3.3 |
| Driver DB | pgx/v5 (pgxpool) |
| Auth | golang-jwt/jwt/v5 |
| Deployment | Docker Compose |

## Struktur Proyek

```
geomap-indonesia/
├── cmd/api/                 → entry point server (Gin + pgx)
├── internal/
│   ├── config/              → konfigurasi via environment / .env
│   ├── handler/             → layer HTTP (wilayah, reverse, kodepos, auth, health)
│   ├── service/             → layer bisnis
│   ├── repository/          → layer akses data (PostgreSQL + PostGIS)
│   ├── models/              → struct data & helper (DetectLevel, BuildReverse, ...)
│   ├── middleware/          → JWT auth + rate limiting
│   └── router/              → registrasi route
├── scripts/
│   ├── import_master/       → import data master wilayah + penduduk + luas + pulau
│   ├── import_kodepos/      → import data kodepos
│   └── import_boundaries/   → import geometry batas wilayah
├── migrations/              → schema SQL (001_schema.sql)
├── docs/                    → dokumen perencanaan & panduan penggunaan
├── data/                    → data sumber (di-clone dari repo cahyadsn)
├── docker-compose.yml       → PostgreSQL + PostGIS + Redis + API
├── Dockerfile
└── Makefile
```

Arsitektur mengikuti *clean architecture*: `handler → service → repository`, sehingga layer HTTP terpisah dari logika bisnis dan akses data.

## Skema Data

```
provinsi (kode 2 digit)
  └── kabupaten (kode 4 digit, FK provinsi_kode)
        └── kecamatan (kode 6 digit, FK kabupaten_kode)
              └── kelurahan (kode 10 digit, FK kecamatan_kode)
                    └── kodepos (FK kelurahan_kode)
pulau (standalone)
```

Kode wilayah mengikuti aturan BPS/Kemendagri: panjang kode menentukan level (`DetectLevel`). Geometry (PostGIS) tersedia untuk level provinsi, kabupaten, dan kecamatan; kelurahan memakai centroid (lat/lng).

## Quickstart

### Prasyarat

- Go 1.25+
- Docker + Docker Compose
- Data sumber di `data/` (sudah ter-clone dari repositori cahyadsn)

### 1. Jalankan database

```bash
docker compose up -d postgres redis
```

### 2. Migrasi schema

```bash
make migrate
# atau: docker compose exec -T postgres psql -U postgres -d geomapping_id -f - < migrations/001_schema.sql
```

### 3. Import data

```bash
make import-all
# atau jalankan satu per satu:
go run ./scripts/import_master      # master wilayah, penduduk, luas, pulau
go run ./scripts/import_kodepos     # kodepos
go run ./scripts/import_boundaries  # geometry batas wilayah
```

### 4. Jalankan API

```bash
go run ./cmd/api
# atau build binary:
make build && ./bin/geomap-api
```

Server berjalan di `http://localhost:8080` (default). Konfigurasi bisa diubah lewat environment variable — lihat [.env.example](.env.example).

### 5. Uji coba cepat

```bash
# Ambil token
curl -s -X POST localhost:8080/api/v1/auth \
  -H 'Content-Type: application/json' \
  -d '{"public_key":"pk_test_12345","private_key":"sk_test_67890"}'
# → {"expires_in":86400,"success":true,"token":"<JWT>","token_type":"Bearer"}

# Reverse geocode dengan token
curl -s -H "Authorization: Bearer <JWT>" \
  "localhost:8080/api/v1/reverse?lat=-6.2&lng=106.8"
```

Dokumentasi lengkap semua endpoint (parameter, contoh request/response, kode error) ada di **[docs/API_USAGE.md](docs/API_USAGE.md)**.

## Endpoint API

Semua endpoint di bawah prefix `/api/v1`. Endpoint bertanda 🔒 memerlukan header `Authorization: Bearer <JWT>`.

| Method | Endpoint | Deskripsi |
|---|---|---|
| `POST` | `/auth` | Tukar API key → JWT token (24 jam) |
| `GET` | `/health` | Health check (status DB) |
| `GET` | `/wilayah` 🔒 | Daftar provinsi |
| `GET` | `/wilayah/:kode` 🔒 | Detail wilayah (semua level) |
| `GET` | `/wilayah/:kode/children` 🔒 | Anak wilayah langsung |
| `GET` | `/reverse?lat=&lng=` 🔒 | Reverse geocode satu titik |
| `POST` | `/batch/reverse` 🔒 | Reverse geocode massal (max 100 titik) |
| `GET` | `/search?q=&tipe=&limit=` 🔒 | Cari wilayah berdasarkan nama |
| `GET` | `/kodepos/:kode` 🔒 | Lookup kodepos → wilayah |
| `GET` | `/kodepos?wilayah=` 🔒 | Daftar kodepos per wilayah |
| `GET` | `/boundaries/:kode` 🔒 | Boundary GeoJSON wilayah |

## Data yang Diimpor

| Data | Jumlah |
|---|---|
| Provinsi | 38 |
| Kabupaten | 514 |
| Kecamatan | 7.285 |
| Kelurahan | 83.762 |
| Kodepos (unik) | 10.632 |
| Pulau (dedup) | 17.193 |
| Boundary geometry (prov/kab/kec) | 7.837 |

Sumber data: [cahyadsn/wilayah](https://github.com/cahyadsn/wilayah) (master), [cahyadsn/wilayah_kodepos](https://github.com/cahyadsn/wilayah_kodepos), [cahyadsn/wilayah_pulau](https://github.com/cahyadsn/wilayah_pulau), [cahyadsn/wilayah_boundaries](https://github.com/cahyadsn/wilayah_boundaries).

Catatan: repositori boundaries tidak menyediakan geometry untuk kelurahan, sehingga reverse geocoding memakai rantai fallback dan kelurahan memakai centroid.

## Konfigurasi

Semua konfigurasi via environment variable (file `.env` di working directory otomatis dibaca). Lihat [.env.example](.env.example) untuk daftar lengkap. Yang utama:

| Variable | Default | Deskripsi |
|---|---|---|
| `SERVER_PORT` | `8080` | Port HTTP server |
| `APP_ENV` | `development` | `production` mengaktifkan release mode & validasi JWT secret |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `localhost` / `5432` / `postgres` / `secret` / `geomapping_id` | Koneksi PostgreSQL |
| `API_PUBLIC_KEY` / `API_PRIVATE_KEY` | `pk_test_12345` / `sk_test_67890` | Pasangan API key untuk `/auth` |
| `JWT_SECRET` | `change-me-in-production` | Secret penandatanganan JWT (wajib diganti di production) |
| `JWT_EXPIRES_HOURS` | `24` | Masa berlaku token |
| `RATE_LIMIT_PER_HOUR` | `100` | Batas request per jam per API key |
| `REDIS_ENABLED` | `false` | Aktifkan Redis (untuk rate limit terdistribusi) |
| `DATA_DIR` | `./data` | Lokasi data sumber untuk skrip import |

## Docker Compose

```bash
docker compose up -d          # jalankan semua service
docker compose up -d postgres redis   # hanya database
docker compose down           # hentikan
```

Service: `api` (Go, port 8080), `postgres` (PostGIS 15-3.3, port 5432), `redis` (opsional, port 6379). Gunakan `make migrate` untuk menjalankan schema setelah postgres siap.

## Makefile

```bash
make run              # go run ./cmd/api
make build            # build binary ke bin/geomap-api
make test             # go test ./...
make vet              # go vet ./...
make db-up            # docker compose up -d postgres redis
make migrate          # jalankan migrations/001_schema.sql
make import-all       # import master + kodepos + boundaries
```

## Development

```bash
go test ./...   # unit test (semua paket)
go vet ./...    # static analysis
```

## Dokumentasi Perencanaan

Dokumen perencanaan awal (goals, data sources, arsitektur, skema, dll.) tetap tersedia di folder [docs/](docs/), dengan panduan penggunaan API di [docs/API_USAGE.md](docs/API_USAGE.md).

---

**Last Updated:** 2026-08-12
