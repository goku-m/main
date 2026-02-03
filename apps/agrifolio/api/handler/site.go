package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goku-m/main/apps/agrifolio/api/model"
	"github.com/goku-m/main/apps/agrifolio/api/model/site"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/goku-m/main/internal/shared/render"
	"github.com/google/uuid"

	"github.com/goku-m/main/apps/agrifolio/api/service"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/labstack/echo/v4"
)

type SiteHandler struct {
	Handler
	siteService *service.SiteService
}

func NewSiteHandler(s *server.Server, siteService *service.SiteService) *SiteHandler {
	return &SiteHandler{
		Handler:     NewHandler(s),
		siteService: siteService,
	}
}

//PAGE HANDLERS

func (h *SiteHandler) GetSitePage(c echo.Context) error {
	// 1) Guard service nil (super common during wiring)
	if h.siteService == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "siteService is nil")
	}

	query := &site.GetSitesQuery{}
	if err := c.Bind(query); err != nil {
		return err
	}

	sites, err := h.siteService.GetSites(c, query)
	if err != nil {
		return err
	}

	// // 2) Avoid panicking when there are no sites
	// var firstTitle string
	// if sites != nil && len(sites.Data) > 0 {
	// 	firstTitle = sites.Data[0].Title
	// }

	td := &render.TemplateData{
		Data: map[string]interface{}{
			"sites": sites.Data, // could be "" when none exist
			// or pass the whole list:
			// "sites": sites.Data,
		},
	}

	if err := c.Render(http.StatusOK, "home", td); err != nil {
		c.Logger().Error("SitePage render error: ", err)
		return err
	}

	return nil
}

func (h *SiteHandler) CreateSitePage(c echo.Context) error {

	if err := c.Render(http.StatusOK, "addTodo", nil); err != nil {
		c.Logger().Error("SitePage render error: ", err)
		return err
	}

	return nil
}

func (h *SiteHandler) UpdateSitePage(c echo.Context) error {
	userID := middleware.GetUserID(c)
	idParam := c.Param("id")
	fmt.Println(idParam)
	siteID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid site id")
	}

	t, err := h.siteService.GetSiteByID(c, userID, siteID) // you need this service method
	if err != nil {
		return err
	}

	td := &render.TemplateData{
		Data: map[string]interface{}{
			"site": t,
		},
	}

	if err := c.Render(http.StatusOK, "updateTodo", td); err != nil {
		c.Logger().Error("SitePage render error: ", err)
		return err
	}

	return nil
}

func (h *SiteHandler) CreateSite(c echo.Context) error {
	userID := middleware.GetUserID(c)

	title := c.FormValue("title")
	description := c.FormValue("description")
	priority := c.FormValue("priority")

	if strings.TrimSpace(title) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	payload := &site.CreateSitePayload{
		Title:       title,
		Description: &description, // only if your struct uses *string
	}

	// only set priority if provided (depends on your types)
	if strings.TrimSpace(priority) != "" {
		p := site.Priority(priority)
		payload.Priority = &p
	}

	if _, err := h.siteService.CreateSite(c, userID, payload); err != nil {
		return err
	}

	// Redirect back to list (refresh)
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *SiteHandler) UpdateSite(c echo.Context) error {
	userID := middleware.GetUserID(c)

	// 1) Get ID from route param
	idParam := c.Param("id")
	siteID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid site id")
	}

	// 2) Read form values
	title := strings.TrimSpace(c.FormValue("title"))
	description := strings.TrimSpace(c.FormValue("description"))
	priorityStr := strings.TrimSpace(c.FormValue("priority"))
	statusStr := strings.TrimSpace(c.FormValue("status"))

	// For update, you can decide whether title is required.
	// If you want to require it on the form:
	if title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	// 3) Build payload (NOTE: pointer fields!)
	payload := &site.UpdateSitePayload{
		ID: siteID,
	}

	// Set Title (payload.Title is *string)
	payload.Title = &title

	// Set Description only if provided (or always set if you want to allow clearing)
	if description != "" {
		payload.Description = &description
	}

	if statusStr != "" {
		s := site.Status(statusStr)
		payload.Status = &s
	}

	// Set Priority if provided (payload.Priority is *site.Priority)
	if priorityStr != "" {
		p := site.Priority(priorityStr)
		switch p {
		case site.PriorityLow, site.PriorityMedium, site.PriorityHigh:
			payload.Priority = &p
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid priority")
		}
	}

	// 4) Update
	if _, err := h.siteService.UpdateSite(c, userID, payload); err != nil {
		return err
	}

	// 5) Redirect back (refresh)
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *SiteHandler) DeleteSite(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.FormValue("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing id")
	}

	siteID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.siteService.DeleteSite(c, userID, siteID); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

//API HANDLERS

func (h *SiteHandler) CreateSiteAPI(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *site.CreateSitePayload) (*site.Site, error) {
			userID := middleware.GetUserID(c)
			return h.siteService.CreateSite(c, userID, payload)
		},
		http.StatusCreated,
		&site.CreateSitePayload{},
	)(c)
}

func (h *SiteHandler) GetSiteByIdAPI(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *site.GetSiteByIDPayload) (*site.PopulatedSite, error) {
			userID := middleware.GetUserID(c)
			return h.siteService.GetSiteByID(c, userID, payload.ID)
		},
		http.StatusOK,
		&site.GetSiteByIDPayload{},
	)(c)
}

func (h *SiteHandler) GetSitesAPI(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, query *site.GetSitesQuery) (*model.PaginatedResponse[site.PopulatedSite], error) {
			// userID := middleware.GetUserID(c)
			return h.siteService.GetSites(c, query)
		},
		http.StatusOK,
		&site.GetSitesQuery{},
	)(c)
}

func (h *SiteHandler) UpdateSiteAPI(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *site.UpdateSitePayload) (*site.Site, error) {
			userID := middleware.GetUserID(c)
			return h.siteService.UpdateSite(c, userID, payload)
		},
		http.StatusOK,
		&site.UpdateSitePayload{},
	)(c)
}

func (h *SiteHandler) DeleteSiteAPI(c echo.Context) error {
	return HandleNoContent(
		h.Handler,
		func(c echo.Context, payload *site.DeleteSitePayload) error {
			userID := middleware.GetUserID(c)
			return h.siteService.DeleteSite(c, userID, payload.ID)
		},
		http.StatusNoContent,
		&site.DeleteSitePayload{},
	)(c)
}
