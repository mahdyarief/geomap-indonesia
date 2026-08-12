# Data Sources

**Document:** 02  
**Status:** Planning  
**Owner:** cahyadsn (Cahya Dwiana)  
**License:** MIT

---

## Overview

Data utama bersumber dari **5 repository GitHub** milik cahyadsn.  
Data official berdasarkan **Kepmendagri No 300.2.2-2430 Tahun 2025**.

Selain itu, terdapat **reference implementation** dari dhanyyudi/wilayah-id yang dapat dijadikan pembelajaran untuk arsitektur dan fitur advanced.

### Quick Reference

| Repo | Purpose | SQL Data | Size |
|------|---------|----------|------|
| [wilayah](#1-cahyadsnwilayah) | Master data wilayah | ✅ Yes | ~30 MB |
| [wilayah_boundaries](#2-cahyadsnwilayah_boundaries) | Geometry/polygon | ✅ Yes | ~500+ MB |
| [wilayah_kodepos](#3-cahyadsnwilayah_kodepos) | Postal codes | ✅ Yes | ~10 MB |
| [wilayah_logo](#4-cahyadsnwilayah_logo) | Logo/lambang | ❌ Images only | ~100+ MB |
| [wilayah_ref](#5-cahyadsnwilayah_ref) | Legal references | ❌ PDFs only | ~50+ MB |

### Reference Implementation

| Repo | Purpose | Stack | Highlights |
|------|---------|-------|------------|
| [dhanyyudi/wilayah-id](#6-dhanyyudiwilayah-id-reference) | Full stack API + Webmap | Next.js + PostGIS | 22 endpoints, Vector Tiles, OGC, MCP |

---

## 1. cahyadsn/wilayah

**URL:** https://github.com/cahyadsn/wilayah  
**Purpose:** Master data wilayah administratif Indonesia (kode, nama, hierarki)  
**Format:** SQL (MySQL/MariaDB)  
**Size:** ~30 MB  
**SQL:** ✅ Yes

### Files

| File | Size | Content | Records | SQL |
|------|------|---------|---------|-----|
| `db/wilayah.sql` | 2.8 MB | Kode + nama semua level | ~87,000 | ✅ |
| `db/wilayah_penduduk.sql` | 34 KB | Data penduduk per wilayah | ~87,000 | ✅ |
| `db/wilayah_luas.sql` | 26 KB | Luas wilayah (km²) | ~548 | ✅ |
| `db/wilayah_pulau.sql` | 1.2 MB | Data pulau Indonesia | ~17,000 | ✅ |
| `db/wilayah_level_1_2.sql` | 22.7 MB | Enhanced prov/kab data | ~548 | ✅ |

### Schema (MySQL)

```sql
-- wilayah.sql
CREATE TABLE wilayah (
    kode VARCHAR(13) PRIMARY KEY,  -- format: xx.xx.xx.xxxx
    nama VARCHAR(100) NOT NULL
);

-- wilayah_penduduk.sql
CREATE TABLE wilayah_penduduk (
    kode VARCHAR(13) PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    pria INTEGER NOT NULL,
    wanita INTEGER NOT NULL,
    total INTEGER NOT NULL
);

-- wilayah_luas.sql
CREATE TABLE wilayah_luas (
    kode VARCHAR(13) PRIMARY KEY,
    nama VARCHAR(100),
    luas DOUBLE PRECISION NOT NULL
);

-- wilayah_pulau.sql
CREATE TABLE wilayah_pulau (
    kode VARCHAR(11) PRIMARY KEY,
    nama VARCHAR(255),
    lat FLOAT,
    lng FLOAT,
    status VARCHAR(3),
    luas FLOAT,
    notes TEXT
);
```

### Data yang Didapat

- ✅ Master data: kode + nama semua level (provinsi → kelurahan/desa)
- ✅ Data penduduk: pria, wanita, total per wilayah
- ✅ Data luas wilayah: km² per provinsi/kabupaten
- ✅ Data pulau: kode, nama, koordinat, luas
- ✅ Extra metadata (level 1-2): elevasi, zona waktu, boundaries JSON
- ❌ Boundaries lengkap (hanya level 1-2, tidak untuk kec/kel)

### Digunakan Untuk

- Source of truth untuk master data (kode, nama, hierarchy)
- Data penduduk & luas wilayah
- Data pulau Indonesia
- Extra metadata (elevasi, zona waktu)

---

## 2. cahyadsn/wilayah_boundaries

**URL:** https://github.com/cahyadsn/wilayah_boundaries  
**Purpose:** Boundary/polygon data untuk SEMUA level wilayah  
**Format:** SQL (MySQL/MariaDB)  
**Size:** ~500+ MB  
**SQL:** ✅ Yes

### Structure

```
db/
├── ddl_wilayah_boundaries.sql     -- Schema definition
├── prov/                           -- 34 files (1 per provinsi)
│   ├── wilayah_boundaries_prov_11.sql
│   ├── wilayah_boundaries_prov_12.sql
│   └── ...
├── kab/                            -- 34 files (1 per provinsi)
│   ├── wilayah_boundaries_kab_11.sql
│   └── ...
├── kec/                            -- 34 files (1 per provinsi)
│   ├── wilayah_boundaries_kec_11.sql
│   └── ...
└── kel/                            -- ~500 files (1 per kabupaten)
    ├── 11/
    │   ├── wilayah_boundaries_kel_11.01.sql
    │   └── ...
    └── ...
```

### Schema (MySQL)

```sql
CREATE TABLE wilayah_boundaries (
    kode VARCHAR(13) PRIMARY KEY,
    nama VARCHAR(100),
    lat DOUBLE PRECISION,    -- centroid latitude
    lng DOUBLE PRECISION,    -- centroid longitude
    path TEXT,               -- JSON string of coordinates
    status INT2 DEFAULT 1
);
```

### Path Format

```
'[[[lat,lng],[lat,lng],...]]'  -- JSON string

Example:
'[[[5.276431,96.848914],[5.275096,96.861011],...]]'
```

### Coverage

| Level | Records | Files | Status |
|-------|---------|-------|--------|
| Provinsi | 34 | 34 | ✅ Complete |
| Kab/Kota | 514 | 34 | ✅ Complete |
| Kecamatan | ~7,000 | 34 | ✅ Complete |
| Kel/Desa | ~80,000 | ~500 | ✅ Complete |

### Data yang Didapat

- ✅ Geometry/polygon data untuk SEMUA level
- ✅ Centroid coordinates (lat, lng)
- ✅ Multipolygon (simplified)
- ❌ High resolution (data sudah di-simplify)

### Digunakan Untuk

- PostGIS geometry (setelah conversion)
- Spatial queries (ST_Contains, ST_Intersects)
- Reverse geocoding (point-in-polygon)
- Boundary visualization (GeoJSON)

### Catatan Penting

- Path format: JSON string dengan `[lat,lng]`
- PostGIS requires: `[lng,lat]` (reverse order)
- Perlu transformation saat import
- Data sudah simplified, untuk high-res perlu repo terpisah

---

## 3. cahyadsn/wilayah_kodepos

**URL:** https://github.com/cahyadsn/wilayah_kodepos  
**Purpose:** Mapping kodepos ke wilayah administratif  
**Format:** SQL (MySQL/MariaDB)  
**Size:** ~10 MB  
**SQL:** ✅ Yes

### Files

| File | Size | Content | Records | SQL |
|------|------|---------|---------|-----|
| `db/wilayah_kodepos.sql` | ~10 MB | Kodepos vs kode wilayah | 83,762 | ✅ |
| `json/wilayah_kodepos.json` | - | JSON format | 83,762 | ❌ |
| `src/pos-data.csv` | - | Source CSV | - | ❌ |

### Schema (MySQL)

```sql
CREATE TABLE wilayah_kodepos (
    kode VARCHAR(13) NOT NULL,  -- format: xx.xx.xx.xxxx
    kodepos VARCHAR(5) DEFAULT NULL,
    PRIMARY KEY (kode)
);
```

### Coverage

- 83,762 desa/kelurahan dengan kodepos
- Format kode: `xx.xx.xx.xxxx` (dengan dot)
- Perlu convert ke format tanpa dot

### Data yang Didapat

- ✅ Mapping kodepos → wilayah
- ✅ Reverse lookup (wilayah → kodepos)
- ❌ Data alamat lengkap
- ❌ Geocoding data

### Digunakan Untuk

- Postal code lookup
- Address validation
- Reverse lookup (kodepos → wilayah)

---

## 4. cahyadsn/wilayah_logo

**URL:** https://github.com/cahyadsn/wilayah_logo  
**Purpose:** Logo/lambang daerah (provinsi & kabupaten/kota)  
**Format:** PNG images  
**Size:** ~100+ MB  
**SQL:** ❌ No (images only)

### Structure

```
prov/
├── img/              -- 34 provinsi logos (full size)
│   ├── 11.png        -- Aceh
│   ├── 12.png        -- Sumatera Utara
│   └── ...
└── thumbs/           -- 34 provinsi logos (thumbnail)
    ├── 11_thumbs.png
    └── ...

kab/
├── 11/               -- Aceh
│   ├── img/          -- Full size logos
│   │   ├── 11.01.png -- Aceh Selatan
│   │   └── ...
│   └── thumbs/       -- Thumbnails
│       └── ...
├── 12/               -- Sumatera Utara
│   └── ...
└── ...
```

### File Naming

```
Provinsi: [kode].png
  Example: 13.png (Sumatera Barat)
  Example: 31.png (DKI Jakarta)

Kab/Kota: [kode].png
  Example: 11.01.png (Aceh Selatan)
  Example: 31.71.png (Jakarta Pusat)

Thumbnail: [kode]_thumbs.png
```

### Coverage

- ✅ 34 provinsi logos
- ✅ 514 kabupaten/kota logos
- ❌ Kecamatan logos (tidak ada)
- ❌ Kelurahan logos (tidak ada)

### Data yang Didapat

- ✅ Logo/lambang resmi daerah
- ✅ Full size (height 1200px)
- ✅ Thumbnail (height 100px)
- ✅ PNG format

### Digunakan Untuk

- Visual representation di UI
- Branding/identification
- Profile daerah

### Storage Options

- Local filesystem (simple)
- MinIO (self-hosted S3)
- Cloud storage (AWS S3, GCS)

---

## 5. cahyadsn/wilayah_ref

**URL:** https://github.com/cahyadsn/wilayah_ref  
**Purpose:** Dokumen referensi legal (Kepmendagri, Permendagri)  
**Format:** PDF documents  
**Size:** ~50+ MB  
**SQL:** ❌ No (PDFs only)

### Content

| Document | Year | File |
|----------|------|------|
| Kepmendagri No 300.2.2-2430 | 2025 | ✅ Latest |
| Kepmendagri No 300.2.2-2138 | 2025 | ✅ |
| Kepmendagri No 100.1.1-6117 | 2022 | ✅ Archive |
| Kepmendagri No 050-145 | 2022 | ✅ Archive |
| Permendagri No 58 | 2021 | ✅ Archive |
| Permendagri No 72 | 2019 | ✅ Archive |
| Historical regulations | 2015-2017 | ✅ Archive |

### Data yang Didapat

- ✅ Legal documents (PDF)
- ✅ Historical data
- ✅ Lampiran (attachments)
- ❌ Machine-readable data

### Digunakan Untuk

- Reference untuk data accuracy
- Audit trail
- Legal compliance
- Version tracking

---

## Summary: SQL Data Sources

### Repo dengan SQL (Bisa Import Langsung)

| Repo | SQL Files | Tables | Records |
|------|-----------|--------|---------|
| **wilayah** | 5 files | wilayah, wilayah_penduduk, wilayah_luas, wilayah_pulau, wilayah_level_1_2 | ~190,000 |
| **wilayah_boundaries** | ~600 files | wilayah_boundaries (per level) | ~87,000 |
| **wilayah_kodepos** | 1 file | wilayah_kodepos | 83,762 |

### Repo tanpa SQL (Asset/Reference Only)

| Repo | Type | Content |
|------|------|---------|
| **wilayah_logo** | Images | PNG logos |
| **wilayah_ref** | Documents | PDF references |

---

## Data Mapping ke Database Kita

```
Source                              → Target Table
─────────────────────────────────────────────────────────
wilayah.sql                         → provinsi, kabupaten, kecamatan, kelurahan
                                    → (split by level based on kode length)

wilayah_penduduk.sql                → provinsi.penduduk_*, kabupaten.penduduk_*

wilayah_luas.sql                    → provinsi.luas, kabupaten.luas

wilayah_pulau.sql                   → pulau

wilayah_level_1_2.sql               → provinsi/kabupaten (extra metadata)
                                    → elevasi, zona_waktu

wilayah_boundaries/prov/*.sql       → provinsi.geometry
wilayah_boundaries/kab/*.sql        → kabupaten.geometry
wilayah_boundaries/kec/*.sql        → kecamatan.geometry
wilayah_boundaries/kel/**/*.sql     → kelurahan.geometry

wilayah_kodepos.sql                 → kodepos

wilayah_logo/**                     → Static storage (logo_url in DB)

wilayah_ref/**                      → Reference only (tidak import)
```

---

## Update Mechanism

### Trigger Update

Update dilakukan ketika:
1. **Kepmendagri baru dirilis** (perubahan data wilayah)
2. **Boundary changes** (perubahan batas wilayah)
3. **New wilayah added** (pemekaran wilayah)
4. **Data corrections** (koreksi data)
5. **Periodic check** (rutin cek update)

### Monitoring Updates

#### Option 1: GitHub Watch
```
1. Watch semua repo cahyadsn
2. Receive email notifications
3. Check commit history
```

#### Option 2: RSS Feed
```
https://github.com/cahyadsn/wilayah/commits/master.atom
https://github.com/cahyadsn/wilayah_boundaries/commits/master.atom
https://github.com/cahyadsn/wilayah_kodepos/commits/master.atom
```

#### Option 3: Automated Script
```bash
#!/bin/bash
# check-updates.sh

REPOS=(
  "cahyadsn/wilayah"
  "cahyadsn/wilayah_boundaries"
  "cahyadsn/wilayah_kodepos"
)

for repo in "${REPOS[@]}"; do
  # Get latest commit date
  latest=$(curl -s "https://api.github.com/repos/$repo/commits" | \
           jq -r '.[0].commit.committer.date')
  
  # Compare with last check
  if [ "$latest" > "$last_check" ]; then
    echo "Update available: $repo"
    # Trigger notification or import
  fi
done
```

### Update Process

```
┌─────────────────────────────────────────┐
│ 1. DETECT UPDATE                        │
│    - Check GitHub notifications         │
│    - Review changelog                   │
│    - Identify changed files             │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│ 2. DOWNLOAD NEW DATA                    │
│    - git pull or download ZIP           │
│    - Backup current data                │
│    - Extract new files                  │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│ 3. VALIDATE DATA                        │
│    - Check file structure               │
│    - Verify record counts               │
│    - Compare with previous version      │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│ 4. TEST IMPORT                          │
│    - Import to staging database         │
│    - Run validation queries             │
│    - Test spatial queries               │
│    - Verify API responses               │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│ 5. DEPLOY TO PRODUCTION                 │
│    - Schedule maintenance window        │
│    - Backup production database         │
│    - Run import scripts                 │
│    - Validate production data           │
│    - Monitor for issues                 │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│ 6. DOCUMENTATION                        │
│    - Update changelog                   │
│    - Record Kepmendagri version         │
│    - Archive old data (optional)        │
└─────────────────────────────────────────┘
```

### Update Scripts

#### Full Re-import
```bash
#!/bin/bash
# full-update.sh

echo "=== Starting Full Data Update ==="

# 1. Backup current database
pg_dump geomapping_id > backup_$(date +%Y%m%d).sql

# 2. Drop and recreate schema
psql geomapping_id -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
psql geomapping_id -f migrations/001_schema.sql

# 3. Import all data
./scripts/import_master.go
./scripts/import_boundaries.go
./scripts/import_kodepos.go

# 4. Validate
./scripts/validate_data.go

echo "=== Update Complete ==="
```

#### Incremental Update
```bash
#!/bin/bash
# incremental-update.sh

echo "=== Starting Incremental Update ==="

# 1. Backup
pg_dump geomapping_id > backup_$(date +%Y%m%d).sql

# 2. Import only changed files
if [ -f "new_data/wilayah.sql" ]; then
  ./scripts/import_master.go --update
fi

if [ -d "new_data/boundaries" ]; then
  ./scripts/import_boundaries.go --update
fi

# 3. Validate
./scripts/validate_data.go

echo "=== Update Complete ==="
```

### Version Tracking

Track data version di database:

```sql
CREATE TABLE data_version (
    id SERIAL PRIMARY KEY,
    repo VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    kepmendagri VARCHAR(100),
    imported_at TIMESTAMP DEFAULT NOW(),
    notes TEXT
);

-- Example:
INSERT INTO data_version (repo, version, kepmendagri, notes)
VALUES 
  ('wilayah', '2025.07', '300.2.2-2430/2025', 'Initial import'),
  ('wilayah_boundaries', '2025.07', '300.2.2-2430/2025', 'Initial import'),
  ('wilayah_kodepos', '2025.07', '300.2.2-2430/2025', 'Initial import');
```

### Rollback Plan

Jika update bermasalah:

```bash
# 1. Stop API service
systemctl stop geomapping-api

# 2. Restore from backup
psql geomapping_id < backup_20260101.sql

# 3. Start API service
systemctl start geomapping-api

# 4. Verify
curl http://localhost:8080/health
```

---

## Data Quality Checks

### Validation Rules

1. **Record Counts:**
   ```sql
   -- Expected counts
   provinsi:   34
   kabupaten:  514
   kecamatan:  ~7,000
   kelurahan:  ~80,000
   kodepos:    ~83,762
   pulau:      ~17,000
   ```

2. **Hierarchy Integrity:**
   ```sql
   -- No orphaned records
   SELECT COUNT(*) FROM kabupaten kb
   LEFT JOIN provinsi p ON kb.provinsi_kode = p.kode
   WHERE p.kode IS NULL;  -- Should be 0
   ```

3. **Geometry Validity:**
   ```sql
   -- All geometries valid
   SELECT COUNT(*) FROM provinsi 
   WHERE NOT ST_IsValid(geometry);  -- Should be 0
   ```

4. **Coordinate Range:**
   ```sql
   -- Indonesia bounds
   -- Latitude: -11 to 6
   -- Longitude: 95 to 141
   ```

---

## 6. dhanyyudi/wilayah-id (Reference)

**URL:** https://github.com/dhanyyudi/wilayah-id  
**Purpose:** Full-stack reference implementation (API + Webmap + OGC + MCP)  
**Format:** Next.js + PostGIS + Vector Tiles  
**Data Source:** Dukcapil Kemendagri 2024  
**Status:** ⭐ Reference only (bukan data source utama)

### Overview

Repository ini adalah **reference implementation** paling lengkap untuk geomapping Indonesia. Bukan data source utama kita (kita pakai data cahyadsn), tapi sangat berguna sebagai:
- ✅ Pembelajaran arsitektur & best practices
- ✅ Referensi fitur advanced yang bisa diadopsi
- ✅ Perbandingan implementasi

### Tech Stack

| Component | Technology |
|-----------|-----------|
| Framework | Next.js 16 |
| Database | PostgreSQL + PostGIS 3.5 |
| Map Library | MapLibre GL JS v5 |
| Vector Tiles | Tippecanoe (MVT .pbf) |
| MCP Server | FastMCP (Python) |
| Deployment | Docker + Cloudflare |

### Features (22 Endpoints)

#### Regions API
```
GET /api/v1/regions/provinces          → List 38 provinsi
GET /api/v1/regions/provinces/:kode    → Detail provinsi
GET /api/v1/regions/regencies          → List kabupaten (filter by province)
GET /api/v1/regions/regencies/:kode    → Detail kabupaten
GET /api/v1/regions/districts          → List kecamatan (filter by regency)
GET /api/v1/regions/districts/:kode    → Detail kecamatan
GET /api/v1/regions/villages           → List desa (filter by district)
GET /api/v1/regions/villages/:kode     → Detail desa + hierarchy
GET /api/v1/regions/search             → Multi-level search
```

#### Boundaries API (GeoJSON)
```
GET /api/v1/boundaries/provinces       → List provinsi + geometry
GET /api/v1/boundaries/provinces/:kode → Single provinsi + geometry
GET /api/v1/boundaries/regencies       → List kabupaten + geometry
GET /api/v1/boundaries/regencies/:kode → Single kabupaten + geometry
GET /api/v1/boundaries/districts       → List kecamatan + geometry
GET /api/v1/boundaries/districts/:kode → Single kecamatan + geometry
GET /api/v1/boundaries/villages        → List desa + geometry
GET /api/v1/boundaries/villages/:kode  → Single desa + geometry
```

#### Reverse Geocoding
```
GET /api/v1/boundaries/reverse?lat=&lng=&level= → Coordinate → wilayah
```

#### Postal Codes
```
GET /api/v1/postal-codes               → Query by village/postal/district
GET /api/v1/postal-codes/lookup        → Prefix lookup
```

#### OGC Services (Advanced)
```
GET /api/v1/ogc/wms                    → WMS 1.3.0 (GetCapabilities, GetMap, GetFeatureInfo)
GET /api/v1/ogc/wfs                    → WFS 2.0 (GetFeature, DescribeFeatureType)
```

#### Vector Tiles
```
GET /tiles/{layer}/{z}/{x}/{y}.pbf     → MVT tiles

Layers:
- provinsi (z3-9, 38 features)
- kabupaten (z7-11, 514 features)
- kecamatan (z10-12, 7,285 features)
- desa (z12-14, 83,762 features)
```

### Key Features untuk Dipelajari

#### 1. OGC Compliance (WMS/WFS)
```xml
<!-- WMS GetCapabilities -->
<WMS_Capabilities version="1.3.0">
  <Service>
    <Name>WMS</Name>
    <Title>Wilayah Indonesia</Title>
  </Service>
  <Capability>
    <Layer>
      <Name>provinsi</Name>
      <Title>Batas Provinsi</Title>
    </Layer>
  </Capability>
</WMS_Capabilities>
```

**Use Case:** Integrasi dengan GIS software (QGIS, ArcGIS)

#### 2. Vector Tiles (MVT)
```python
# Generate tiles dengan Tippecanoe
python etl/generate_tiles.py

# Output: /tiles/{layer}/{z}/{x}/{y}.pbf
# Served statik via Cloudflare/NGINX
```

**Benefits:**
- ✅ Fast rendering di map
- ✅ Client-side styling
- ✅ Reduced server load

#### 3. MCP Server (AI Integration)
```python
# Model Context Protocol untuk AI agents
# Tools:
- describe_spatial_service
- resolve_spatial_entity
- query_spatial_data
- reverse_geocode
```

**Use Case:** Integrasi dengan Claude, Cursor, AI agents

### Data Comparison

| Aspect | cahyadsn | dhanyyudi/wilayah-id |
|--------|----------|---------------------|
| **Data Source** | Kepmendagri 2025 | Dukcapil 2024 |
| **Provinsi** | 34 | 38 (includes new Papua provinces) |
| **Kabupaten** | 514 | 514 |
| **Kecamatan** | ~7,000 | 7,285 |
| **Desa** | ~80,000 | 83,762 |
| **Geometry** | MultiPolygon (simplified) | MultiPolygon (full) |
| **Format** | SQL (MySQL) | PostGIS + GeoJSON |

### What We Can Learn

#### ✅ Good Practices
1. **Separate regions & boundaries endpoints**
   - `/regions/*` untuk metadata
   - `/boundaries/*` untuk geometry
   - Cleaner API design

2. **OGC Compliance**
   - WMS/WFS untuk GIS integration
   - Standard protocol
   - Wide compatibility

3. **Vector Tiles**
   - Performance optimization
   - Client-side rendering
   - Reduced bandwidth

4. **MCP Server**
   - AI-ready architecture
   - Future-proof design
   - Easy integration

5. **Comprehensive Documentation**
   - OpenAPI specs
   - Interactive docs
   - Examples & tutorials

#### ⚠️ Considerations
1. **Data Source Difference**
   - Mereka pakai Dukcapil 2024 (38 provinsi)
   - Kita pakai Kepmendagri 2025 (34 provinsi)
   - Perlu clarify mana yang lebih update

2. **Tech Stack Complexity**
   - Next.js + PostGIS + Vector Tiles + MCP
   - More complex daripada Golang + PostGIS
   - Trade-off: features vs simplicity

3. **Performance**
   - Vector tiles require preprocessing
   - OGC services add complexity
   - MCP server additional maintenance

### Features to Consider Adopting

#### Phase 1 (MVP)
- ✅ Separate regions/boundaries endpoints
- ✅ Comprehensive API documentation

#### Phase 2 (Enhanced)
- ⏳ Vector tiles untuk map performance
- ⏳ OGC WMS/WFS untuk GIS integration

#### Phase 3 (Advanced)
- ⏳ MCP server untuk AI integration
- ⏳ Interactive webmap

### Implementation Notes

**Jika ingin adopsi fitur:**

1. **Vector Tiles:**
   ```bash
   # Generate dengan Tippecanoe
   tippecanoe -o tiles.mbtiles \
     -z14 -Z0 \
     --drop-densest-as-needed \
     input.geojson
   
   # Serve via NGINX/Cloudflare
   ```

2. **OGC WMS:**
   ```go
   // Implement WMS handler
   func handleWMS(c *gin.Context) {
       request := c.Query("REQUEST")
       switch request {
       case "GetCapabilities":
           // Return XML capabilities
       case "GetMap":
           // Render map image
       }
   }
   ```

3. **MCP Server:**
   ```python
   # Python FastMCP server
   from fastmcp import FastMCP
   
   mcp = FastMCP("wilayah")
   
   @mcp.tool()
   def reverse_geocode(lat: float, lng: float) -> dict:
       # Query PostGIS
       return result
   ```

### Links

- **Repository:** https://github.com/dhanyyudi/wilayah-id
- **Demo:** https://wilayah-id.vercel.app
- **API Docs:** https://wilayah-id.vercel.app/docs
- **Data Source:** [Ditjen Dukcapil Kemendagri](https://gis.dukcapil.kemendagri.go.id/peta/)

---

## External References

### Official Sources
- **Kepmendagri:** https://ditjenbinaadwil.kemendagri.go.id/
- **Data Wilayah:** https://github.com/cahyadsn/wilayah
- **Data Boundaries:** https://github.com/cahyadsn/wilayah_boundaries
- **Data Kodepos:** https://github.com/cahyadsn/wilayah_kodepos
- **Data Logo:** https://github.com/cahyadsn/wilayah_logo
- **Data Referensi:** https://github.com/cahyadsn/wilayah_ref

### Reference Implementations
- **wilayah-id:** https://github.com/dhanyyudi/wilayah-id (Full stack: API + Webmap + OGC + MCP)

### Contact
- **Author (cahyadsn):** Cahya Dwiana
- **Email:** cahyadsn@gmail.com
- **GitHub:** https://github.com/cahyadsn

---

[← Previous: Goals & Scope](./01-goals-scope.md) | [Back to Index](./README.md) | [Next: Architecture →](./03-architecture.md)
