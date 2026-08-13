package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mahdyarief/geomap-indonesia/internal/repository"
	"github.com/mahdyarief/geomap-indonesia/internal/service"
)

type BoundaryHandler struct {
	svc *service.BoundaryService
}

func NewBoundaryHandler(svc *service.BoundaryService) *BoundaryHandler {
	return &BoundaryHandler{svc: svc}
}

// GET /boundaries/:kode
func (h *BoundaryHandler) GetGeoJSON(c *gin.Context) {
	kode := c.Param("kode")
	feature, err := h.svc.GetGeoJSON(c.Request.Context(), kode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			failure(c, http.StatusNotFound, codeNotFound, "boundary not found")
			return
		}
		failure(c, http.StatusInternalServerError, codeInternalError, "boundary retrieval failed")
		return
	}
	success(c, http.StatusOK, feature)
}
