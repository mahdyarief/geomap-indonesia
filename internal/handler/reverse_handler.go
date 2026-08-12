package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
	"github.com/mahdyarief/geomap-indonesia/internal/service"
)

const maxBatchPoints = 100

type ReverseHandler struct {
	svc *service.ReverseService
}

func NewReverseHandler(svc *service.ReverseService) *ReverseHandler {
	return &ReverseHandler{svc: svc}
}

// GET /reverse?lat=..&lng=..
func (h *ReverseHandler) Reverse(c *gin.Context) {
	lat, okLat := parseFloatQuery(c, "lat")
	lng, okLng := parseFloatQuery(c, "lng")
	if !okLat || !okLng {
		failure(c, http.StatusBadRequest, codeBadRequest, "lat and lng are required numeric parameters")
		return
	}
	if lat < -11 || lat > 6 || lng < 95 || lng > 141 {
		failure(c, http.StatusBadRequest, codeBadRequest, "coordinates out of Indonesia bounds")
		return
	}

	res, err := h.svc.Reverse(c.Request.Context(), lat, lng)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			failure(c, http.StatusNotFound, codeNotFound, "no wilayah found for coordinates")
			return
		}
		failure(c, http.StatusInternalServerError, codeInternalError, "reverse geocoding failed")
		return
	}
	success(c, http.StatusOK, res)
}

// POST /batch/reverse
func (h *ReverseHandler) BatchReverse(c *gin.Context) {
	var req models.BatchReverseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failure(c, http.StatusBadRequest, codeBadRequest, "invalid request body")
		return
	}
	if len(req.Points) == 0 {
		failure(c, http.StatusBadRequest, codeBadRequest, "points must not be empty")
		return
	}
	if len(req.Points) > maxBatchPoints {
		failure(c, http.StatusBadRequest, codeBadRequest, "max 100 points per request")
		return
	}
	results := h.svc.BatchReverse(c.Request.Context(), req.Points)
	success(c, http.StatusOK, results)
}
