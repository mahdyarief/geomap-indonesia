package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
)

const (
	codeBadRequest    = "BAD_REQUEST"
	codeUnauthorized  = "UNAUTHORIZED"
	codeNotFound      = "NOT_FOUND"
	codeRateLimit     = "RATE_LIMIT"
	codeInternalError = "INTERNAL_ERROR"
)

// ErrorBody is the standard error payload.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"success": true, "data": data})
}

func failure(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   ErrorBody{Code: code, Message: message},
	})
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}

func parseFloatQuery(c *gin.Context, key string) (float64, bool) {
	v := c.Query(key)
	if v == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func validType(t models.WilayahType) bool {
	switch t {
	case models.TypeProvinsi, models.TypeKabupaten, models.TypeKecamatan, models.TypeKelurahan:
		return true
	}
	return false
}
