package service

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	//"github.com/goku-m/mains/internal/shared/lib/aws"

	"github.com/goku-m/main/apps/agrifolio/api/model"
	"github.com/goku-m/main/apps/agrifolio/api/model/user"
	"github.com/goku-m/main/apps/agrifolio/api/repository"
	"github.com/goku-m/main/internal/shared/middleware"
	"github.com/goku-m/main/internal/shared/server"
)

type UserService struct {
	server   *server.Server
	userRepo *repository.UserRepository
}

func NewUserService(server *server.Server, userRepo *repository.UserRepository,
) *UserService {
	return &UserService{
		server:   server,
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUser(ctx echo.Context,  payload *user.CreateUserPayload) (*user.User, error) {
	logger := middleware.GetLogger(ctx)

	// Validate parent user exists and belongs to user (if provided)

	userItem, err := s.userRepo.CreateUser(ctx.Request().Context(), payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create user")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "user_created").
		Str("user_id", userItem.ID.String()).
		Str("name", userItem.Name).
		Str("email", userItem.Email).
		Msg("User created successfully")

	return userItem, nil
}

func (s *UserService) GetUserByID(ctx echo.Context, userID uuid.UUID) (*user.User, error) {
	logger := middleware.GetLogger(ctx)

	userItem, err := s.userRepo.GetUserByID(ctx.Request().Context(), userID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch user by ID")
		return nil, err
	}

	return userItem, nil
}

func (s *UserService) GetUsers(ctx echo.Context, query *user.GetUserQuery) (*model.PaginatedResponse[user.User], error) {
	logger := middleware.GetLogger(ctx)

	result, err := s.userRepo.GetUsers(ctx.Request().Context(), query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch users")
		return nil, err
	}

	return result, nil
}

func (s *UserService) UpdateUser(ctx echo.Context, payload *user.UpdateUserPayload) (*user.User, error) {
	logger := middleware.GetLogger(ctx)

	updatedUser, err := s.userRepo.UpdateUser(ctx.Request().Context(), payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update user")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "user_updated").
		Str("user_id", updatedUser.ID.String()).
		Str("name", updatedUser.Name).
		Str("email", updatedUser.Email).
		Msg("User updated successfully")

	return updatedUser, nil
}

func (s *UserService) DeleteUser(ctx echo.Context, userID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	err := s.userRepo.DeleteUser(ctx.Request().Context(), userID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete user")
		return err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "user_deleted").
		Str("user_id", userID.String()).
		Msg("User deleted successfully")

	return nil
}

