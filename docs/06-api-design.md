# API Design

**Document:** 06  
**Status:** Planning  
**Baseline:** [cahyadsn/wilayah_api](https://github.com/cahyadsn/wilayah_api)

---

## Overview

API design berdasarkan **minimum features dari wilayah_api** + **value-added features** untuk geomapping service.

**Stack:**
- Language: Golang
- Framework: Gin/Echo
- Database: PostgreSQL + PostGIS
- Cache: Redis (optional)

---

## Feature Comparison

### Features dari wilayah_api (Minimum Required)

| Feature | wilayah_api (PHP) | Our API (Golang) | Status |
|---------|-------------------|------------------|--------|
| List wilayah by type | ✅ | ✅ | Required |
| Search by name | ✅ | ✅ | Required |
| Filter by parent_id | ✅ | ✅ | Required |
| Pagination | ✅ | ✅ | Required |
| JWT Authentication | ✅ | ✅ | Required |
| Rate Limiting | ✅ | ✅ | Required |
| Caching | ✅ (file-based) | ✅ (Redis) | Required |

### Value-Added Features (Our Differentiation)

| Feature | wilayah_api | Our API | Status |
|---------|-------------|---------|--------|
| **Reverse Geocoding** | ❌ | ✅ | **Core Feature** |
| **Spatial Queries (PostGIS)** | ❌ | ✅ | **Core Feature** |
| **Boundary Retrieval (GeoJSON)** | ❌ | ✅ | Value-add |
| **Postal Code Lookup** | ❌ | ✅ | Value-add |
| **Batch Operations** | ❌ | ✅ | Value-add |
| **Hierarchy Navigation** | Partial | ✅ | Enhanced |
| **Centroid Coordinates** | ❌ | ✅ | Value-add |
| **Area/Luas Info** | ❌ | ✅ | Value-add |
| **Population Data** | ❌ | ✅ | Value-add |

---

## Endpoints Summary

### Phase 1: Minimum Features (from wilayah_api)

| No | Method | Endpoint | Description | Priority |
|----|--------|----------|-------------|----------|
| 1 | POST | `/auth` | Generate JWT token | P0 |
| 2 | GET | `/wilayah` | List wilayah by type | P0 |
| 3 | GET | `/wilayah/:kode` | Get wilayah detail | P0 |
| 4 | GET | `/health` | Health check | P0 |

### Phase 2: Enhanced Features

| No | Method | Endpoint | Description | Priority |
|----|--------|----------|-------------|----------|
| 5 | GET | `/reverse` | Reverse geocoding | P0 |
| 6 | GET | `/search` | Advanced search | P1 |
| 7 | GET | `/hierarchy/:kode` | Get full hierarchy | P1 |
| 8 | GET | `/boundaries/:kode` | Get GeoJSON boundary | P1 |

### Phase 3: Value-Added Features

| No | Method | Endpoint | Description | Priority |
|----|--------|----------|-------------|----------|
| 9 | GET | `/kodepos/:kode` | Lookup by postal code | P2 |
| 10 | GET | `/kodepos` | Reverse postal code lookup | P2 |
| 11 | POST | `/batch/reverse` | Batch reverse geocoding | P2 |
| 12 | GET | `/provinsi` | List all provinsi | P2 |
| 13 | GET | `/provinsi/:kode/kabupaten` | List kabupaten in provinsi | P2 |
| 14 | GET | `/kabupaten/:kode/kecamatan` | List kecamatan | P2 |
| 15 | GET | `/kecamatan/:kode/kelurahan` | List kelurahan | P2 |

---

## Detailed API Specifications

### Phase 1: Minimum Features

---

#### 1. POST /auth

**Purpose:** Generate JWT token (same as wilayah_api)

**Request:**
```http
POST /api/v1/auth
Content-Type: application/json

{
  "public_key": "pk_test_12345",
  "private_key": "sk_test_67890"
}
```

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400
}
```

**Implementation Notes:**
```go
// Golang implementation
func handleAuth(c *gin.Context) {
    var req AuthRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse("Invalid request"))
        return
    }
    
    // Validate credentials
    if !validateCredentials(req.PublicKey, req.PrivateKey) {
        c.JSON(401, ErrorResponse("Invalid credentials"))
        return
    }
    
    // Generate JWT token
    token, err := generateJWT(req.PublicKey)
    if err != nil {
        c.JSON(500, ErrorResponse("Failed to generate token"))
        return
    }
    
    c.JSON(200, AuthResponse{
        Success: true,
        Token: token,
        ExpiresIn: 86400, // 24 hours
    })
}
```

---

#### 2. GET /wilayah

**Purpose:** List wilayah by type with filtering, search, pagination (same as wilayah_api)

**Request:**
```http
GET /api/v1/wilayah?type=kabupaten&parent_id=11&search=aceh&page=1&limit=10
Authorization: Bearer {token}
```

**Parameters:**
| Param | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| type | string | ❌ | provinsi | provinsi/kabupaten/kecamatan/kelurahan |
| parent_id | string | ❌ | - | Filter by parent kode |
| search | string | ❌ | - | Search by nama |
| page | int | ❌ | 1 | Page number |
| limit | int | ❌ | 10 | Items per page (max: 100) |

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "kode": "11.01",
      "nama": "Kabupaten Aceh Selatan"
    },
    {
      "kode": "11.02",
      "nama": "Kabupaten Aceh Tenggara"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 23,
    "total_pages": 3
  },
  "rate_limit": {
    "limit": 100,
    "remaining": 95
  }
}
```

**Implementation Notes:**
```go
func handleWilayahList(c *gin.Context) {
    // Parse query parameters
    tipe := c.DefaultQuery("type", "provinsi")
    parentID := c.Query("parent_id")
    search := c.Query("search")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    // Validate
    if limit > 100 {
        limit = 100
    }
    
    // Build query based on type
    var query string
    var args []interface{}
    
    switch tipe {
    case "provinsi":
        query = "SELECT kode, nama FROM provinsi"
    case "kabupaten":
        query = "SELECT kode, nama FROM kabupaten"
        if parentID != "" {
            query += " WHERE provinsi_kode = $1"
            args = append(args, parentID)
        }
    case "kecamatan":
        query = "SELECT kode, nama FROM kecamatan"
        if parentID != "" {
            query += " WHERE kabupaten_kode = $1"
            args = append(args, parentID)
        }
    case "kelurahan":
        query = "SELECT kode, nama FROM kelurahan"
        if parentID != "" {
            query += " WHERE kecamatan_kode = $1"
            args = append(args, parentID)
        }
    }
    
    // Add search filter
    if search != "" {
        query += " AND nama ILIKE $2"
        args = append(args, "%"+search+"%")
    }
    
    // Add pagination
    offset := (page - 1) * limit
    query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
    args = append(args, limit, offset)
    
    // Execute query
    rows, err := db.Query(c, query, args...)
    // ... handle results
}
```

---

#### 3. GET /wilayah/:kode

**Purpose:** Get detail wilayah by kode

**Request:**
```http
GET /api/v1/wilayah/11.01
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "kode": "11.01",
    "nama": "Kabupaten Aceh Selatan",
    "type": "kabupaten",
    "parent": {
      "kode": "11",
      "nama": "Aceh"
    },
    "centroid": {
      "lat": -3.2345,
      "lng": 97.4567
    },
    "luas": 4200.5,
    "penduduk": {
      "total": 250000,
      "pria": 126000,
      "wanita": 124000
    },
    "logo_url": "/logos/kab/11.01.png"
  }
}
```

**Implementation Notes:**
```go
func handleWilayahDetail(c *gin.Context) {
    kode := c.Param("kode")
    
    // Detect level by kode length
    level := detectLevel(kode)
    
    var query string
    switch level {
    case "provinsi":
        query = `
            SELECT kode, nama, lat, lng, luas, 
                   penduduk_total, penduduk_pria, penduduk_wanita,
                   logo_url
            FROM provinsi WHERE kode = $1
        `
    case "kabupaten":
        query = `
            SELECT k.kode, k.nama, k.lat, k.lng, k.luas,
                   k.penduduk_total, k.penduduk_pria, k.penduduk_wanita,
                   k.logo_url, p.kode as parent_kode, p.nama as parent_nama
            FROM kabupaten k
            JOIN provinsi p ON k.provinsi_kode = p.kode
            WHERE k.kode = $1
        `
    // ... similar for kecamatan, kelurahan
    }
    
    // Execute query
    // ...
}
```

---

### Phase 2: Enhanced Features (Our Value-Add)

---

#### 5. GET /reverse ⭐ CORE FEATURE

**Purpose:** Reverse geocoding - find wilayah by coordinates (NOT in wilayah_api)

**Request:**
```http
GET /api/v1/reverse?lat=-6.18&lng=106.82
Authorization: Bearer {token}
```

**Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| lat | float | ✅ | Latitude (-11 to 6) |
| lng | float | ✅ | Longitude (95 to 141) |
| level | string | ❌ | Filter level (default: most specific) |

**Response:**
```json
{
  "success": true,
  "data": {
    "input": {
      "lat": -6.18,
      "lng": 106.82
    },
    "provinsi": {
      "kode": "31",
      "nama": "DKI Jakarta"
    },
    "kabupaten": {
      "kode": "3171",
      "nama": "Jakarta Pusat"
    },
    "kecamatan": {
      "kode": "3171010",
      "nama": "Menteng"
    },
    "kelurahan": {
      "kode": "3171010001",
      "nama": "Gondangdia"
    },
    "kodepos": ["10350"],
    "centroid": {
      "lat": -6.1944,
      "lng": 106.8472
    }
  }
}
```

**Implementation Notes (PostGIS):**
```go
func handleReverseGeocoding(c *gin.Context) {
    lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
    lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
    
    // Validate coordinates (Indonesia bounds)
    if lat < -11 || lat > 6 || lng < 95 || lng > 141 {
        c.JSON(400, ErrorResponse("Coordinates out of Indonesia bounds"))
        return
    }
    
    // PostGIS query - find kelurahan containing point
    query := `
        SELECT 
            kl.kode as kel_kode, kl.nama as kel_nama,
            kc.kode as kec_kode, kc.nama as kec_nama,
            kb.kode as kab_kode, kb.nama as kab_nama,
            p.kode as prov_kode, p.nama as prov_nama
        FROM kelurahan kl
        JOIN kecamatan kc ON kl.kecamatan_kode = kc.kode
        JOIN kabupaten kb ON kc.kabupaten_kode = kb.kode
        JOIN provinsi p ON kb.provinsi_kode = p.kode
        WHERE ST_Contains(
            kl.geometry, 
            ST_SetSRID(ST_MakePoint($1, $2), 4326)
        )
        LIMIT 1
    `
    
    var result ReverseGeocodingResult
    err := db.QueryRow(c, query, lng, lat).Scan(
        &result.Kelurahan.Kode, &result.Kelurahan.Nama,
        &result.Kecamatan.Kode, &result.Kecamatan.Nama,
        &result.Kabupaten.Kode, &result.Kabupaten.Nama,
        &result.Provinsi.Kode, &result.Provinsi.Nama,
    )
    
    if err == sql.ErrNoRows {
        c.JSON(404, ErrorResponse("No wilayah found for coordinates"))
        return
    }
    
    // Get postal codes
    kodeposQuery := `
        SELECT kodepos FROM kodepos 
        WHERE kelurahan_kode = $1
    `
    // ... fetch kodepos
    
    c.JSON(200, SuccessResponse(result))
}
```

**Why This is Core Feature:**
- ✅ Main use case: "Where am I?"
- ✅ Uses PostGIS spatial index (fast)
- ✅ Differentiator from wilayah_api
- ✅ High value for users

---

#### 6. GET /search

**Purpose:** Advanced search with fuzzy matching

**Request:**
```http
GET /api/v1/search?q=menteng&type=kecamatan&limit=10
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "kode": "3171010",
      "nama": "Menteng",
      "type": "kecamatan",
      "parent": {
        "kode": "3171",
        "nama": "Jakarta Pusat"
      },
      "province": "DKI Jakarta"
    }
  ],
  "total": 1
}
```

**Implementation Notes:**
```go
func handleSearch(c *gin.Context) {
    q := c.Query("q")
    tipe := c.Query("type") // optional filter
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    // Use pg_trgm for fuzzy search
    query := `
        SELECT kode, nama, 'kelurahan' as type
        FROM kelurahan
        WHERE nama % $1  -- similarity search
        ORDER BY similarity(nama, $1) DESC
        LIMIT $2
        
        UNION ALL
        
        SELECT kode, nama, 'kecamatan' as type
        FROM kecamatan
        WHERE nama % $1
        ORDER BY similarity(nama, $1) DESC
        LIMIT $2
        
        -- ... repeat for kabupaten, provinsi
    `
    
    // Execute and return results
}
```

---

#### 8. GET /boundaries/:kode

**Purpose:** Get GeoJSON boundary for visualization

**Request:**
```http
GET /api/v1/boundaries/31
Authorization: Bearer {token}
```

**Response:**
```json
{
  "type": "Feature",
  "properties": {
    "kode": "31",
    "nama": "DKI Jakarta",
    "type": "provinsi"
  },
  "geometry": {
    "type": "MultiPolygon",
    "coordinates": [[[[106.6,-6.1],[106.7,-6.1],...]]]
  }
}
```

**Implementation Notes:**
```go
func handleBoundary(c *gin.Context) {
    kode := c.Param("kode")
    level := detectLevel(kode)
    
    var tableName string
    switch level {
    case "provinsi":
        tableName = "provinsi"
    case "kabupaten":
        tableName = "kabupaten"
    // ...
    }
    
    // Get geometry as GeoJSON
    query := fmt.Sprintf(`
        SELECT 
            kode, nama,
            ST_AsGeoJSON(geometry) as geojson
        FROM %s
        WHERE kode = $1
    `, tableName)
    
    var geojson string
    err := db.QueryRow(c, query, kode).Scan(&kode, &nama, &geojson)
    
    // Return as GeoJSON Feature
    c.JSON(200, gin.H{
        "type": "Feature",
        "properties": gin.H{
            "kode": kode,
            "nama": nama,
            "type": level,
        },
        "geometry": json.RawMessage(geojson),
    })
}
```

---

### Phase 3: Value-Added Features

---

#### 9-10. Postal Code Endpoints

**GET /kodepos/:kode**
```http
GET /api/v1/kodepos/10350
```

**Response:**
```json
{
  "success": true,
  "data": {
    "kodepos": "10350",
    "wilayah": {
      "kode": "3171010001",
      "nama": "Gondangdia",
      "kecamatan": "Menteng",
      "kabupaten": "Jakarta Pusat",
      "provinsi": "DKI Jakarta"
    }
  }
}
```

**GET /kodepos?wilayah=3171010001**
```http
GET /api/v1/kodepos?wilayah=3171010001
```

**Response:**
```json
{
  "success": true,
  "data": {
    "kodepos": ["10350"]
  }
}
```

---

#### 11. POST /batch/reverse

**Purpose:** Batch reverse geocoding (up to 100 points)

**Request:**
```http
POST /api/v1/batch/reverse
Authorization: Bearer {token}
Content-Type: application/json

{
  "points": [
    {"lat": -6.18, "lng": 106.82},
    {"lat": -6.91, "lng": 107.61},
    {"lat": -7.79, "lng": 110.36}
  ]
}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "input": {"lat": -6.18, "lng": 106.82},
      "result": {
        "provinsi": "DKI Jakarta",
        "kabupaten": "Jakarta Pusat",
        ...
      }
    },
    ...
  ]
}
```

**Implementation Notes:**
```go
func handleBatchReverse(c *gin.Context) {
    var req BatchReverseRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse("Invalid request"))
        return
    }
    
    if len(req.Points) > 100 {
        c.JSON(400, ErrorResponse("Max 100 points per request"))
        return
    }
    
    // Process each point
    results := make([]BatchResult, len(req.Points))
    for i, point := range req.Points {
        result := reverseGeocode(point.Lat, point.Lng)
        results[i] = BatchResult{
            Input: point,
            Result: result,
        }
    }
    
    c.JSON(200, SuccessResponse(results))
}
```

---

## Implementation Priority

### P0: Must Have (Week 1-2)
- ✅ POST /auth
- ✅ GET /wilayah (list by type)
- ✅ GET /wilayah/:kode (detail)
- ✅ GET /reverse (core feature)
- ✅ GET /health

### P1: Should Have (Week 2-3)
- ✅ GET /search (advanced search)
- ✅ GET /hierarchy/:kode
- ✅ GET /boundaries/:kode

### P2: Nice to Have (Week 3-4)
- ✅ GET /kodepos/:kode
- ✅ GET /kodepos
- ✅ POST /batch/reverse
- ✅ GET /provinsi
- ✅ GET /provinsi/:kode/kabupaten
- ✅ GET /kabupaten/:kode/kecamatan
- ✅ GET /kecamatan/:kode/kelurahan

---

## Error Handling

### Standard Error Response
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Wilayah not found",
    "details": {}
  }
}
```

### Error Codes
| Code | HTTP Status | Description |
|------|-------------|-------------|
| SUCCESS | 200 | OK |
| BAD_REQUEST | 400 | Invalid parameters |
| UNAUTHORIZED | 401 | Invalid/missing token |
| NOT_FOUND | 404 | Resource not found |
| RATE_LIMIT | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Server error |

---

## Rate Limiting

**Strategy:** Per API key, sliding window

**Limits:**
- Free tier: 100 requests/hour
- Pro tier: 1000 requests/hour
- Enterprise: Custom

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

**Implementation:**
```go
func rateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("Authorization")
        
        // Check rate limit
        count, err := redis.Incr(apiKey + ":rate_limit")
        if count == 1 {
            redis.Expire(apiKey + ":rate_limit", time.Hour)
        }
        
        if count > 100 {
            c.JSON(429, ErrorResponse("Rate limit exceeded"))
            c.Abort()
            return
        }
        
        // Set headers
        c.Header("X-RateLimit-Limit", "100")
        c.Header("X-RateLimit-Remaining", strconv.Itoa(100-int(count)))
        
        c.Next()
    }
}
```

---

## Caching Strategy

**Cache Layers:**
1. **Redis Cache** - Query results (TTL: 1 hour)
2. **HTTP Cache** - Static responses (Cache-Control)

**Cache Keys:**
```
wilayah:list:{type}:{parent_id}:{search}:{page}:{limit}
wilayah:detail:{kode}
reverse:{lat}:{lng}
boundaries:{kode}
```

**Implementation:**
```go
func cacheMiddleware(ttl time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        cacheKey := generateCacheKey(c)
        
        // Try cache first
        cached, err := redis.Get(cacheKey)
        if err == nil {
            c.Header("X-Cache", "HIT")
            c.Data(200, "application/json", []byte(cached))
            c.Abort()
            return
        }
        
        // Cache miss, continue to handler
        c.Header("X-Cache", "MISS")
        c.Next()
        
        // Cache response
        redis.Set(cacheKey, c.Writer.Body.String(), ttl)
    }
}
```

---

## API Documentation

**Tools:** Swagger/OpenAPI 3.0

**Auto-generated from code:**
```go
// @Summary Reverse geocoding
// @Description Get wilayah hierarchy by coordinates
// @Tags Geocoding
// @Accept json
// @Produce json
// @Param lat query float true "Latitude"
// @Param lng query float true "Longitude"
// @Success 200 {object} ReverseResponse
// @Failure 400 {object} ErrorResponse
// @Router /reverse [get]
// @Security BearerAuth
func handleReverse(c *gin.Context) {
    // ...
}
```

**Generate docs:**
```bash
swag init -g cmd/main.go
```

---

## Comparison Summary

### What We Have vs wilayah_api

| Aspect | wilayah_api (PHP) | Our API (Golang) |
|--------|-------------------|------------------|
| **Language** | PHP 7.4+ | Golang 1.21+ |
| **Database** | MySQL | PostgreSQL + PostGIS |
| **Performance** | ~5ms (optimized) | <50ms target |
| **Spatial Queries** | ❌ No | ✅ Yes (PostGIS) |
| **Reverse Geocoding** | ❌ No | ✅ Yes (Core) |
| **Authentication** | ✅ JWT | ✅ JWT |
| **Rate Limiting** | ✅ File-based | ✅ Redis |
| **Caching** | ✅ File-based | ✅ Redis |
| **API Docs** | ✅ OpenAPI | ✅ OpenAPI |

### Our Advantages

1. ✅ **PostGIS Spatial Queries** - Fast point-in-polygon
2. ✅ **Reverse Geocoding** - Core feature not in wilayah_api
3. ✅ **Golang Performance** - Better concurrency
4. ✅ **PostgreSQL** - More robust than MySQL
5. ✅ **Redis Caching** - Faster than file-based
6. ✅ **GeoJSON Boundaries** - For map visualization
7. ✅ **Batch Operations** - Process multiple points

---

[← Previous: Data Import](./05-data-import.md) | [Back to Index](./README.md) | [Next: Tech Stack →](./07-tech-stack.md)
