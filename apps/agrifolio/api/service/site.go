package service

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	//"github.com/goku-m/mains/internal/shared/lib/aws"
	"github.com/goku-m/main/apps/agrifolio/api/model"
	"github.com/goku-m/main/apps/agrifolio/api/model/site"
	"github.com/goku-m/main/apps/agrifolio/api/repository"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/goku-m/main/internal/shared/server"
)

type SiteService struct {
	server   *server.Server
	siteRepo *repository.SiteRepository
}

func NewSiteService(server *server.Server, siteRepo *repository.SiteRepository,
) *SiteService {
	return &SiteService{
		server:   server,
		siteRepo: siteRepo,
	}
}

func (s *SiteService) CreateSite(ctx echo.Context, userID string, payload *site.CreateSitePayload) (*site.Site, error) {
	logger := middleware.GetLogger(ctx)

	// Validate parent site exists and belongs to user (if provided)

	siteItem, err := s.siteRepo.CreateSite(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create site")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "site_created").
		Str("site_id", siteItem.ID.String()).
		Str("title", siteItem.Title).
		Str("priority", string(siteItem.Priority)).
		Msg("Site created successfully")

	return siteItem, nil
}

func (s *SiteService) GetSiteByID(ctx echo.Context, userID string, siteID uuid.UUID) (*site.PopulatedSite, error) {
	logger := middleware.GetLogger(ctx)

	siteItem, err := s.siteRepo.GetSiteByID(ctx.Request().Context(), userID, siteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch site by ID")
		return nil, err
	}

	return siteItem, nil
}

func (s *SiteService) GetSites(ctx echo.Context, query *site.GetSitesQuery) (*model.PaginatedResponse[site.PopulatedSite], error) {
	logger := middleware.GetLogger(ctx)

	result, err := s.siteRepo.GetSites(ctx.Request().Context(), query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch sites")
		return nil, err
	}

	return result, nil
}

func (s *SiteService) UpdateSite(ctx echo.Context, userID string, payload *site.UpdateSitePayload) (*site.Site, error) {
	logger := middleware.GetLogger(ctx)

	updatedSite, err := s.siteRepo.UpdateSite(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update site")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "site_updated").
		Str("site_id", updatedSite.ID.String()).
		Str("title", updatedSite.Title).
		Str("priority", string(updatedSite.Priority)).
		Str("status", string(updatedSite.Status)).
		Msg("Site updated successfully")

	return updatedSite, nil
}

func (s *SiteService) DeleteSite(ctx echo.Context, userID string, siteID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	err := s.siteRepo.DeleteSite(ctx.Request().Context(), userID, siteID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete site")
		return err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "site_deleted").
		Str("site_id", siteID.String()).
		Msg("Site deleted successfully")

	return nil
}

func (s *SiteService) GetSiteStats(ctx echo.Context, userID string) (*site.SiteStats, error) {
	logger := middleware.GetLogger(ctx)

	stats, err := s.siteRepo.GetSiteStats(ctx.Request().Context(), userID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch site statistics")
		return nil, err
	}

	return stats, nil
}
