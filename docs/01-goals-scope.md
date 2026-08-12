# Goals & Scope

**Document:** 01  
**Status:** Planning

---

## Project Overview

Membangun **standalone geomapping service** untuk Indonesia.

### Core Function
- **Input:** GPS coordinates (lat, lng)
- **Output:** Complete administrative hierarchy
- **Example:** 
  - Input: `-6.18, 106.82`
  - Output: `DKI Jakarta → Jakarta Pusat → Menteng → Gondangdia`

---

## Key Objectives

### 1. Standalone Service
- API yang bisa diintegrasikan
- Tidak bergantung ke aplikasi lain
- Independent deployment

### 2. High Performance
- Reverse geocoding: < 50ms (p95)
- Wilayah lookup: < 20ms (p95)
- Search: < 100ms (p95)

### 3. Scalable
- Handle 1000+ concurrent requests
- Support horizontal scaling
- Stateless design

### 4. Maintainable
- Clean architecture
- Proper schema
- Production-ready

### 5. Accurate
- Data sesuai Kepmendagri 2025
- 100% coverage
- Up-to-date

---

## Scope

### ✅ In Scope

**Data Coverage:**
- 34 Provinsi
- 514 Kabupaten/Kota
- ~7,000 Kecamatan
- ~80,000 Kelurahan/Desa
- ~85,000 Kodepos
- ~17,000 Pulau

**Features:**
- Reverse geocoding
- Wilayah hierarchy navigation
- Search by name
- Postal code lookup
- Boundary retrieval (GeoJSON)
- Batch operations

**Technical:**
- REST API
- PostgreSQL + PostGIS
- Golang implementation
- Docker deployment
- Comprehensive documentation

---

### ❌ Out of Scope

**Not Included:**
- Forward geocoding (address → coordinates)
- Route planning / navigation
- POI (Points of Interest) data
- Real-time traffic data
- Map rendering / tiles
- User authentication
- Payment / billing
- Mobile app development

---

## Core Use Cases

### 1. Reverse Geocoding
```
Input: lat, lng
Output: Complete hierarchy
Example: 
  GET /reverse?lat=-6.18&lng=106.82
  → {
      "provinsi": "DKI Jakarta",
      "kabupaten": "Jakarta Pusat",
      "kecamatan": "Menteng",
      "kelurahan": "Gondangdia",
      "kodepos": ["10350"]
    }
```

### 2. Wilayah Lookup
```
Input: wilayah code
Output: Detail + hierarchy
Example:
  GET /wilayah/3171010001
  → {
      "kode": "3171010001",
      "nama": "Gondangdia",
      "kecamatan": "Menteng",
      "kabupaten": "Jakarta Pusat",
      "provinsi": "DKI Jakarta"
    }
```

### 3. Search
```
Input: name query
Output: Matching results
Example:
  GET /search?q=menteng
  → [
      {"kode": "3171010", "nama": "Menteng", ...},
      ...
    ]
```

### 4. Postal Code
```
Input: wilayah code OR postal code
Output: Postal code mapping
Example:
  GET /kodepos?wilayah=3171010001
  → {"kodepos": ["10350"]}
  
  GET /kodepos/10350
  → {"wilayah": "Gondangdia", ...}
```

### 5. Boundaries
```
Input: wilayah code
Output: GeoJSON polygon
Example:
  GET /boundaries/31
  → {
      "type": "Feature",
      "geometry": {
        "type": "MultiPolygon",
        "coordinates": [...]
      }
    }
```

---

## Target Users

### Primary
- **Developers** - Integrate into their apps
- **E-commerce** - Delivery zones, address validation
- **Logistics** - Route planning, coverage areas

### Secondary
- **Government** - Public services, reporting
- **Researchers** - Geographic analysis
- **General Apps** - Location-based features

---

## Success Criteria

### Performance
- ✅ Reverse geocoding < 50ms (p95)
- ✅ Wilayah lookup < 20ms (p95)
- ✅ Search < 100ms (p95)

### Reliability
- ✅ Uptime: 99.9%
- ✅ Error rate: < 0.1%

### Scalability
- ✅ 1000+ concurrent requests
- ✅ Horizontal scaling support

### Data Quality
- ✅ 100% Kepmendagri coverage
- ✅ Accurate boundaries
- ✅ Latest regulations (2025)

---

## Constraints

### Technical
- Must use PostgreSQL + PostGIS
- Must be containerized (Docker)
- Must follow REST API standards

### Data
- Must use official Kepmendagri data
- Must maintain data integrity
- Must support future updates

### Timeline
- Phase-based approach
- Each phase deliverable
- Learning-friendly pace

---

## Assumptions

1. Data sources (cahyadsn repos) are accurate
2. Kepmendagri 2025 is latest regulation
3. Users have basic geolocation knowledge
4. Deployment environment supports Docker
5. Sufficient server resources available

---

## Dependencies

### External
- cahyadsn/wilayah repos
- PostgreSQL + PostGIS
- Golang ecosystem

### Internal
- Development environment
- Testing infrastructure
- Documentation tools

---

[← Back to Index](./README.md) | [Next: Data Sources →](./02-data-sources.md)
