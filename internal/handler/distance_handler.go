package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
	"github.com/mahdyarief/geomap-indonesia/internal/service"
)

// DistanceHandler serves the road-network distance endpoint.
type DistanceHandler struct {
	svc *service.DistanceService
}

// NewDistanceHandler creates a DistanceHandler.
func NewDistanceHandler(svc *service.DistanceService) *DistanceHandler {
	return &DistanceHandler{svc: svc}
}

// Indonesia bounding box for sane coordinate validation.
const (
	indonesiaMinLat = -11.0
	indonesiaMaxLat = 6.0
	indonesiaMinLng = 95.0
	indonesiaMaxLng = 141.0
)

func validCoord(lat, lng float64) bool {
	return lat >= indonesiaMinLat && lat <= indonesiaMaxLat &&
		lng >= indonesiaMinLng && lng <= indonesiaMaxLng
}

// Calculate handles POST /distance.
func (h *DistanceHandler) Calculate(c *gin.Context) {
	var req struct {
		Origin              *models.Centroid `json:"origin"`
		Destination         *models.Centroid `json:"destination"`
		AllowedWilayaCodes []string          `json:"allowed_wilayah_codes,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		failure(c, http.StatusBadRequest, codeBadRequest, "invalid request body")
		return
	}
	if req.Origin == nil || req.Destination == nil {
		failure(c, http.StatusBadRequest, codeBadRequest, "origin and destination are required")
		return
	}
	if !validCoord(req.Origin.Lat, req.Origin.Lng) || !validCoord(req.Destination.Lat, req.Destination.Lng) {
		failure(c, http.StatusBadRequest, codeBadRequest, "coordinates must be within Indonesia bounds")
		return
	}
	result, err := h.svc.Calculate(c.Request.Context(), *req.Origin, *req.Destination, req.AllowedWilayaCodes)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOutsideAllowedRegions):
			failure(c, http.StatusBadRequest, codeBadRequest, "coordinates must be within the allowed wilayah")
			return
		case errors.Is(err, service.ErrInvalidRegionCodes):
			failure(c, http.StatusBadRequest, codeBadRequest, "allowed_wilayah_codes contains unknown wilayah kode")
			return
		case errors.Is(err, repository.ErrNoRoute):
			failure(c, http.StatusNotFound, codeNotFound, "no road route between the given points")
			return
		}
		failure(c, http.StatusInternalServerError, codeInternalError, "distance calculation failed")
		return
	}
	success(c, http.StatusOK, result)
}