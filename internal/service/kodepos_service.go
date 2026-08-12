package service

import (
	"context"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
)

// KodeposService implements postal code lookups.
type KodeposService struct {
	repo *repository.KodeposRepository
}

func NewKodeposService(repo *repository.KodeposRepository) *KodeposService {
	return &KodeposService{repo: repo}
}

// Lookup resolves a postal code to its wilayah hierarchy.
func (s *KodeposService) Lookup(ctx context.Context, kodepos string) (*models.KodeposLookup, error) {
	return s.repo.Lookup(ctx, kodepos)
}

// ByWilayah returns all postal codes under a wilayah kode.
func (s *KodeposService) ByWilayah(ctx context.Context, wilayahKode string) (*models.KodeposByWilayah, error) {
	return s.repo.ByWilayah(ctx, wilayahKode)
}
