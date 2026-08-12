package service

import (
	"context"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
)

// WilayahService implements list/detail/children operations.
type WilayahService struct {
	repo *repository.WilayahRepository
}

func NewWilayahService(repo *repository.WilayahRepository) *WilayahService {
	return &WilayahService{repo: repo}
}

// List returns a paginated list of wilayah for a given type.
func (s *WilayahService) List(ctx context.Context, tipe models.WilayahType, parentID, search string, page, limit int) ([]models.WilayahListItem, models.Pagination, error) {
	items, total, err := s.repo.List(ctx, tipe, parentID, search, page, limit)
	if err != nil {
		return nil, models.Pagination{}, err
	}
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return items, models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// Detail returns the full detail for a wilayah kode.
func (s *WilayahService) Detail(ctx context.Context, kode string) (*models.WilayahDetail, error) {
	return s.repo.Detail(ctx, kode)
}

// Children returns the direct children of a wilayah kode.
func (s *WilayahService) Children(ctx context.Context, kode string) ([]models.WilayahListItem, error) {
	return s.repo.Children(ctx, kode)
}
