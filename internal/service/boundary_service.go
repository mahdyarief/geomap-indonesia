package service

import (
	"context"

	"github.com/mahdyarief/geomap-indonesia/internal/repository"
)

// BoundaryService implements GeoJSON boundary retrieval.
type BoundaryService struct {
	repo *repository.BoundaryRepository
}

func NewBoundaryService(repo *repository.BoundaryRepository) *BoundaryService {
	return &BoundaryService{repo: repo}
}

// GetGeoJSON returns the boundary of a wilayah as a GeoJSON Feature.
func (s *BoundaryService) GetGeoJSON(ctx context.Context, kode string) (*repository.GeoJSONFeature, error) {
	return s.repo.GetGeoJSON(ctx, kode)
}
