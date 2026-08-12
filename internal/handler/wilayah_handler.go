package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
	"github.com/mahdyarief/geomap-indonesia/internal/service"
)

type WilayahHandler struct {
	svc *service.WilayahService
}

func NewWilayahHandler(svc *service.WilayahService) *WilayahHandler {
	return &WilayahHandler{svc: svc}
}

// GET /wilayah
func (h *WilayahHandler) List(c *gin.Context) {
	tipe := models.WilayahType(strings.ToLower(c.DefaultQuery("type", "provinsi")))
	if !validType(tipe) {
		failure(c, http.StatusBadRequest, codeBadRequest, "invalid type, must be one of: provinsi, kabupaten, kecamatan, kelurahan")
		return
	}
	parentID := c.Query("parent_id")
	search := c.Query("search")
	page := atoiDefault(c.DefaultQuery("page", "1"), 1)
	limit := atoiDefault(c.DefaultQuery("limit", "10"), 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	items, pagination, err := h.svc.List(c.Request.Context(), tipe, parentID, search, page, limit)
	if err != nil {
		failure(c, http.StatusInternalServerError, codeInternalError, "failed to list wilayah")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       items,
		"pagination": pagination,
	})
}

// GET /wilayah/:kode
func (h *WilayahHandler) Detail(c *gin.Context) {
	kode := c.Param("kode")
	detail, err := h.svc.Detail(c.Request.Context(), kode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			failure(c, http.StatusNotFound, codeNotFound, "wilayah not found")
			return
		}
		failure(c, http.StatusInternalServerError, codeInternalError, "failed to get wilayah detail")
		return
	}
	success(c, http.StatusOK, detail)
}

// GET /wilayah/:kode/children
func (h *WilayahHandler) Children(c *gin.Context) {
	kode := c.Param("kode")
	items, err := h.svc.Children(c.Request.Context(), kode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			failure(c, http.StatusNotFound, codeNotFound, "wilayah not found")
			return
		}
		failure(c, http.StatusInternalServerError, codeInternalError, "failed to get children")
		return
	}
	success(c, http.StatusOK, items)
}
