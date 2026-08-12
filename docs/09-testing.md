# Testing Strategy

**Document:** 09  
**Status:** Planning

---

## Testing Levels

```
┌─────────────────────────┐
│   E2E Tests             │  ← Full flow
├─────────────────────────┤
│   Integration Tests     │  ← API + DB
├─────────────────────────┤
│   Unit Tests            │  ← Functions
└─────────────────────────┘
```

---

## Unit Tests

**Scope:** Individual functions, no external dependencies

### Examples
```go
func TestCleanKode(t *testing.T) {
    tests := []struct{
        input    string
        expected string
    }{
        {"11.01.01.2001", "1101012001"},
        {"31.71", "3171"},
        {"11", "11"},
    }
    
    for _, tt := range tests {
        result := cleanKode(tt.input)
        assert.Equal(t, tt.expected, result)
    }
}

func TestDetectLevel(t *testing.T) {
    assert.Equal(t, "provinsi", detectLevel("11"))
    assert.Equal(t, "kabupaten", detectLevel("1101"))
    assert.Equal(t, "kecamatan", detectLevel("110101"))
    assert.Equal(t, "kelurahan", detectLevel("1101012001"))
}
```

---

## Integration Tests

**Scope:** API endpoints + database

### Setup
```go
func TestMain(m *testing.M) {
    // Setup test database
    db = setupTestDB()
    
    // Seed test data
    seedTestData(db)
    
    // Run tests
    code := m.Run()
    
    // Cleanup
    teardown(db)
    os.Exit(code)
}
```

### Examples
```go
func TestReverseGeocoding(t *testing.T) {
    req := httptest.NewRequest("GET", 
        "/api/v1/reverse?lat=-6.18&lng=106.82", nil)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    
    var resp ReverseResponse
    json.Unmarshal(w.Body.Bytes(), &resp)
    
    assert.Equal(t, "DKI Jakarta", resp.Data.Provinsi.Nama)
}

func TestWilayahLookup(t *testing.T) {
    req := httptest.NewRequest("GET", 
        "/api/v1/wilayah/3171010001", nil)
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

---

## Spatial Query Tests

**Scope:** Verify PostGIS queries work correctly

### Examples
```go
func TestSpatialQuery(t *testing.T) {
    // Test point in Jakarta
    result := reverseGeocode(-6.18, 106.82)
    assert.Equal(t, "31", result.ProvinsiKode)
    
    // Test point in Surabaya
    result = reverseGeocode(-7.25, 112.75)
    assert.Equal(t, "35", result.ProvinsiKode)
    
    // Test point in ocean (should return nil)
    result = reverseGeocode(-10.0, 120.0)
    assert.Nil(t, result)
}
```

---

## Performance Tests

**Scope:** Response time, throughput

### Tools
- `hey` - Load testing
- `wrk` - Benchmarking
- Go benchmark tests

### Examples
```go
func BenchmarkReverseGeocoding(b *testing.B) {
    for i := 0; i < b.N; i++ {
        reverseGeocode(-6.18, 106.82)
    }
}

// Run: go test -bench=. -benchmem
```

### Load Test
```bash
# 1000 requests, 50 concurrent
hey -n 1000 -c 50 \
  "http://localhost:8080/api/v1/reverse?lat=-6.18&lng=106.82"

# Expected: p95 < 50ms
```

---

## Data Validation Tests

**Scope:** Verify data integrity

### Examples
```go
func TestDataIntegrity(t *testing.T) {
    // Check record counts
    count := countProvinsi()
    assert.Equal(t, 34, count)
    
    count = countKabupaten()
    assert.Equal(t, 514, count)
    
    // Check hierarchy
    orphans := findOrphanedKabupaten()
    assert.Equal(t, 0, orphans)
    
    // Check geometry validity
    invalid := countInvalidGeometry()
    assert.Equal(t, 0, invalid)
}
```

---

## Test Coverage

**Target:** 80%+ coverage

### Run Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgis/postgis:15-3.3
        env:
          POSTGRES_DB: test_db
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test -v -cover ./...
```

---

[← Previous: Implementation Phases](./08-implementation-phases.md) | [Back to Index](./README.md) | [Next: Deployment →](./10-deployment.md)
