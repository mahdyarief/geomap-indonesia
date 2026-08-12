# Panduan Penggunaan API

Dokumentasi lengkap API service **Indonesia Geomapping Service**. Semua endpoint berada di bawah prefix `/api/v1`.

## Daftar Isi

1. [Base URL & Autentikasi](#base-url--autentikasi)
2. [Format Respons](#format-respons)
3. [Kode Error](#kode-error)
4. [Endpoint](#endpoint)
   - [POST /auth](#post-auth)
   - [GET /health](#get-health)
   - [GET /wilayah](#get-wilayah)
   - [GET /wilayah/:kode](#get-wilayahkode)
   - [GET /wilayah/:kode/children](#get-wilayahkodechildren)
   - [GET /reverse](#get-reverse)
   - [POST /batch/reverse](#post-batchreverse)
   - [GET /search](#get-search)
   - [GET /kodepos/:kode](#get-kodeposkode)
   - [GET /kodepos?wilayah=](#get-kodeposwilayah)
   - [GET /boundaries/:kode](#get-boundarieskode)
5. [Rate Limiting](#rate-limiting)
6. [Contoh Integrasi (JavaScript)](#contoh-integrasi-javascript)

---

## Base URL & Autentikasi

**Base URL lokal:** `http://localhost:8080/api/v1`

Autentikasi memakai **JWT Bearer token** dengan alur dua langkah:

1. **Tukar API key → JWT**: kirim `public_key` dan `private_key` ke `POST /auth`. Dalam mode default, gunakan `pk_test_12345` dan `sk_test_67890`.
2. **Gunakan token**: sertakan header `Authorization: Bearer <TOKEN>` pada semua endpoint selain `/auth` dan `/health`.

Token berlaku **24 jam** (default, bisa diubah via `JWT_EXPIRES_HOURS`). Tanpa token, server merespons `401` dengan kode `UNAUTHORIZED`.

---

## Format Respons

### Sukses

```json
{
  "success": true,
  "data": { ... }
}
```

Pengecualian: `/auth` dan `/health` tidak membungkus payload dalam `data`.

### Gagal

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "wilayah not found"
  }
}
```

---

## Kode Error

| HTTP | Code | Deskripsi |
|---|---|---|
| 400 | `BAD_REQUEST` | Parameter/body tidak valid |
| 401 | `UNAUTHORIZED` | Token tidak ada / tidak valid / kedaluwarsa |
| 404 | `NOT_FOUND` | Data tidak ditemukan |
| 429 | `RATE_LIMITED` | Melebihi batas request per jam |
| 500 | `INTERNAL_ERROR` | Kesalahan server |

---

## Endpoint

### POST /auth

Tukar API key menjadi JWT token. **Tidak butuh token.**

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth \
  -H 'Content-Type: application/json' \
  -d '{"public_key":"pk_test_12345","private_key":"sk_test_67890"}'
```

**Response 200:**
```json
{
  "expires_in": 86400,
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

`expires_in` dalam detik (86400 = 24 jam). Token dipakai sebagai `Authorization: Bearer <token>`.

**Response 401** (API key salah):
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid credentials"
  }
}
```

---

### GET /health

Cek status server dan koneksi database. **Tidak butuh token.**

**Request:**
```bash
curl http://localhost:8080/api/v1/health
```

**Response 200:**
```json
{
  "database": "up",
  "status": "ok",
  "success": true,
  "time": "2026-08-12T10:52:49Z"
}
```

---

### GET /wilayah

Daftar semua provinsi. 🔒

**Request:**
```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/wilayah
```

**Response 200:**
```json
{
  "success": true,
  "data": [
    { "kode": "11", "nama": "Aceh", "type": "provinsi" },
    { "kode": "12", "nama": "Sumatera Utara", "type": "provinsi" },
    { "kode": "13", "nama": "Sumatera Barat", "type": "provinsi" }
  ]
}
```

---

### GET /wilayah/:kode

Detail wilayah di semua level — provinsi (2 digit), kabupaten (4 digit), kecamatan (6 digit), kelurahan (10 digit). 🔒

**Request:**
```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/wilayah/12
```

**Response 200 (provinsi):**
```json
{
  "success": true,
  "data": {
    "kode": "12",
    "nama": "Sumatera Utara",
    "type": "provinsi",
    "centroid": { "lat": 2.188438, "lng": 99.0580572 },
    "luas": 72437.76,
    "penduduk": { "total": 15640905, "pria": 7828098, "wanita": 7812807 },
    "zona_waktu": "WIB",
    "elevasi": 32
  }
}
```

Untuk level kabupaten/kecamatan, `parent` ditambahkan:
```json
{
  "success": true,
  "data": {
    "kode": "1201",
    "nama": "Kabupaten Tapanuli Tengah",
    "type": "kabupaten",
    "parent": { "kode": "12", "nama": "Sumatera Utara" },
    "centroid": { "lat": 2.0220178, "lng": 98.4100147 }
  }
}
```

**Response 404:**
```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "wilayah not found" }
}
```

---

### GET /wilayah/:kode/children

Anak wilayah langsung dari satu wilayah (provinsi → kabupaten, kabupaten → kecamatan, kecamatan → kelurahan). 🔒

**Request:**
```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/wilayah/12/children
```

**Response 200:**
```json
{
  "success": true,
  "data": [
    { "kode": "1201", "nama": "Kabupaten Tapanuli Tengah", "type": "kabupaten" },
    { "kode": "1202", "nama": "Kabupaten Tapanuli Utara", "type": "kabupaten" }
  ]
}
```

---

### GET /reverse

Reverse geocoding: ubah koordinat GPS menjadi hierarki wilayah paling spesifik. Parameter query: `lat` dan `lng`. 🔒

Jika titik berada di dalam geometry kelurahan, seluruh hierarki terisi. Karena sumber data tidak menyediakan geometry kelurahan, endpoint ini memakai **fallback berantai**: kelurahan → kecamatan → kabupaten → provinsi, sehingga hasil paling spesifik yang tersedia yang dikembalikan.

**Request:**
```bash
curl -H "Authorization: Bearer <TOKEN>" \
  "http://localhost:8080/api/v1/reverse?lat=2.022&lng=98.410"
```

**Response 200 (tingkat kecamatan):**
```json
{
  "success": true,
  "data": {
    "input": { "lat": 2.022, "lng": 98.41 },
    "provinsi": { "kode": "12", "nama": "Sumatera Utara" },
    "kabupaten": { "kode": "1201", "nama": "Kabupaten Tapanuli Tengah" },
    "kecamatan": { "kode": "120101", "nama": "Barus" },
    "kelurahan": { "kode": "", "nama": "" },
    "kodepos": null,
    "centroid": { "lat": 2.0220178, "lng": 98.4100147 }
  }
}
```

Catatan: `kelurahan` kosong dan `kodepos` `null` saat hasil fallback berada di level kecamatan atau di atasnya (kodepos hanya tersedia di level kelurahan).

**Response 404** (koordinat di luar wilayah Indonesia):
```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "no wilayah found for coordinates" }
}
```

---

### POST /batch/reverse

Reverse geocoding massal — maksimal **100 titik** per request. 🔒

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/batch/reverse \
  -H "Authorization: Bearer <TOKEN>" \
  -H 'Content-Type: application/json' \
  -d '{"points":[{"lat":2.022,"lng":98.410},{"lat":-6.2,"lng":106.8}]}'
```

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "input": { "lat": 2.022, "lng": 98.41 },
      "result": {
        "input": { "lat": 2.022, "lng": 98.41 },
        "provinsi": { "kode": "12", "nama": "Sumatera Utara" },
        "kabupaten": { "kode": "1201", "nama": "Kabupaten Tapanuli Tengah" },
        "kecamatan": { "kode": "120101", "nama": "Barus" },
        "kelurahan": { "kode": "", "nama": "" },
        "kodepos": null,
        "centroid": { "lat": 2.0220178, "lng": 98.4100147 }
      }
    },
    {
      "input": { "lat": -6.2, "lng": 106.8 },
      "result": {
        "input": { "lat": -6.2, "lng": 106.8 },
        "provinsi": { "kode": "31", "nama": "DKI Jakarta" },
        "kabupaten": { "kode": "3175", "nama": "Kota Jakarta Timur" },
        "kecamatan": { "kode": "317503", "nama": "Jatinegara" },
        "kelurahan": { "kode": "", "nama": "" },
        "kodepos": null,
        "centroid": { "lat": -6.2222, "lng": 106.8565 }
      }
    }
  ]
}
```

Titik yang tidak ditemukan tetap dikembalikan dalam array dengan `result` berisi `null` (atau objek error sesuai implementasi), dan request tetap sukses.

---

### GET /search

Cari wilayah berdasarkan nama di semua level. 🔒

**Parameter query:**
| Parameter | Wajib | Deskripsi |
|---|---|---|
| `q` | Ya | Kata kunci nama wilayah (case-insensitive) |
| `tipe` | Tidak | Filter level: `provinsi`, `kabupaten`, `kecamatan`, `kelurahan` |
| `limit` | Tidak | Maksimal hasil per level (default 10) |

Pencarian memakai trigram similarity (`pg_trgm`) sehingga hasil diurutkan berdasarkan kemiripan.

**Request:**
```bash
curl -H "Authorization: Bearer <TOKEN>" \
  "http://localhost:8080/api/v1/search?q=barus&limit=3"
```

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "kode": "1207082007",
      "nama": "Lau Barus Baru",
      "type": "kelurahan",
      "parent": { "kode": "120708", "nama": "Sinembah Tanjung Muda Hilir" },
      "province": "Sumatera Utara"
    },
    {
      "kode": "120101",
      "nama": "Barus",
      "type": "kecamatan",
      "parent": { "kode": "1201", "nama": "Kabupaten Tapanuli Tengah" },
      "province": "Sumatera Utara"
    }
  ]
}
```

Untuk hasil bertipe `provinsi`, `parent` dan `province` tidak ada.

---

### GET /kodepos/:kode

Lookup kodepos: kembalikan kelurahan + hierarki wilayah tempat kodepos berada. 🔒

**Request:**
```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/kodepos/22564
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "kodepos": "22564",
    "wilayah": {
      "kode": "1201011001",
      "nama": "Pasar Batu Gerigis",
      "kecamatan": "Barus",
      "kabupaten": "Kabupaten Tapanuli Tengah",
      "provinsi": "Sumatera Utara"
    }
  }
}
```

**Response 404:**
```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "kodepos not found" }
}
```

---

### GET /kodepos?wilayah=

Daftar semua kodepos dalam satu wilayah. Kode wilayah bisa di level mana pun — provinsi (2 digit), kabupaten (4 digit), kecamatan (6 digit), atau kelurahan (10 digit). 🔒

**Request:**
```bash
curl -H "Authorization: Bearer <TOKEN>" \
  "http://localhost:8080/api/v1/kodepos?wilayah=120101"
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "kodepos": ["22564"]
  }
}
```

**Response 404** (kode tidak dikenal atau tidak punya kodepos):
```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "kodepos not found" }
}
```

---

### GET /boundaries/:kode

Boundary geometry (GeoJSON) untuk satu wilayah — tersedia untuk level provinsi, kabupaten, dan kecamatan. 🔒

**Request:**
```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/boundaries/12
```

**Response 200** (GeoJSON Feature):
```json
{
  "type": "Feature",
  "properties": {
    "kode": "12",
    "nama": "Sumatera Utara",
    "type": "provinsi"
  },
  "geometry": {
    "type": "MultiPolygon",
    "coordinates": [ [ [ [98.516294267, -0.638775138], [98.489492343, -0.61898774] ] ] ]
  }
}
```

Catatan: geometry tidak tersedia untuk kelurahan (sumber data cahyadsn/wilayah_boundaries hanya menyediakan provinsi, kabupaten, kecamatan). Response berformat GeoJSON standar sehingga langsung bisa dirender oleh Leaflet/MapLibre/Google Maps.

---

## Rate Limiting

- Default: **100 request per jam per API key** (bisa diubah via `RATE_LIMIT_PER_HOUR`).
- Saat batas tercapai, server merespons `429`:
```json
{
  "success": false,
  "error": { "code": "RATE_LIMITED", "message": "rate limit exceeded" }
}
```
- Rate limiter disimpan di memori; untuk deployment multi-instance bisa diaktifkan Redis (`REDIS_ENABLED=true`).

---

## Contoh Integrasi (JavaScript)

```javascript
const BASE = "http://localhost:8080/api/v1";

async function getToken() {
  const res = await fetch(`${BASE}/auth`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      public_key: "pk_test_12345",
      private_key: "sk_test_67890",
    }),
  });
  const json = await res.json();
  return json.token;
}

async function reverseGeocode(lat, lng) {
  const token = await getToken();
  const res = await fetch(`${BASE}/reverse?lat=${lat}&lng=${lng}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.json();
}

reverseGeocode(-6.2, 106.8).then((json) => {
  const d = json.data;
  console.log(
    `${d.provinsi.nama} > ${d.kabupaten.nama} > ${d.kecamatan.nama} > ${d.kelurahan.nama}`
  );
});
```

---

**Terakhir diperbarui:** 2026-08-12
