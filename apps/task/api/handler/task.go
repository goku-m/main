package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goku-m/main/apps/task/api/model/task"
	"github.com/goku-m/main/apps/task/ui/pages"
	"github.com/google/uuid"

	"github.com/goku-m/main/apps/task/api/service"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/labstack/echo/v4"
)

type TaskHandler struct {
	Handler
	taskService *service.TaskService
}

func NewTaskHandler(s *server.Server, taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		Handler:     NewHandler(s),
		taskService: taskService,
	}
}

func (h *TaskHandler) GetTaskPage(c echo.Context) error {
	if h.taskService == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "taskService is nil")
	}

	query := &task.GetTasksQuery{}
	if err := c.Bind(query); err != nil {
		return err
	}

	tasks, err := h.taskService.GetTasks(c, query)
	if err != nil {
		return err
	}

	// Map your domain users -> view model (keep templ clean & stable)
	view := make([]pages.TaskView, 0, len(tasks.Data))
	for _, t := range tasks.Data {

		desc := ""
		if t.Description != nil {
			desc = *t.Description
		}

		view = append(view, pages.TaskView{
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

func (h *TaskHandler) CreateTaskPage(c echo.Context) error {

	return pages.CreateTask().Render(
		c.Request().Context(),
		c.Response(),
	)

}

func (h *TaskHandler) UpdateTaskPage(c echo.Context) error {
	idParam := c.Param("id")
	fmt.Println(idParam)
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	t, err := h.taskService.GetTaskByID(c, taskID) // you need this service method
	if err != nil {
		return err
	}

	desc := ""
	if t.Description != nil {
		desc = *t.Description
	}

	view := pages.TaskView{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: desc,               // if pointer
		Priority:    string(t.Priority), // adjust types
		Status:      string(t.Status),
	}

	return pages.EditTask(view).Render(c.Request().Context(), c.Response())

}

func (h *TaskHandler) CreateTask(c echo.Context) error {
	// taskID := middleware.GetTaskID(c)

	title := c.FormValue("title")
	description := c.FormValue("description")
	priority := c.FormValue("priority")

	if strings.TrimSpace(title) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	payload := &task.CreateTaskPayload{
		Title:       title,
		Description: &description,
	}

	if strings.TrimSpace(priority) != "" {
		p := task.Priority(priority)
		payload.Priority = &p
	}

	if _, err := h.taskService.CreateTask(c, payload); err != nil {
		return err
	}

	// Redirect back to list (refresh)
	return c.Redirect(http.StatusSeeOther, "/task")
}

func (h *TaskHandler) UpdateTask(c echo.Context) error {

	idParam := c.Param("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
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
	payload := &task.UpdateTaskPayload{
		ID: taskID,
	}

	// Set Title (payload.Title is *string)
	payload.Title = &title

	// Set Description only if provided (or always set if you want to allow clearing)
	if description != "" {
		payload.Description = &description
	}

	if statusStr != "" {
		s := task.Status(statusStr)
		payload.Status = &s
	}

	if priorityStr != "" {
		p := task.Priority(priorityStr)
		switch p {
		case task.PriorityLow, task.PriorityMedium, task.PriorityHigh:
			payload.Priority = &p
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid priority")
		}
	}

	// 4) Update
	if _, err := h.taskService.UpdateTask(c, payload); err != nil {
		return err
	}

	// 5) Redirect back (refresh)
	return c.Redirect(http.StatusSeeOther, "/task")
}

func (h *TaskHandler) DeleteTask(c echo.Context) error {
	// taskID := middleware.GetTaskID(c)
	id := c.FormValue("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing id")
	}

	taskID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.taskService.DeleteTask(c, taskID); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/task")
}
