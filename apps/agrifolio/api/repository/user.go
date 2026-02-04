package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/goku-m/main/apps/agrifolio/api/model"
	"github.com/goku-m/main/apps/agrifolio/api/model/user"
	"github.com/goku-m/main/internal/shared/errs"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	server *server.Server
}

func NewUserRepository(server *server.Server) *UserRepository {
	return &UserRepository{server: server}
}

func (r *UserRepository) CreateUser(ctx context.Context, payload *user.CreateUserPayload) (*user.User, error) {
	stmt := `
		INSERT INTO
			users (
				name,
				email,
				paswoard_hash,
							)
		VALUES
			(
				@name,
				@email,
				@paswoard_hash,
				
			)
		RETURNING
		*
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"name":          payload.Name,
		"email":         payload.Email,
		"password_hash": payload.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create user query for user=%s email=%s: %w", payload.Name, *payload.Email, err)
	}

	userItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:users for user=%s email=%s: %w", payload.Name, *payload.Email, err)
	}

	return &userItem, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	stmt := `
	SELECT
		u.*
		
	FROM
		users u
	
	WHERE
		u.id=@id
	GROUP BY
		u.id
		
`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get user by id query for user_id=%s: %w", userID.String(), err)
	}

	userItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:users for user_id=%s user_id=%s: %w", userID.String(), err)
	}

	return &userItem, nil
}

func (r *UserRepository) CheckUserExists(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	stmt := `
		SELECT
			*
		FROM
			users
		WHERE
			id=@id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists for user_id=%s: %w", userID.String(), err)
	}

	userItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:users for user_id=%s: %w", userID.String(), err)
	}

	return &userItem, nil
}

func (r *UserRepository) GetUsers(
	ctx context.Context,
	query *user.GetUserQuery,
) (*model.PaginatedResponse[user.User], error) {

	stmt := `
	SELECT
		u.*
	FROM
		users u
	`

	args := pgx.NamedArgs{}
	conditions := []string{}

	if query != nil {
		if query.Search != nil {
			conditions = append(conditions, "(t.title ILIKE @search OR t.description ILIKE @search)")
			args["search"] = "%" + *query.Search + "%"
		}
	}

	if len(conditions) > 0 {
		stmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	// ----- count query -----
	countStmt := "SELECT COUNT(*) FROM users u"
	if len(conditions) > 0 {
		countStmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	// If server/DB/Pool can be nil in your wiring, you may also want to guard here:
	// if r == nil || r.server == nil || r.server.DB == nil || r.server.DB.Pool == nil { ... }

	var total int
	err := r.server.DB.Pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count for users: %w", err)
	}

	// ----- safe defaults for pagination -----
	page := 1
	limit := 10
	if query != nil {
		if query.Page != nil && *query.Page > 0 {
			page = *query.Page
		}
		if query.Limit != nil && *query.Limit > 0 {
			limit = *query.Limit
		}
	}

	// ----- safe sorting (whitelist to prevent SQL injection) -----
	sortCol := "created_at"
	orderDesc := true

	allowedSort := map[string]bool{
		"created_at": true,
		"title":      true,
		"priority":   true,
		"status":     true,
		"updated_at": true,
	}

	if query != nil && query.Sort != nil && allowedSort[*query.Sort] {
		sortCol = *query.Sort
	}

	if query != nil && query.Order != nil && strings.EqualFold(*query.Order, "desc") {
		orderDesc = true
	}

	stmt += " ORDER BY u." + sortCol
	if orderDesc {
		stmt += " DESC"
	} else {
		stmt += " ASC"
	}

	// ----- pagination -----
	stmt += " LIMIT @limit OFFSET @offset"
	args["limit"] = limit
	args["offset"] = (page - 1) * limit

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get users query: %w", err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		// NOTE: CollectRows typically returns nil error with empty slice if no rows,
		// but keeping your fallback logic is fine.
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[user.User]{
				Data:       []user.User{},
				Page:       page,
				Limit:      limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:users: %w", err)
	}

	return &model.PaginatedResponse[user.User]{
		Data:       users,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, payload *user.UpdateUserPayload) (*user.User, error) {
	stmt := "UPDATE users SET "
	args := pgx.NamedArgs{
		"user_id": payload.ID,
	}
	setClauses := []string{}

	if payload.Name != nil {
		setClauses = append(setClauses, "name = @name")
		args["name"] = *&payload.Name
	}

	if payload.Password != nil {
		setClauses = append(setClauses, "password_hash = @password_hash")
		args["description"] = *payload.Password
	}

	if len(setClauses) == 0 {
		return nil, errs.NewBadRequestError("no fields to update", false, nil, nil, nil)
	}

	stmt += strings.Join(setClauses, ", ")
	stmt += " WHERE id = @user_id AND user_id = @user_id RETURNING *"

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	updatedUser, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:users: %w", err)
	}

	return &updatedUser, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	stmt := `
		DELETE FROM users
		WHERE
			id=@user_id
	`

	result, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if result.RowsAffected() == 0 {
		code := "TODO_NOT_FOUND"
		return errs.NewNotFoundError("user not found", false, &code)
	}

	return nil
}
