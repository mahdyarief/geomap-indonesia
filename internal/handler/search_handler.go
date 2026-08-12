package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mahdyarief/geomap-indonesia/internal/service"
)

type SearchHandler struct {
	svc *service.SearchService
}

func NewSearchHandler(svc *service.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// GET /search?q=..&type=..&limit=..
func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		failure(c, http.StatusBadRequest, codeBadRequest, "q is required")
		return
	}
	tipe := c.Query("type")
	limit := atoiDefault(c.DefaultQuery("limit", "10"), 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	results, err := h.svc.Search(c.Request.Context(), q, tipe, limit)
	if err != nil {
		failure(c, http.StatusInternalServerError, codeInternalError, "search failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"total":   len(results),
	})
}
