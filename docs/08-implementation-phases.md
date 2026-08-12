# Implementation Phases

**Document:** 08  
**Status:** Planning

---

## Overview

Implementasi dibagi menjadi 6 fase. Setiap fase deliverable & bisa di-test.

---

## Phase 1: Database Setup

**Duration:** 1-2 hari  
**Goal:** PostgreSQL + PostGIS ready dengan schema

### Tasks
- [ ] Setup Docker PostgreSQL + PostGIS
- [ ] Create database schema (all tables)
- [ ] Create indexes (spatial + btree)
- [ ] Test connection dari Golang
- [ ] Verify schema correctness

### Deliverables
- ✅ `docker-compose.yml`
- ✅ `migrations/001_schema.sql`
- ✅ Basic Golang connection test

### Validation
```bash
# Test connection
psql -h localhost -U postgres -d geomapping_id

# Check tables
\dt

# Check PostGIS
SELECT PostGIS_Version();
```

---

## Phase 2: Import Master Data

**Duration:** 1-2 hari  
**Goal:** Import data wilayah dari cahyadsn/wilayah

### Tasks
- [ ] Write import script untuk `wilayah.sql`
- [ ] Split data by level (provinsi/kabupaten/kecamatan/kelurahan)
- [ ] Import `wilayah_penduduk.sql`
- [ ] Import `wilayah_luas.sql`
- [ ] Import `wilayah_pulau.sql`
- [ ] Validate record counts

### Deliverables
- ✅ `scripts/import_master.go`
- ✅ Data imported ke database
- ✅ Validation report

### Validation
```sql
SELECT 'provinsi', COUNT(*) FROM provinsi
UNION ALL SELECT 'kabupaten', COUNT(*) FROM kabupaten
UNION ALL SELECT 'kecamatan', COUNT(*) FROM kecamatan
UNION ALL SELECT 'kelurahan', COUNT(*) FROM kelurahan;
```

---

## Phase 3: Import Geometry Data

**Duration:** 2-3 hari  
**Goal:** Import boundaries dari cahyadsn/wilayah_boundaries

### Tasks
- [ ] Write conversion script (path TEXT → PostGIS GEOMETRY)
- [ ] Parse JSON path format
- [ ] Swap coordinate order (lat/lng → lng/lat)
- [ ] Import provinsi boundaries
- [ ] Import kabupaten boundaries
- [ ] Import kecamatan boundaries
- [ ] Import kelurahan boundaries
- [ ] Create spatial indexes
- [ ] Validate geometry

### Deliverables
- ✅ `scripts/import_boundaries.go`
- ✅ All geometry imported
- ✅ Spatial indexes created
- ✅ Validation report

### Validation
```sql
-- Check geometry validity
SELECT COUNT(*) FROM provinsi WHERE NOT ST_IsValid(geometry);

-- Test spatial query
SELECT nama FROM provinsi 
WHERE ST_Contains(geometry, ST_SetSRID(ST_MakePoint(106.82, -6.18), 4326));
```

---

## Phase 4: Import Kodepos + Logos

**Duration:** 1 hari  
**Goal:** Import postal codes & setup logo storage

### Tasks
- [ ] Write import script untuk `wilayah_kodepos.sql`
- [ ] Convert kode format (remove dots)
- [ ] Setup logo storage (local/MinIO)
- [ ] Copy logo files
- [ ] Update logo_url di database
- [ ] Validate data

### Deliverables
- ✅ `scripts/import_kodepos.go`
- ✅ Logo storage setup
- ✅ Data validated

### Validation
```sql
-- Check kodepos
SELECT COUNT(*) FROM kodepos;

-- Check logo URLs
SELECT COUNT(*) FROM provinsi WHERE logo_url IS NOT NULL;
```

---

## Phase 5: Build API Service

**Duration:** 3-5 hari  
**Goal:** Golang REST API dengan core endpoints

### Tasks
- [ ] Setup project structure
- [ ] Implement reverse geocoding endpoint
- [ ] Implement wilayah lookup endpoint
- [ ] Implement search endpoint
- [ ] Implement hierarchy endpoints
- [ ] Implement kodepos endpoints
- [ ] Implement boundaries endpoint
- [ ] Add error handling
- [ ] Add logging
- [ ] Test all endpoints

### Deliverables
- ✅ Complete API service
- ✅ All endpoints working
- ✅ Error handling
- ✅ Logging

### Validation
```bash
# Test reverse geocoding
curl "http://localhost:8080/api/v1/reverse?lat=-6.18&lng=106.82"

# Test wilayah lookup
curl "http://localhost:8080/api/v1/wilayah/3171010001"

# Test search
curl "http://localhost:8080/api/v1/search?q=menteng"
```

---

## Phase 6: Testing & Documentation

**Duration:** 2-3 hari  
**Goal:** Comprehensive testing & documentation

### Tasks
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Performance testing
- [ ] Load testing
- [ ] Write API documentation
- [ ] Write deployment guide
- [ ] Write user guide
- [ ] Code review
- [ ] Final validation

### Deliverables
- ✅ Test suite
- ✅ API documentation
- ✅ Deployment guide
- ✅ User guide

### Validation
```bash
# Run tests
go test ./...

# Load test
hey -n 1000 -c 50 http://localhost:8080/api/v1/reverse?lat=-6.18&lng=106.82
```

---

## Total Timeline

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 1: Database Setup | 1-2 days | 1-2 days |
| Phase 2: Import Master | 1-2 days | 2-4 days |
| Phase 3: Import Geometry | 2-3 days | 4-7 days |
| Phase 4: Import Kodepos | 1 day | 5-8 days |
| Phase 5: Build API | 3-5 days | 8-13 days |
| Phase 6: Testing | 2-3 days | 10-16 days |

**Total:** 10-16 hari (belajar sambil jalan)

---

## Phase Dependencies

```
Phase 1 (Database)
    ↓
Phase 2 (Master Data)
    ↓
Phase 3 (Geometry)
    ↓
Phase 4 (Kodepos + Logos)
    ↓
Phase 5 (API Service)
    ↓
Phase 6 (Testing)
```

**Note:** Phase 1-4 harus sequential. Phase 5 bisa mulai setelah Phase 3 (minimal).

---

## Risk Mitigation

### If stuck di Phase 3 (Geometry):
- Skip geometry import temporarily
- Lanjut ke Phase 5 dengan basic endpoints
- Return ke Phase 3 nanti

### If performance issues:
- Add Redis cache
- Optimize queries
- Add materialized views

### If data quality issues:
- Validate data early
- Fix import scripts
- Re-import if needed

---

## Success Criteria per Phase

### Phase 1
- ✅ Database running
- ✅ Schema created
- ✅ Connection working

### Phase 2
- ✅ Master data imported
- ✅ Record counts correct
- ✅ Hierarchy intact

### Phase 3
- ✅ Geometry imported
- ✅ Spatial queries working
- ✅ Performance acceptable

### Phase 4
- ✅ Kodepos imported
- ✅ Logos accessible
- ✅ Data validated

### Phase 5
- ✅ All endpoints working
- ✅ Response time < 50ms
- ✅ Error handling correct

### Phase 6
- ✅ Tests passing
- ✅ Documentation complete
- ✅ Ready for deployment

---

[← Previous: Tech Stack](./07-tech-stack.md) | [Back to Index](./README.md) | [Next: Testing Strategy →](./09-testing.md)
