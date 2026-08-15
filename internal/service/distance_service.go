package service

import (
	"context"
	"errors"

	"github.com/mahdyarief/geomap-indonesia/internal/models"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
)

// ErrOutsideAllowedRegions is returned when a route point falls outside the
// administrative boundaries supplied via allowed_wilayah_codes.
var ErrOutsideAllowedRegions = errors.New("point outside allowed wilayah regions")

// ErrInvalidRegionCodes is returned when none of the supplied allowed wilayah
// codes resolve to an existing boundary geometry.
var ErrInvalidRegionCodes = errors.New("none of the allowed wilayah codes are known")

// DistanceService computes road-network distances via pgRouting.
type DistanceService struct {
	repo         *repository.DistanceRepository
	boundaryRepo *repository.BoundaryRepository
}

// NewDistanceService creates a DistanceService.
func NewDistanceService(repo *repository.DistanceRepository, boundaryRepo *repository.BoundaryRepository) *DistanceService {
	return &DistanceService{repo: repo, boundaryRepo: boundaryRepo}
}

// Calculate returns the shortest road route between origin and destination.
// If allowedCodes is non-empty, both points must fall inside at least one of
// the allowed administrative regions, otherwise the request is rejected.
func (s *DistanceService) Calculate(ctx context.Context, origin, destination models.Centroid, allowedCodes []string) (*models.DistanceResult, error) {
	if len(allowedCodes) > 0 {
		if err := s.validateAllowed(ctx, origin, destination, allowedCodes); err != nil {
			return nil, err
		}
	}
	route, err := s.repo.Route(ctx, origin.Lat, origin.Lng, destination.Lat, destination.Lng)
	if err != nil {
		return nil, err
	}
	return &models.DistanceResult{
		DistanceKm:      route.DistanceKm,
		DurationMinutes: route.DurationMinutes,
		Geometry:        route.Geometry,
	}, nil
}

func (s *DistanceService) validateAllowed(ctx context.Context, origin, destination models.Centroid, allowedCodes []string) error {
	originWithin, originKnown, err := s.boundaryRepo.PointInRegions(ctx, origin.Lat, origin.Lng, allowedCodes)
	if err != nil {
		return err
	}
	if !originKnown {
		return ErrInvalidRegionCodes
	}
	if !originWithin {
		return ErrOutsideAllowedRegions
	}

	dstWithin, dstKnown, err := s.boundaryRepo.PointInRegions(ctx, destination.Lat, destination.Lng, allowedCodes)
	if err != nil {
		return err
	}
	if !dstKnown {
		return ErrInvalidRegionCodes
	}
	if !dstWithin {
		return ErrOutsideAllowedRegions
	}
	return nil
}