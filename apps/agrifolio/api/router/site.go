package router

import (
	"github.com/goku-m/main/apps/agrifolio/api/handler"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

func registerSiteRoutes(r *echo.Group, h *handler.SiteHandler, auth *middleware.AuthMiddleware) {
	// Site operations
	sites := r.Group("/sites")
	sites.Use(auth.RequireAuthIP)

	// Collection operations for pages
	sites.POST("/create", h.CreateSite)
	sites.POST("/delete", h.DeleteSite)
	sites.POST("/update/:id", h.UpdateSite)

	// Collection operations for api
	sites.GET("", h.GetSitesAPI)
	//Individual site operations
	dynamicSite := sites.Group("/:id")
	dynamicSite.GET("", h.GetSiteByIdAPI)
	dynamicSite.PATCH("", h.UpdateSiteAPI)
	dynamicSite.DELETE("", h.DeleteSiteAPI)

}
