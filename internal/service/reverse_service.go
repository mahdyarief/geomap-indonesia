package service

import (
	"context"
	"errors"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
)

// ReverseService implements reverse geocoding, including batch mode.
type ReverseService struct {
	repo *repository.ReverseRepository
}

func NewReverseService(repo *repository.ReverseRepository) *ReverseService {
	return &ReverseService{repo: repo}
}

// Reverse geocodes a lat/lng point into the full administrative hierarchy,
// attaching the postal codes of the matched kelurahan when available.
func (s *ReverseService) Reverse(ctx context.Context, lat, lng float64) (*models.ReverseResult, error) {
	res, err := s.repo.ReverseGeocode(ctx, lat, lng)
	if err != nil {
		return nil, err
	}

	codes, err := s.repo.KodeposByKelurahan(ctx, res.Kelurahan.Kode)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
	} else {
		res.Kodepos = codes
	}
	return res, nil
}

// BatchReverse geocodes a list of points, collecting per-point results even
// when some points fail to resolve.
func (s *ReverseService) BatchReverse(ctx context.Context, points []models.Centroid) []models.BatchReverseResult {
	results := make([]models.BatchReverseResult, 0, len(points))
	for _, p := range points {
		item := models.BatchReverseResult{Input: p}
		res, err := s.Reverse(ctx, p.Lat, p.Lng)
		if err != nil {
			msg := err.Error()
			item.Error = &msg
		} else {
			item.Result = res
		}
		results = append(results, item)
	}
	return results
}
