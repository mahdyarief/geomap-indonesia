# Data Import Strategy

**Document:** 05  
**Status:** Planning

---

## Overview

Import data dari 4 source repos → transform → unified PostgreSQL database.

---

## Import Order

```
1. Setup database + PostGIS extension
2. Create schema (tables, indexes)
3. Import master data (wilayah.sql)
4. Import metadata (penduduk, luas, pulau)
5. Import boundaries (geometry)
6. Import postal codes
7. Setup logos (static files)
8. Validate data
```

---

## Step 1: Setup Database

```bash
# Create database
createdb geomapping_id

# Enable PostGIS
psql geomapping_id -c "CREATE EXTENSION postgis;"
```

---

## Step 2: Create Schema

```bash
# Run migration
psql geomapping_id -f migrations/001_schema.sql
```

---

## Step 3: Import Master Data

**Source:** `wilayah/db/wilayah.sql`

**Transformation:**
```
Split flat table by level:
- 2 digits  → provinsi
- 4 digits  → kabupaten
- 6 digits  → kecamatan
- 10 digits → kelurahan
```

**Import Script Logic:**
```go
func importMasterData() {
    // Read wilayah.sql
    // Parse INSERT statements
    // Split by level
    // Insert to respective tables
    
    for _, row := range rows {
        kode := cleanKode(row.kode)  // remove dots
        level := detectLevel(kode)
        
        switch level {
        case "provinsi":
            insertProvinsi(kode, row.nama)
        case "kabupaten":
            insertKabupaten(kode, row.nama, extractProvinsiKode(kode))
        case "kecamatan":
            insertKecamatan(kode, row.nama, extractKabupatenKode(kode))
        case "kelurahan":
            insertKelurahan(kode, row.nama, extractKecamatanKode(kode))
        }
    }
}
```

---

## Step 4: Import Metadata

### 4a. Data Penduduk
**Source:** `wilayah/db/wilayah_penduduk.sql`

```sql
UPDATE provinsi SET 
    penduduk_total = src.total,
    penduduk_pria = src.pria,
    penduduk_wanita = src.wanita
FROM wilayah_penduduk src
WHERE LEFT(REPLACE(src.kode, '.', ''), 2) = provinsi.kode;
```

### 4b. Data Luas
**Source:** `wilayah/db/wilayah_luas.sql`

```sql
UPDATE provinsi SET luas = src.luas
FROM wilayah_luas src
WHERE LEFT(REPLACE(src.kode, '.', ''), 2) = provinsi.kode;

UPDATE kabupaten SET luas = src.luas
FROM wilayah_luas src
WHERE LEFT(REPLACE(src.kode, '.', ''), 4) = kabupaten.kode;
```

### 4c. Data Pulau
**Source:** `wilayah/db/wilayah_pulau.sql`

```sql
INSERT INTO pulau (kode, nama, lat, lng, luas, status, notes)
SELECT kode, nama, lat, lng, luas, status, notes
FROM wilayah_pulau;
```

---

## Step 5: Import Boundaries

**Source:** `wilayah_boundaries/db/`

**Challenge:**
- Path format: `'[[lat,lng],[lat,lng],...]'` (JSON string)
- PostGIS needs: GeoJSON dengan `[lng,lat]` order

**Transformation Steps:**
```
1. Read SQL file
2. Parse path JSON
3. Swap [lat,lng] → [lng,lat]
4. Build GeoJSON structure
5. Convert to PostGIS geometry
6. Update table
```

**Import Script Logic:**
```go
func importBoundaries() {
    // Read wilayah_boundaries SQL
    for _, row := range rows {
        // Parse path JSON
        coords := parsePath(row.path)
        
        // Swap lat/lng order
        swapped := swapCoordinates(coords)
        
        // Build GeoJSON
        geojson := buildGeoJSON(swapped)
        
        // Update geometry
        updateGeometry(row.kode, geojson)
    }
}
```

**SQL Update:**
```sql
UPDATE provinsi
SET geometry = ST_SetSRID(
    ST_GeomFromGeoJSON($1), 
    4326
),
lat = $2,
lng = $3
WHERE kode = $4;
```

---

## Step 6: Import Postal Codes

**Source:** `wilayah_kodepos/db/wilayah_kodepos.sql`

**Transformation:**
```
Source: kode = "11.01.01.2001"
Target: kelurahan_kode = "1101012001"
```

**Import Script:**
```go
func importKodepos() {
    for _, row := range rows {
        kelurahanKode := cleanKode(row.kode)
        insertKodepos(row.kodepos, kelurahanKode)
    }
}
```

---

## Step 7: Setup Logos

**Source:** `wilayah_logo/`

**Options:**

### Option A: Local Storage
```bash
# Copy logos to static directory
cp -r wilayah_logo/prov/img/* ./static/logos/prov/
cp -r wilayah_logo/kab/*/img/* ./static/logos/kab/

# Update database
UPDATE provinsi 
SET logo_url = '/logos/prov/' || kode || '.png';

UPDATE kabupaten 
SET logo_url = '/logos/kab/' || kode || '.png';
```

### Option B: MinIO/S3
```go
// Upload to MinIO
for _, logo := range logos {
    uploadToMinIO(logo.path, "logos/" + logo.filename)
    updateLogoURL(logo.kode, minioURL)
}
```

---

## Step 8: Validation

### Check Record Counts
```sql
SELECT 'provinsi' AS level, COUNT(*) FROM provinsi
UNION ALL
SELECT 'kabupaten', COUNT(*) FROM kabupaten
UNION ALL
SELECT 'kecamatan', COUNT(*) FROM kecamatan
UNION ALL
SELECT 'kelurahan', COUNT(*) FROM kelurahan
UNION ALL
SELECT 'kodepos', COUNT(*) FROM kodepos
UNION ALL
SELECT 'pulau', COUNT(*) FROM pulau;
```

**Expected:**
```
provinsi:   34
kabupaten:  514
kecamatan:  ~7,000
kelurahan:  ~80,000
kodepos:    ~85,000
pulau:      ~17,000
```

### Check Geometry Validity
```sql
SELECT COUNT(*) 
FROM provinsi 
WHERE NOT ST_IsValid(geometry);
```

### Check Hierarchy Integrity
```sql
-- Orphaned kabupaten
SELECT COUNT(*) 
FROM kabupaten kb
LEFT JOIN provinsi p ON kb.provinsi_kode = p.kode
WHERE p.kode IS NULL;
```

---

## Error Handling

### Common Issues

1. **Duplicate codes**
   - Solution: Use INSERT ... ON CONFLICT DO UPDATE

2. **Invalid geometry**
   - Solution: ST_MakeValid(geometry)

3. **Missing parent**
   - Solution: Import in order (parent first)

4. **Encoding issues**
   - Solution: Use UTF-8 encoding

---

## Performance Tips

1. **Batch inserts** (1000 rows per INSERT)
2. **Disable indexes during import** (rebuild after)
3. **Use transactions** (rollback on error)
4. **Parallel import** (different tables)

---

## Rollback Plan

```bash
# Drop all tables
psql geomapping_id -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Re-run migrations
psql geomapping_id -f migrations/001_schema.sql

# Re-import data
./import-all
```

---

[← Previous: Database Schema](./04-database-schema.md) | [Back to Index](./README.md) | [Next: API Design →](./06-api-design.md)
