# Deployment

**Document:** 10  
**Status:** Planning

---

## Deployment Options

### Option A: Single Server (Simple)

```
┌─────────────────────────┐
│  VPS / Cloud Server     │
│  ├─ Docker              │
│  │  ├─ API Container    │
│  │  ├─ PostgreSQL       │
│  │  └─ Redis (opt)      │
│  └─ Nginx (reverse proxy)│
└─────────────────────────┘
```

**Providers:**
- DigitalOcean ($12-48/bulan)
- Vultr ($10-40/bulan)
- AWS EC2 ($15-50/bulan)
- GCP Compute ($12-45/bulan)

**Recommended Specs:**
```
CPU: 2 vCPU minimum
RAM: 4 GB minimum
Storage: 50 GB SSD
OS: Ubuntu 22.04 LTS
```

---

### Option B: Managed Services

```
┌──────────┐  ┌──────────────┐  ┌──────────┐
│ API      │  │ Managed DB   │  │ Managed  │
│ (VPS)    │──│ (RDS/Cloud   │──│ Redis    │
│          │  │  SQL)        │  │          │
└──────────┘  └──────────────┘  └──────────┘
```

**Providers:**
- AWS RDS PostgreSQL
- Google Cloud SQL
- DigitalOcean Managed DB

**Pros:**
- ✅ Automated backups
- ✅ High availability
- ✅ Less maintenance

**Cons:**
- ❌ Higher cost
- ❌ Less control

---

### Option C: Kubernetes (Advanced)

```
┌──────────────┐
│ K8s Cluster  │
│ ├─ API Pods  │
│ ├─ DB Pods   │
│ └─ Services  │
└──────────────┘
```

**Use case:** Large scale, high availability

---

## Docker Deployment

### Dockerfile
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o main .

# Runtime stage
FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
EXPOSE 8080
CMD ["./main"]
```

### docker-compose.yml (Production)
```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=geomapping_id
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=redis
    depends_on:
      postgres:
        condition: service_healthy
    restart: always
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M

  postgres:
    image: postgis/postgis:15-3.3
    environment:
      POSTGRES_DB: geomapping_id
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: always
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G

  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redisdata:/data
    restart: always

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - api
    restart: always

volumes:
  pgdata:
  redisdata:
```

---

## Nginx Configuration

```nginx
events {
    worker_connections 1024;
}

http {
    upstream api {
        server api:8080;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=100r/m;

    server {
        listen 443 ssl;
        server_name api.example.com;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        location /api/ {
            limit_req zone=api burst=20;
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        location /health {
            proxy_pass http://api/health;
        }
    }

    server {
        listen 80;
        server_name api.example.com;
        return 301 https://$server_name$request_uri;
    }
}
```

---

## Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=geomapping_id
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Server
SERVER_PORT=8080
SERVER_ENV=production

# Logging
LOG_LEVEL=info

# Cache
CACHE_ENABLED=true
CACHE_TTL=24h
```

---

## Backup Strategy

### Database Backup
```bash
# Daily backup (cron)
0 2 * * * pg_dump -U postgres geomapping_id | gzip > /backups/db_$(date +\%Y\%m\%d).sql.gz

# Restore
gunzip -c backup.sql.gz | psql -U postgres geomapping_id
```

### WAL Archiving (Point-in-time Recovery)
```
postgresql.conf:
  wal_level = replica
  archive_mode = on
  archive_command = 'cp %p /archive/%f'
```

---

## Monitoring

### Health Check Endpoint
```go
func healthCheck(c *gin.Context) {
    // Check DB connection
    err := db.Ping()
    if err != nil {
        c.JSON(503, gin.H{"status": "unhealthy"})
        return
    }
    
    c.JSON(200, gin.H{"status": "healthy"})
}
```

### Metrics
- Request count
- Response time
- Error rate
- DB connection pool usage
- Cache hit rate

---

## SSL/TLS

### Let's Encrypt (Free)
```bash
certbot certonly --standalone -d api.example.com
```

### Auto-renewal
```bash
0 0 1 * * certbot renew --quiet
```

---

## Deployment Checklist

- [ ] Server provisioned & secured
- [ ] Docker installed
- [ ] SSL certificate obtained
- [ ] Database backup configured
- [ ] Monitoring setup
- [ ] Firewall rules configured
- [ ] Environment variables set
- [ ] Data imported & validated
- [ ] API tested
- [ ] Documentation updated

---

[← Previous: Testing](./09-testing.md) | [Back to Index](./README.md) | [Next: Risks →](./11-risks.md)
