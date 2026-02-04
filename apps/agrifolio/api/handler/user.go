package handler

import (
	"fmt"
	"net/http"
	"strings"

	user "github.com/goku-m/main/apps/agrifolio/api/model/user"
	"github.com/goku-m/main/internal/shared/render"
	"github.com/google/uuid"

	"github.com/goku-m/main/apps/agrifolio/api/service"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	Handler
	userService *service.UserService
}

func NewUserHandler(s *server.Server, userService *service.UserService) *UserHandler {
	return &UserHandler{
		Handler:     NewHandler(s),
		userService: userService,
	}
}

//PAGE HANDLERS

func (h *UserHandler) GetUserPage(c echo.Context) error {
	// 1) Guard service nil (super common during wiring)
	if h.userService == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "userService is nil")
	}

	query := &user.GetUserQuery{}
	if err := c.Bind(query); err != nil {
		return err
	}

	users, err := h.userService.GetUsers(c, query)
	if err != nil {
		return err
	}

	td := &render.TemplateData{
		Data: map[string]interface{}{
			"users": users.Data, // could be "" when none exist
			// or pass the whole list:
			// "users": users.Data,
		},
	}

	if err := c.Render(http.StatusOK, "home", td); err != nil {
		c.Logger().Error("UserPage render error: ", err)
		return err
	}

	return nil
}

func (h *UserHandler) CreateUserPage(c echo.Context) error {

	if err := c.Render(http.StatusOK, "addTodo", nil); err != nil {
		c.Logger().Error("UserPage render error: ", err)
		return err
	}

	return nil
}

func (h *UserHandler) UpdateUserPage(c echo.Context) error {
	idParam := c.Param("id")
	fmt.Println(idParam)
	userID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	t, err := h.userService.GetUserByID(c, userID) // you need this service method
	if err != nil {
		return err
	}

	td := &render.TemplateData{
		Data: map[string]interface{}{
			"user": t,
		},
	}

	if err := c.Render(http.StatusOK, "updateTodo", td); err != nil {
		c.Logger().Error("UserPage render error: ", err)
		return err
	}

	return nil
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	// userID := middleware.GetUserID(c)

	title := c.FormValue("title")
	description := c.FormValue("description")

	if strings.TrimSpace(title) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	payload := &user.CreateUserPayload{
		Name:     title,
		Email:    &description,
		Password: &title, // only if your struct uses *string
	}

	if _, err := h.userService.CreateUser(c, payload); err != nil {
		return err
	}

	// Redirect back to list (refresh)
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) UpdateUser(c echo.Context) error {

	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	// 2) Read form values
	title := strings.TrimSpace(c.FormValue("title"))
	description := strings.TrimSpace(c.FormValue("description"))

	// For update, you can decide whether title is required.
	// If you want to require it on the form:
	if title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	// 3) Build payload (NOTE: pointer fields!)
	payload := &user.UpdateUserPayload{
		ID: userID,
	}

	// Set Title (payload.Title is *string)
	payload.Name = &title

	// Set Description only if provided (or always set if you want to allow clearing)
	if description != "" {
		payload.Password = &description
	}

	// 4) Update
	if _, err := h.userService.UpdateUser(c, payload); err != nil {
		return err
	}

	// 5) Redirect back (refresh)
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	// userID := middleware.GetUserID(c)
	id := c.FormValue("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing id")
	}

	userID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.userService.DeleteUser(c, userID); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

//API HANDLERS
