package service

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	//"github.com/goku-m/mains/internal/shared/lib/aws"

	"github.com/goku-m/main/apps/todo/api/model"
	"github.com/goku-m/main/apps/todo/api/model/todo"
	"github.com/goku-m/main/apps/todo/api/repository"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/goku-m/main/internal/shared/server"
)

type TodoService struct {
	server   *server.Server
	todoRepo *repository.TodoRepository
}

func NewTodoService(server *server.Server, todoRepo *repository.TodoRepository,
) *TodoService {
	return &TodoService{
		server:   server,
		todoRepo: todoRepo,
	}
}

func (s *TodoService) CreateTodo(ctx echo.Context, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)

	// Validate parent todo exists and belongs to todo (if provided)

	todoItem, err := s.todoRepo.CreateTodo(ctx.Request().Context(), payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create todo")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_created").
		Str("todo_id", todoItem.ID.String()).
		Str("title", todoItem.Title).
		Str("priority", string(todoItem.Priority)).
		Msg("Todo created successfully")

	return todoItem, nil
}

func (s *TodoService) GetTodoByID(ctx echo.Context, todoID uuid.UUID) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)

	todoItem, err := s.todoRepo.GetTodoByID(ctx.Request().Context(), todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch todo by ID")
		return nil, err
	}

	return todoItem, nil
}

func (s *TodoService) GetTodos(ctx echo.Context, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
	logger := middleware.GetLogger(ctx)

	result, err := s.todoRepo.GetTodos(ctx.Request().Context(), query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch todos")
		return nil, err
	}

	return result, nil
}

func (s *TodoService) UpdateTodo(ctx echo.Context, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)

	updatedTodo, err := s.todoRepo.UpdateTodo(ctx.Request().Context(), payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update todo")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_updated").
		Str("todo_id", updatedTodo.ID.String()).
		Str("title", updatedTodo.Title).
		Str("priority", string(updatedTodo.Priority)).
		Str("status", string(updatedTodo.Status)).
		Msg("Todo updated successfully")

	return updatedTodo, nil
}

func (s *TodoService) DeleteTodo(ctx echo.Context, todoID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	err := s.todoRepo.DeleteTodo(ctx.Request().Context(), todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete todo")
		return err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_deleted").
		Str("todo_id", todoID.String()).
		Msg("Todo deleted successfully")

	return nil
}
