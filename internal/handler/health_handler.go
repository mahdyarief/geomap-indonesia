package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// GET /health
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	dbStatus := "up"
	code := http.StatusOK
	if err := h.pool.Ping(ctx); err != nil {
		status = "degraded"
		dbStatus = "down"
		code = http.StatusServiceUnavailable
	}
	c.JSON(code, gin.H{
		"success":  true,
		"status":   status,
		"database": dbStatus,
		"time":     time.Now().Format(time.RFC3339),
	})
}
