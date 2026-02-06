package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goku-m/main/apps/task/api/model"
	"github.com/goku-m/main/apps/task/api/model/task"
	"github.com/goku-m/main/internal/shared/errs"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	server *server.Server
}

func NewTaskRepository(server *server.Server) *TaskRepository {
	return &TaskRepository{server: server}
}

func (r *TaskRepository) CreateTask(ctx context.Context, payload *task.CreateTaskPayload) (*task.Task, error) {
	stmt := `
		INSERT INTO
			tasks (
				title,
				description,
				priority,
				due_date
							)
		VALUES
			(
				@title,
				@description,
				@priority,
				@due_date
				
			)
		RETURNING
		*
	`
	priority := task.PriorityMedium
	if payload.Priority != nil {
		priority = *payload.Priority
	}

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"title":       payload.Title,
		"description": payload.Description,
		"priority":    priority,
		"due_date":    payload.DueDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create task query for  title=%s: %w", payload.Title, err)
	}

	taskItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[task.Task])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:tasks for  title=%s: %w", payload.Title, err)
	}

	return &taskItem, nil
}

func (r *TaskRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*task.Task, error) {
	stmt := `
	SELECT
		u.*
		
	FROM
		tasks u
	
	WHERE
		u.id=@id
	GROUP BY
		u.id
		
`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": taskID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get task by id query for task_id=%s: %w", taskID.String(), err)
	}

	taskItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[task.Task])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:tasks for task_id=%s task_id=%s: %w", taskID.String(), err)
	}

	return &taskItem, nil
}

func (r *TaskRepository) CheckTaskExists(ctx context.Context, taskID uuid.UUID) (*task.Task, error) {
	stmt := `
		SELECT
			*
		FROM
			tasks
		WHERE
			id=@id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": taskID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check if task exists for task_id=%s: %w", taskID.String(), err)
	}

	taskItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[task.Task])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:tasks for task_id=%s: %w", taskID.String(), err)
	}

	return &taskItem, nil
}

func (r *TaskRepository) GetTasks(
	ctx context.Context,
	query *task.GetTasksQuery,
) (*model.PaginatedResponse[task.PopulatedTask], error) {

	stmt := `
	SELECT
		t.*
	FROM
		tasks t
	`

	args := pgx.NamedArgs{}
	conditions := []string{}

	if query != nil {
		if query.Status != nil {
			conditions = append(conditions, "t.status = @status")
			args["status"] = *query.Status
		}

		if query.Priority != nil {
			conditions = append(conditions, "t.priority = @priority")
			args["priority"] = *query.Priority
		}

		if query.Completed != nil {
			if *query.Completed {
				conditions = append(conditions, "t.status = 'completed'")
			} else {
				conditions = append(conditions, "t.status != 'completed'")
			}
		}

		if query.Search != nil {
			conditions = append(conditions, "(t.title ILIKE @search OR t.description ILIKE @search)")
			args["search"] = "%" + *query.Search + "%"
		}
	}

	if len(conditions) > 0 {
		stmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	// ----- count query -----
	countStmt := "SELECT COUNT(*) FROM tasks t"
	if len(conditions) > 0 {
		countStmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	// If server/DB/Pool can be nil in your wiring, you may also want to guard here:
	// if r == nil || r.server == nil || r.server.DB == nil || r.server.DB.Pool == nil { ... }

	var total int
	err := r.server.DB.Pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count for tasks: %w", err)
	}

	// You don't need GROUP BY for "SELECT t.*" from a single table.
	// But keeping your intent: remove it to avoid unnecessary work.
	// stmt += " GROUP BY t.id"

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

	stmt += " ORDER BY t." + sortCol
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
		return nil, fmt.Errorf("failed to execute get tasks query: %w", err)
	}
	defer rows.Close()

	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[task.PopulatedTask])
	if err != nil {
		// NOTE: CollectRows typically returns nil error with empty slice if no rows,
		// but keeping your fallback logic is fine.
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[task.PopulatedTask]{
				Data:       []task.PopulatedTask{},
				Page:       page,
				Limit:      limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:tasks: %w", err)
	}

	return &model.PaginatedResponse[task.PopulatedTask]{
		Data:       tasks,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}, nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, payload *task.UpdateTaskPayload) (*task.Task, error) {
	stmt := "UPDATE tasks SET "
	args := pgx.NamedArgs{
		"task_id": payload.ID,
	}
	setClauses := []string{}

	if payload.Title != nil {
		setClauses = append(setClauses, "title = @title")
		args["title"] = *payload.Title
	}

	if payload.Description != nil {
		setClauses = append(setClauses, "description = @description")
		args["description"] = *payload.Description
	}

	if payload.Status != nil {
		setClauses = append(setClauses, "status = @status")
		args["status"] = *payload.Status

		// Auto-set completed_at when status changes to completed
		if *payload.Status == task.StatusCompleted {
			setClauses = append(setClauses, "completed_at = @completed_at")
			args["completed_at"] = time.Now()
		} else if *payload.Status != task.StatusCompleted {
			setClauses = append(setClauses, "completed_at = NULL")
		}
	}

	if payload.Priority != nil {
		setClauses = append(setClauses, "priority = @priority")
		args["priority"] = *payload.Priority
	}

	if len(setClauses) == 0 {
		return nil, errs.NewBadRequestError("no fields to update", false, nil, nil, nil)
	}

	stmt += strings.Join(setClauses, ", ")
	stmt += " WHERE id = @task_id  RETURNING *"

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	updatedTask, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[task.Task])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:tasks: %w", err)
	}

	return &updatedTask, nil
}

func (r *TaskRepository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	stmt := `
		DELETE FROM tasks
		WHERE
			id=@task_id
	`

	result, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"task_id": taskID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if result.RowsAffected() == 0 {
		code := "TODO_NOT_FOUND"
		return errs.NewNotFoundError("task not found", false, &code)
	}

	return nil
}
