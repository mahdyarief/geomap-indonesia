package service

import (
	"context"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
)

// SearchService implements advanced name search across all wilayah levels.
type SearchService struct {
	repo *repository.ReverseRepository
}

func NewSearchService(repo *repository.ReverseRepository) *SearchService {
	return &SearchService{repo: repo}
}

// Search runs a fuzzy name search, optionally restricted to one level.
func (s *SearchService) Search(ctx context.Context, q, tipe string, limit int) ([]models.SearchResult, error) {
	if q == "" {
		return []models.SearchResult{}, nil
	}
	return s.repo.Search(ctx, q, tipe, limit)
}
