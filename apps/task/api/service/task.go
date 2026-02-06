package service

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	//"github.com/goku-m/mains/internal/shared/lib/aws"

	"github.com/goku-m/main/apps/task/api/model"
	"github.com/goku-m/main/apps/task/api/model/task"
	"github.com/goku-m/main/apps/task/api/repository"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/goku-m/main/internal/shared/server"
)

type TaskService struct {
	server   *server.Server
	taskRepo *repository.TaskRepository
}

func NewTaskService(server *server.Server, taskRepo *repository.TaskRepository,
) *TaskService {
	return &TaskService{
		server:   server,
		taskRepo: taskRepo,
	}
}

func (s *TaskService) CreateTask(ctx echo.Context, payload *task.CreateTaskPayload) (*task.Task, error) {
	logger := middleware.GetLogger(ctx)

	// Validate parent task exists and belongs to task (if provided)

	taskItem, err := s.taskRepo.CreateTask(ctx.Request().Context(), payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create task")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "task_created").
		Str("task_id", taskItem.ID.String()).
		Str("title", taskItem.Title).
		Str("priority", string(taskItem.Priority)).
		Msg("Task created successfully")

	return taskItem, nil
}

func (s *TaskService) GetTaskByID(ctx echo.Context, taskID uuid.UUID) (*task.Task, error) {
	logger := middleware.GetLogger(ctx)

	taskItem, err := s.taskRepo.GetTaskByID(ctx.Request().Context(), taskID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch task by ID")
		return nil, err
	}

	return taskItem, nil
}

func (s *TaskService) GetTasks(ctx echo.Context, query *task.GetTasksQuery) (*model.PaginatedResponse[task.PopulatedTask], error) {
	logger := middleware.GetLogger(ctx)

	result, err := s.taskRepo.GetTasks(ctx.Request().Context(), query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch tasks")
		return nil, err
	}

	return result, nil
}

func (s *TaskService) UpdateTask(ctx echo.Context, payload *task.UpdateTaskPayload) (*task.Task, error) {
	logger := middleware.GetLogger(ctx)

	updatedTask, err := s.taskRepo.UpdateTask(ctx.Request().Context(), payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update task")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "task_updated").
		Str("task_id", updatedTask.ID.String()).
		Str("title", updatedTask.Title).
		Str("priority", string(updatedTask.Priority)).
		Str("status", string(updatedTask.Status)).
		Msg("Task updated successfully")

	return updatedTask, nil
}

func (s *TaskService) DeleteTask(ctx echo.Context, taskID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	err := s.taskRepo.DeleteTask(ctx.Request().Context(), taskID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete task")
		return err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "task_deleted").
		Str("task_id", taskID.String()).
		Msg("Task deleted successfully")

	return nil
}
