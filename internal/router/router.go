package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mahdyarief/geomap-indonesia/internal/config"
	"github.com/mahdyarief/geomap-indonesia/internal/handler"
	"github.com/mahdyarief/geomap-indonesia/internal/middleware"
	"github.com/mahdyarief/geomap-indonesia/internal/repository"
	"github.com/mahdyarief/geomap-indonesia/internal/service"
)

// New builds the Gin engine with all routes wired.
func New(cfg *config.Config, pool *pgxpool.Pool) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	wilayahRepo := repository.NewWilayahRepository(pool)
	reverseRepo := repository.NewReverseRepository(pool)
	kodeposRepo := repository.NewKodeposRepository(pool)
	boundaryRepo := repository.NewBoundaryRepository(pool)

	authSvc := service.NewAuthService(cfg)
	wilayahSvc := service.NewWilayahService(wilayahRepo)
	reverseSvc := service.NewReverseService(reverseRepo)
	searchSvc := service.NewSearchService(reverseRepo)
	kodeposSvc := service.NewKodeposService(kodeposRepo)
	boundarySvc := service.NewBoundaryService(boundaryRepo)

	authH := handler.NewAuthHandler(authSvc)
	wilayahH := handler.NewWilayahHandler(wilayahSvc)
	reverseH := handler.NewReverseHandler(reverseSvc)
	searchH := handler.NewSearchHandler(searchSvc)
	kodeposH := handler.NewKodeposHandler(kodeposSvc)
	boundaryH := handler.NewBoundaryHandler(boundarySvc)
	healthH := handler.NewHealthHandler(pool)

	limiter := middleware.NewRateLimiter(cfg.RateLimitPerHour)
	limiter.Start()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	v1 := r.Group("/api/v1")

	v1.POST("/auth", limiter.Limit(), authH.Authenticate)
	v1.GET("/health", limiter.Limit(), healthH.Check)

	auth := v1.Group("")
	auth.Use(limiter.Limit(), middleware.Auth(cfg.JWTSecret))
	{
		auth.GET("/wilayah", wilayahH.List)
		auth.GET("/wilayah/:kode", wilayahH.Detail)
		auth.GET("/wilayah/:kode/children", wilayahH.Children)
		auth.GET("/reverse", reverseH.Reverse)
		auth.POST("/batch/reverse", reverseH.BatchReverse)
		auth.GET("/search", searchH.Search)
		auth.GET("/kodepos/:kode", kodeposH.Lookup)
		auth.GET("/kodepos", kodeposH.ByWilayah)
		auth.GET("/boundaries/:kode", boundaryH.GetGeoJSON)
	}

	return r
}
