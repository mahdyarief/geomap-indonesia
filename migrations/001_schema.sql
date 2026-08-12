-- Indonesia Geomapping Service - Schema
-- PostgreSQL 15+ with PostGIS 3.3+
-- Based on docs/04-database-schema.md

CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ---------------------------------------------------------------------------
-- provinsi (level 1)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS provinsi (
    kode VARCHAR(2) PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,

    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),

    luas DECIMAL(10, 2),
    penduduk_total INTEGER,
    penduduk_pria INTEGER,
    penduduk_wanita INTEGER,
    zona_waktu VARCHAR(10),
    elevasi INTEGER,

    logo_url VARCHAR(255),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_provinsi_geometry ON provinsi USING GIST(geometry);
CREATE INDEX IF NOT EXISTS idx_provinsi_nama ON provinsi USING gin (nama gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_provinsi_coords ON provinsi(lat, lng);

-- ---------------------------------------------------------------------------
-- kabupaten (level 2)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS kabupaten (
    kode VARCHAR(4) PRIMARY KEY,
    provinsi_kode VARCHAR(2) NOT NULL REFERENCES provinsi(kode) ON DELETE CASCADE,
    nama VARCHAR(100) NOT NULL,

    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),

    luas DECIMAL(10, 2),
    penduduk_total INTEGER,
    penduduk_pria INTEGER,
    penduduk_wanita INTEGER,
    zona_waktu VARCHAR(10),
    elevasi INTEGER,

    logo_url VARCHAR(255),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kabupaten_geometry ON kabupaten USING GIST(geometry);
CREATE INDEX IF NOT EXISTS idx_kabupaten_nama ON kabupaten USING gin (nama gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_kabupaten_provinsi ON kabupaten(provinsi_kode);
CREATE INDEX IF NOT EXISTS idx_kabupaten_coords ON kabupaten(lat, lng);

-- ---------------------------------------------------------------------------
-- kecamatan (level 3)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS kecamatan (
    kode VARCHAR(6) PRIMARY KEY,
    kabupaten_kode VARCHAR(4) NOT NULL REFERENCES kabupaten(kode) ON DELETE CASCADE,
    nama VARCHAR(100) NOT NULL,

    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kecamatan_geometry ON kecamatan USING GIST(geometry);
CREATE INDEX IF NOT EXISTS idx_kecamatan_nama ON kecamatan USING gin (nama gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_kecamatan_kabupaten ON kecamatan(kabupaten_kode);
CREATE INDEX IF NOT EXISTS idx_kecamatan_coords ON kecamatan(lat, lng);

-- ---------------------------------------------------------------------------
-- kelurahan (level 4)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS kelurahan (
    kode VARCHAR(10) PRIMARY KEY,
    kecamatan_kode VARCHAR(6) NOT NULL REFERENCES kecamatan(kode) ON DELETE CASCADE,
    nama VARCHAR(100) NOT NULL,

    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kelurahan_geometry ON kelurahan USING GIST(geometry);
CREATE INDEX IF NOT EXISTS idx_kelurahan_nama ON kelurahan USING gin (nama gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_kelurahan_kecamatan ON kelurahan(kecamatan_kode);
CREATE INDEX IF NOT EXISTS idx_kelurahan_coords ON kelurahan(lat, lng);

-- ---------------------------------------------------------------------------
-- kodepos
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS kodepos (
    kode VARCHAR(5) PRIMARY KEY,
    kelurahan_kode VARCHAR(10) NOT NULL REFERENCES kelurahan(kode) ON DELETE CASCADE,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kodepos_kelurahan ON kodepos(kelurahan_kode);

-- ---------------------------------------------------------------------------
-- pulau (standalone)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS pulau (
    kode VARCHAR(11) PRIMARY KEY,
    nama VARCHAR(255) NOT NULL,

    lat DECIMAL(10, 7),
    lng DECIMAL(10, 7),
    geometry GEOMETRY(MULTIPOLYGON, 4326),
    luas DECIMAL(10, 2),

    status VARCHAR(50),
    notes TEXT,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pulau_geometry ON pulau USING GIST(geometry);
CREATE INDEX IF NOT EXISTS idx_pulau_nama ON pulau USING gin (nama gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_pulau_coords ON pulau(lat, lng);
