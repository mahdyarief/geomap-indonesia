# Architecture Design

**Document:** 03  
**Status:** Planning

---

## System Architecture Overview

```
┌─────────────────────────────────────────┐
│         Client Applications             │
│  (Mobile, Web, Third-party)             │
└──────────────┬──────────────────────────┘
               │ HTTPS
               ▼
┌─────────────────────────────────────────┐
│      Load Balancer (Optional)           │
│  (Nginx / Traefik / Cloud LB)           │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│    Geomapping API (Golang)              │
│  ┌─────────────────────────────────┐   │
│  │  Handler Layer (HTTP)           │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  Service Layer (Business)       │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  Repository Layer (Data)        │   │
│  └─────────────────────────────────┘   │
└──────────────┬──────────────────────────┘
               │
       ┌───────┴───────┐
       ▼               ▼
┌─────────────┐  ┌─────────────┐
│ PostgreSQL  │  │ Redis       │
│ + PostGIS   │  │ (Optional)  │
└─────────────┘  └─────────────┘
       │
       ▼
┌─────────────┐
│ Static      │
│ Assets      │
│ (Logos)     │
└─────────────┘
```

---

## Components

### 1. API Service (Golang)

**Responsibilities:**
- Handle HTTP requests
- Validate input
- Execute business logic
- Query database
- Format responses
- Error handling
- Logging

**Architecture Pattern:** Clean Architecture

**Layers:**
```
┌─────────────────────────────────────┐
│ Handler Layer                       │
│ - HTTP request/response             │
│ - Input validation                  │
│ - Response formatting               │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│ Service Layer                       │
│ - Business logic                    │
│ - Orchestration                     │
│ - Cache handling                    │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│ Repository Layer                    │
│ - Database queries                  │
│ - Spatial queries                   │
│ - Data access                       │
└─────────────────────────────────────┘
```

**Key Features:**
- Connection pooling
- Request validation
- CORS support
- Health checks
- Graceful shutdown

---

### 2. PostgreSQL + PostGIS

**Responsibilities:**
- Store master data
- Store geometry data
- Execute spatial queries
- Maintain data integrity
- Handle concurrent access

**Extensions:**
- **PostGIS:** Spatial types, functions, indexes
- **pg_trgm:** Fuzzy search (optional)

**Optimization:**
- Spatial indexes (GiST)
- B-tree indexes
- Connection pooling (PgBouncer)

---

### 3. Redis Cache (Optional)

**When to Use:**
- High traffic (> 1000 req/s)
- Expensive queries
- Read-heavy workloads

**Cache Strategy:**
- Grid-based caching for reverse geocoding
- TTL: 24 hours
- Cache invalidation on updates

---

### 4. Static Asset Storage

**Options:**

| Option | Pros | Cons | Use Case |
|--------|------|------|----------|
| Local FS | Simple | Not scalable | Development |
| MinIO | S3-compatible, self-hosted | Extra infra | Medium-large |
| Cloud S3 | Managed, CDN | Cost, lock-in | Production |

**Recommendation:** Start with local FS, migrate later

---

## Data Flow

### Reverse Geocoding Flow

```
1. Client Request
   GET /reverse?lat=-6.18&lng=106.82
              │
              ▼
2. API Handler
   - Validate lat/lng
   - Extract params
              │
              ▼
3. Service Layer
   - Check cache
   - If miss → query DB
              │
              ▼
4. Repository Layer
   - Spatial query:
     ST_Contains(geometry, point)
              │
              ▼
5. PostgreSQL
   - Use GiST index
   - Find polygon
   - Return result
              │
              ▼
6. Service Layer
   - Build hierarchy
   - Fetch related data
              │
              ▼
7. API Handler
   - Format JSON
   - Return response
```

---

## Scalability

### Vertical Scaling
- Increase CPU/RAM
- Optimize queries
- Add indexes

### Horizontal Scaling
- Load balancer
- Multiple API instances
- Read replicas
- Stateless design

### Database Scaling
- Connection pooling
- Read replicas
- Partitioning (if needed)
- Materialized views

---

## Security

### API Security
- HTTPS only
- Rate limiting
- Input validation
- SQL injection prevention

### Database Security
- Strong passwords
- Limited privileges
- Network isolation
- Regular backups

### Data Security
- No sensitive data (public admin data)
- CORS configuration
- Optional API auth

---

## Deployment Options

### Option A: Single Server (Simple)
```
┌─────────────────────┐
│  Server             │
│  ├─ API Service     │
│  ├─ PostgreSQL      │
│  └─ Redis (opt)     │
└─────────────────────┘
```
**Use case:** Development, small production

### Option B: Separated Services
```
┌──────────┐  ┌──────────┐  ┌──────────┐
│ API      │  │ Database │  │ Redis    │
│ Server   │──│ Server   │──│ Server   │
└──────────┘  └──────────┘  └──────────┘
```
**Use case:** Medium production

### Option C: Cluster (Advanced)
```
┌──────────┐
│ LB       │
└────┬─────┘
     │
┌────┴─────┐
│ API × N  │
└────┬─────┘
     │
┌────┴─────┐
│ DB Master│
│ + Replicas│
└──────────┘
```
**Use case:** Large scale production

---

## Technology Choices

| Component | Choice | Reason |
|-----------|--------|--------|
| Language | Golang | Performance, simplicity |
| Database | PostgreSQL + PostGIS | Gold standard for spatial |
| Cache | Redis (optional) | High performance |
| Container | Docker | Portability |
| API Style | REST | Universal, simple |

---

[← Previous: Data Sources](./02-data-sources.md) | [Back to Index](./README.md) | [Next: Database Schema →](./04-database-schema.md)
