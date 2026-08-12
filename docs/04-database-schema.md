# Database Schema Design

**Document:** 04  
**Status:** Planning

---

## Design Principles

1. **Normalization** - Reduce redundancy
2. **Hierarchy** - Proper parent-child with FK
3. **Spatial Optimization** - PostGIS types + indexes
4. **Performance** - Indexes on frequent queries
5. **Maintainability** - Clear naming

---

## Schema Overview

```
provinsi (34)
    │
    └─→ kabupaten (514)
            │
            └─→ kecamatan (~7K)
                    │
                    └─→ kelurahan (~80K)
                            │
                            └─→ kodepos (~85K)

pulau (~17K) - standalone
```

---

## Table: provinsi

```sql
CREATE TABLE provinsi (
    kode VARCHAR(2) PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    
    -- Geographic
    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),
    
    -- Metadata
    luas DECIMAL(10, 2),
    penduduk_total INTEGER,
    penduduk_pria INTEGER,
    penduduk_wanita INTEGER,
    zona_waktu VARCHAR(10),
    elevasi INTEGER,
    
    -- Asset
    logo_url VARCHAR(255),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_provinsi_geometry 
    ON provinsi USING GIST(geometry);
CREATE INDEX idx_provinsi_nama 
    ON provinsi USING gin(to_tsvector('simple', nama));
CREATE INDEX idx_provinsi_coords 
    ON provinsi(lat, lng);
```

---

## Table: kabupaten

```sql
CREATE TABLE kabupaten (
    kode VARCHAR(4) PRIMARY KEY,
    provinsi_kode VARCHAR(2) NOT NULL 
        REFERENCES provinsi(kode) ON DELETE CASCADE,
    nama VARCHAR(100) NOT NULL,
    
    -- Geographic
    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),
    
    -- Metadata
    luas DECIMAL(10, 2),
    penduduk_total INTEGER,
    penduduk_pria INTEGER,
    penduduk_wanita INTEGER,
    zona_waktu VARCHAR(10),
    elevasi INTEGER,
    
    -- Asset
    logo_url VARCHAR(255),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_kabupaten_geometry 
    ON kabupaten USING GIST(geometry);
CREATE INDEX idx_kabupaten_nama 
    ON kabupaten USING gin(to_tsvector('simple', nama));
CREATE INDEX idx_kabupaten_provinsi 
    ON kabupaten(provinsi_kode);
CREATE INDEX idx_kabupaten_coords 
    ON kabupaten(lat, lng);
```

---

## Table: kecamatan

```sql
CREATE TABLE kecamatan (
    kode VARCHAR(6) PRIMARY KEY,
    kabupaten_kode VARCHAR(4) NOT NULL 
        REFERENCES kabupaten(kode) ON DELETE CASCADE,
    nama VARCHAR(100) NOT NULL,
    
    -- Geographic
    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_kecamatan_geometry 
    ON kecamatan USING GIST(geometry);
CREATE INDEX idx_kecamatan_nama 
    ON kecamatan USING gin(to_tsvector('simple', nama));
CREATE INDEX idx_kecamatan_kabupaten 
    ON kecamatan(kabupaten_kode);
CREATE INDEX idx_kecamatan_coords 
    ON kecamatan(lat, lng);
```

---

## Table: kelurahan

```sql
CREATE TABLE kelurahan (
    kode VARCHAR(10) PRIMARY KEY,
    kecamatan_kode VARCHAR(6) NOT NULL 
        REFERENCES kecamatan(kode) ON DELETE CASCADE,
    nama VARCHAR(100) NOT NULL,
    
    -- Geographic
    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_kelurahan_geometry 
    ON kelurahan USING GIST(geometry);
CREATE INDEX idx_kelurahan_nama 
    ON kelurahan USING gin(to_tsvector('simple', nama));
CREATE INDEX idx_kelurahan_kecamatan 
    ON kelurahan(kecamatan_kode);
CREATE INDEX idx_kelurahan_coords 
    ON kelurahan(lat, lng);
```

---

## Table: kodepos

```sql
CREATE TABLE kodepos (
    kode VARCHAR(5) PRIMARY KEY,
    kelurahan_kode VARCHAR(10) NOT NULL 
        REFERENCES kelurahan(kode) ON DELETE CASCADE,
    
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_kodepos_kelurahan 
    ON kodepos(kelurahan_kode);
```

---

## Table: pulau

```sql
CREATE TABLE pulau (
    kode VARCHAR(11) PRIMARY KEY,
    nama VARCHAR(255) NOT NULL,
    
    -- Geographic
    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),
    luas DECIMAL(10, 2),
    
    -- Metadata
    status VARCHAR(50),
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_pulau_geometry 
    ON pulau USING GIST(geometry);
CREATE INDEX idx_pulau_nama 
    ON pulau USING gin(to_tsvector('simple', nama));
CREATE INDEX idx_pulau_coords 
    ON pulau(lat, lng);
```

---

## Design Decisions

### Why Separate Tables per Level?

**Pros:**
- ✅ Clear hierarchy with FK
- ✅ Better query performance
- ✅ Level-specific fields (logo only for prov/kab)
- ✅ Easier maintenance

**Cons:**
- ⚠️ More JOINs needed
- ⚠️ More tables

### Why Not Single Table?

```sql
-- Alternative: Single table
CREATE TABLE wilayah (
    kode VARCHAR(10) PRIMARY KEY,
    parent_kode VARCHAR(10) REFERENCES wilayah(kode),
    level INTEGER NOT NULL,
    nama VARCHAR(100),
    ...
);
```

**Why rejected:**
- ❌ Harder to enforce hierarchy
- ❌ Level-specific fields awkward
- ❌ Less clear schema

---

## Spatial Query Examples

### Reverse Geocoding
```sql
-- Find kelurahan containing point
SELECT k.kode, k.nama
FROM kelurahan k
WHERE ST_Contains(
    k.geometry, 
    ST_SetSRID(ST_MakePoint(106.82, -6.18), 4326)
)
LIMIT 1;
```

### Hierarchy Lookup
```sql
-- Get full hierarchy for a point
SELECT 
    p.nama AS provinsi,
    kb.nama AS kabupaten,
    kc.nama AS kecamatan,
    kl.nama AS kelurahan
FROM kelurahan kl
JOIN kecamatan kc ON kl.kecamatan_kode = kc.kode
JOIN kabupaten kb ON kc.kabupaten_kode = kb.kode
JOIN provinsi p ON kb.provinsi_kode = p.kode
WHERE ST_Contains(
    kl.geometry, 
    ST_SetSRID(ST_MakePoint(106.82, -6.18), 4326)
)
LIMIT 1;
```

### Search by Name
```sql
-- Fuzzy search
SELECT kode, nama
FROM kelurahan
WHERE nama ILIKE '%menteng%'
LIMIT 10;
```

---

## ER Diagram

```
┌──────────────┐
│  provinsi    │
│──────────────│
│ kode (PK)    │
│ nama         │
│ geometry     │
│ luas         │
│ penduduk_*   │
│ logo_url     │
└──────┬───────┘
       │ 1:N
       ▼
┌──────────────┐
│  kabupaten   │
│──────────────│
│ kode (PK)    │
│ provinsi_kode│──→ FK
│ nama         │
│ geometry     │
│ logo_url     │
└──────┬───────┘
       │ 1:N
       ▼
┌──────────────┐
│  kecamatan   │
│──────────────│
│ kode (PK)    │
│ kabupaten_kod│──→ FK
│ nama         │
│ geometry     │
└──────┬───────┘
       │ 1:N
       ▼
┌──────────────┐
│  kelurahan   │
│──────────────│
│ kode (PK)    │
│ kecamatan_kod│──→ FK
│ nama         │
│ geometry     │
└──────┬───────┘
       │ 1:N
       ▼
┌──────────────┐
│  kodepos     │
│──────────────│
│ kode (PK)    │
│ kelurahan_kod│──→ FK
└──────────────┘

┌──────────────┐
│  pulau       │
│──────────────│
│ kode (PK)    │
│ nama         │
│ geometry     │
│ luas         │
│ status       │
└──────────────┘
```

---

[← Previous: Architecture](./03-architecture.md) | [Back to Index](./README.md) | [Next: Data Import →](./05-data-import.md)
