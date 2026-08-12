package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mahdyarief/geomap-indonesia/internal/repository"
	"github.com/mahdyarief/geomap-indonesia/internal/service"
)

type KodeposHandler struct {
	svc *service.KodeposService
}

func NewKodeposHandler(svc *service.KodeposService) *KodeposHandler {
	return &KodeposHandler{svc: svc}
}

// GET /kodepos/:kode
func (h *KodeposHandler) Lookup(c *gin.Context) {
	kode := c.Param("kode")
	res, err := h.svc.Lookup(c.Request.Context(), kode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			failure(c, http.StatusNotFound, codeNotFound, "kodepos not found")
			return
		}
		failure(c, http.StatusInternalServerError, codeInternalError, "kodepos lookup failed")
		return
	}
	success(c, http.StatusOK, res)
}

// GET /kodepos?wilayah=..
func (h *KodeposHandler) ByWilayah(c *gin.Context) {
	wilayah := c.Query("wilayah")
	if wilayah == "" {
		failure(c, http.StatusBadRequest, codeBadRequest, "wilayah query parameter is required")
		return
	}
	res, err := h.svc.ByWilayah(c.Request.Context(), wilayah)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			failure(c, http.StatusNotFound, codeNotFound, "wilayah not found")
			return
		}
		failure(c, http.StatusInternalServerError, codeInternalError, "kodepos lookup failed")
		return
	}
	success(c, http.StatusOK, res)
}
