# Technology Stack

**Document:** 07  
**Status:** Planning

---

## Overview

Technology choices untuk Indonesia Geomapping Service.

---

## Core Stack

| Layer | Technology | Version | Reason |
|-------|-----------|---------|--------|
| Language | Golang | 1.21+ | Performance, simplicity |
| Database | PostgreSQL | 15+ | Reliability, features |
| Spatial | PostGIS | 3.3+ | Gold standard for spatial |
| Cache | Redis | 7+ (optional) | High performance |
| Container | Docker | 24+ | Portability |

---

## Golang Libraries

### HTTP Framework
**Choice:** Gin or Echo

**Gin:**
```go
import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/reverse", handleReverse)
    r.Run(":8080")
}
```

**Why:**
- ✅ Fast performance
- ✅ Simple API
- ✅ Good documentation
- ✅ Large community

**Alternative:** Echo (similar performance, different API)

---

### Database Driver
**Choice:** pgx

```go
import "github.com/jackc/pgx/v5"

func queryWilayah(kode string) {
    row := pool.QueryRow(ctx, 
        "SELECT nama FROM wilayah WHERE kode = $1", kode)
}
```

**Why:**
- ✅ Native PostgreSQL support
- ✅ High performance
- ✅ Connection pooling built-in
- ✅ PostGIS compatible

---

### Configuration
**Choice:** Viper

```go
import "github.com/spf13/viper"

viper.SetConfigFile(".env")
viper.ReadInConfig()
dbHost := viper.GetString("DB_HOST")
```

**Why:**
- ✅ Multiple config formats
- ✅ Environment variables
- ✅ Remote config support

---

### Logging
**Choice:** Zap

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
logger.Info("Request received", zap.String("path", "/reverse"))
```

**Why:**
- ✅ Structured logging
- ✅ High performance
- ✅ Low overhead

---

### Validation
**Choice:** go-playground/validator

```go
type ReverseRequest struct {
    Lat float64 `validate:"required,min=-11,max=6"`
    Lng float64 `validate:"required,min=95,max=141"`
}
```

**Why:**
- ✅ Tag-based validation
- ✅ Custom validators
- ✅ Good error messages

---

## Database Extensions

### PostGIS
```sql
CREATE EXTENSION postgis;

-- Spatial types
GEOMETRY(MULTIPOLYGON, 4326)

-- Spatial functions
ST_Contains(geometry, point)
ST_Intersects(geometry, geometry)
ST_Distance(point1, point2)

-- Spatial indexes
CREATE INDEX idx_geom ON wilayah USING GIST(geometry);
```

**Why:**
- ✅ Industry standard
- ✅ Comprehensive spatial functions
- ✅ Excellent performance
- ✅ Active development

---

### pg_trgm (Optional)
```sql
CREATE EXTENSION pg_trgm;

-- Fuzzy search
SELECT * FROM wilayah 
WHERE nama % 'menteng';  -- similarity search
```

**Why:**
- ✅ Fuzzy text search
- ✅ Trigram indexes
- ✅ Good for typos

---

## Infrastructure Tools

### Docker
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:latest
COPY --from=builder /app/main /main
CMD ["/main"]
```

**Why:**
- ✅ Containerization
- ✅ Reproducible builds
- ✅ Easy deployment

---

### Docker Compose
```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
    depends_on:
      - postgres
  
  postgres:
    image: postgis/postgis:15-3.3
    environment:
      POSTGRES_DB: geomapping_id
      POSTGRES_PASSWORD: secret
    volumes:
      - pgdata:/var/lib/postgresql/data
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  pgdata:
```

**Why:**
- ✅ Local development
- ✅ Service orchestration
- ✅ Easy setup

---

## Optional Tools

### Load Balancer
**Choice:** Nginx or Traefik

**Why:**
- ✅ Traffic distribution
- ✅ SSL termination
- ✅ Health checks

---

### Monitoring
**Choice:** Prometheus + Grafana

**Why:**
- ✅ Metrics collection
- ✅ Visualization
- ✅ Alerting

---

### API Documentation
**Choice:** Swagger/OpenAPI

```go
// @Summary Reverse geocoding
// @Description Get wilayah by coordinates
// @Accept json
// @Produce json
// @Param lat query float true "Latitude"
// @Param lng query float true "Longitude"
// @Success 200 {object} ReverseResponse
// @Router /reverse [get]
func handleReverse(c *gin.Context) {
    // ...
}
```

**Why:**
- ✅ Auto-generated docs
- ✅ Interactive testing
- ✅ Client code generation

---

## Development Tools

### Testing
```go
func TestReverseGeocoding(t *testing.T) {
    req := httptest.NewRequest("GET", "/reverse?lat=-6.18&lng=106.82", nil)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

**Libraries:**
- `testing` (built-in)
- `testify` (assertions)
- `httptest` (HTTP testing)

---

### Code Quality
- **golangci-lint:** Linting
- **gofmt:** Formatting
- **go vet:** Static analysis

---

## Version Requirements

| Tool | Minimum | Recommended |
|------|---------|-------------|
| Golang | 1.21 | 1.22+ |
| PostgreSQL | 15 | 16+ |
| PostGIS | 3.3 | 3.4+ |
| Docker | 24 | 25+ |
| Redis | 7 | 7.2+ |

---

## Alternative Considerations

### Why Not Node.js?
- ❌ Slower performance
- ❌ Less mature spatial libraries
- ❌ Callback complexity

### Why Not Python?
- ❌ Slower performance
- ❌ GIL limitations
- ❌ Less suitable for high-concurrency

### Why Not SQLite + SpatiaLite?
- ❌ Limited concurrent writes
- ❌ Less spatial functions
- ❌ Not production-grade for this scale

### Why Not MySQL?
- ❌ No native spatial types (MySQL has basic support)
- ❌ Less spatial functions
- ❌ PostGIS is gold standard

---

[← Previous: API Design](./06-api-design.md) | [Back to Index](./README.md) | [Next: Implementation Phases →](./08-implementation-phases.md)
