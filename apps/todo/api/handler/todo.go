package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goku-m/main/apps/todo/api/model/todo"
	"github.com/goku-m/main/apps/todo/ui/pages"
	"github.com/google/uuid"

	"github.com/goku-m/main/apps/todo/api/service"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/labstack/echo/v4"
)

type TodoHandler struct {
	Handler
	todoService *service.TodoService
}

func NewTodoHandler(s *server.Server, todoService *service.TodoService) *TodoHandler {
	return &TodoHandler{
		Handler:     NewHandler(s),
		todoService: todoService,
	}
}

func (h *TodoHandler) GetTodoPage(c echo.Context) error {
	if h.todoService == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "todoService is nil")
	}

	query := &todo.GetTodosQuery{}
	if err := c.Bind(query); err != nil {
		return err
	}

	todos, err := h.todoService.GetTodos(c, query)
	if err != nil {
		return err
	}

	// Map your domain users -> view model (keep templ clean & stable)
	view := make([]pages.TodoView, 0, len(todos.Data))
	for _, t := range todos.Data {

		desc := ""
		if t.Description != nil {
			desc = *t.Description
		}

		view = append(view, pages.TodoView{
			ID:          t.ID.String(),
			Title:       t.Title,
			Description: desc,               // if pointer
			Priority:    string(t.Priority), // adjust types
			Status:      string(t.Status),
		})
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return pages.Home(view).Render(c.Request().Context(), c.Response())
}

func (h *TodoHandler) CreateTodoPage(c echo.Context) error {

	return pages.CreateTodo().Render(
		c.Request().Context(),
		c.Response(),
	)

}

func (h *TodoHandler) UpdateTodoPage(c echo.Context) error {
	idParam := c.Param("id")
	fmt.Println(idParam)
	todoID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid todo id")
	}

	t, err := h.todoService.GetTodoByID(c, todoID) // you need this service method
	if err != nil {
		return err
	}

	desc := ""
	if t.Description != nil {
		desc = *t.Description
	}

	view := pages.TodoView{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: desc,               // if pointer
		Priority:    string(t.Priority), // adjust types
		Status:      string(t.Status),
	}

	return pages.EditTodo(view).Render(c.Request().Context(), c.Response())

}

func (h *TodoHandler) CreateTodo(c echo.Context) error {
	// todoID := middleware.GetTodoID(c)

	title := c.FormValue("title")
	description := c.FormValue("description")
	priority := c.FormValue("priority")

	if strings.TrimSpace(title) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	payload := &todo.CreateTodoPayload{
		Title:       title,
		Description: &description,
	}

	if strings.TrimSpace(priority) != "" {
		p := todo.Priority(priority)
		payload.Priority = &p
	}

	if _, err := h.todoService.CreateTodo(c, payload); err != nil {
		return err
	}

	// Redirect back to list (refresh)
	return c.Redirect(http.StatusSeeOther, "/todo")
}

func (h *TodoHandler) UpdateTodo(c echo.Context) error {

	idParam := c.Param("id")
	todoID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid todo id")
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
	payload := &todo.UpdateTodoPayload{
		ID: todoID,
	}

	// Set Title (payload.Title is *string)
	payload.Title = &title

	// Set Description only if provided (or always set if you want to allow clearing)
	if description != "" {
		payload.Description = &description
	}

	if statusStr != "" {
		s := todo.Status(statusStr)
		payload.Status = &s
	}

	if priorityStr != "" {
		p := todo.Priority(priorityStr)
		switch p {
		case todo.PriorityLow, todo.PriorityMedium, todo.PriorityHigh:
			payload.Priority = &p
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid priority")
		}
	}

	// 4) Update
	if _, err := h.todoService.UpdateTodo(c, payload); err != nil {
		return err
	}

	// 5) Redirect back (refresh)
	return c.Redirect(http.StatusSeeOther, "/todo")
}

func (h *TodoHandler) DeleteTodo(c echo.Context) error {
	// todoID := middleware.GetTodoID(c)
	id := c.FormValue("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing id")
	}

	todoID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.todoService.DeleteTodo(c, todoID); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/todo")
}
