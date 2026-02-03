package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goku-m/main/apps/agrifolio/api/model"
	"github.com/goku-m/main/apps/agrifolio/api/model/site"
	"github.com/goku-m/main/internal/shared/errs"
	"github.com/goku-m/main/internal/shared/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SiteRepository struct {
	server *server.Server
}

func NewSiteRepository(server *server.Server) *SiteRepository {
	return &SiteRepository{server: server}
}

func (r *SiteRepository) CreateSite(ctx context.Context, userID string, payload *site.CreateSitePayload) (*site.Site, error) {
	stmt := `
		INSERT INTO
			sites (
				user_id,
				title,
				description,
				priority,
				due_date
							)
		VALUES
			(
				@user_id,
				@title,
				@description,
				@priority,
				@due_date
				
			)
		RETURNING
		*
	`
	priority := site.PriorityMedium
	if payload.Priority != nil {
		priority = *payload.Priority
	}

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":     userID,
		"title":       payload.Title,
		"description": payload.Description,
		"priority":    priority,
		"due_date":    payload.DueDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create site query for user_id=%s title=%s: %w", userID, payload.Title, err)
	}

	siteItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[site.Site])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:sites for user_id=%s title=%s: %w", userID, payload.Title, err)
	}

	return &siteItem, nil
}

func (r *SiteRepository) GetSiteByID(ctx context.Context, userID string, siteID uuid.UUID) (*site.PopulatedSite, error) {
	stmt := `
	SELECT
		t.*
		
	FROM
		sites t
	
	WHERE
		t.id=@id
		AND t.user_id=@user_id
	GROUP BY
		t.id
		
`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      siteID,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get site by id query for site_id=%s user_id=%s: %w", siteID.String(), userID, err)
	}

	siteItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[site.PopulatedSite])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:sites for site_id=%s user_id=%s: %w", siteID.String(), userID, err)
	}

	return &siteItem, nil
}

func (r *SiteRepository) CheckSiteExists(ctx context.Context, userID string, siteID uuid.UUID) (*site.Site, error) {
	stmt := `
		SELECT
			*
		FROM
			sites
		WHERE
			id=@id
			AND user_id=@user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      siteID,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check if site exists for site_id=%s user_id=%s: %w", siteID.String(), userID, err)
	}

	siteItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[site.Site])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:sites for site_id=%s user_id=%s: %w", siteID.String(), userID, err)
	}

	return &siteItem, nil
}

func (r *SiteRepository) GetSites(
	ctx context.Context,
	query *site.GetSitesQuery,
) (*model.PaginatedResponse[site.PopulatedSite], error) {

	stmt := `
	SELECT
		t.*
	FROM
		sites t
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
	countStmt := "SELECT COUNT(*) FROM sites t"
	if len(conditions) > 0 {
		countStmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	// If server/DB/Pool can be nil in your wiring, you may also want to guard here:
	// if r == nil || r.server == nil || r.server.DB == nil || r.server.DB.Pool == nil { ... }

	var total int
	err := r.server.DB.Pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count for sites: %w", err)
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
		return nil, fmt.Errorf("failed to execute get sites query: %w", err)
	}
	defer rows.Close()

	sites, err := pgx.CollectRows(rows, pgx.RowToStructByName[site.PopulatedSite])
	if err != nil {
		// NOTE: CollectRows typically returns nil error with empty slice if no rows,
		// but keeping your fallback logic is fine.
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[site.PopulatedSite]{
				Data:       []site.PopulatedSite{},
				Page:       page,
				Limit:      limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:sites: %w", err)
	}

	return &model.PaginatedResponse[site.PopulatedSite]{
		Data:       sites,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}, nil
}

func (r *SiteRepository) UpdateSite(ctx context.Context, userID string, payload *site.UpdateSitePayload) (*site.Site, error) {
	stmt := "UPDATE sites SET "
	args := pgx.NamedArgs{
		"site_id": payload.ID,
		"user_id": userID,
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
		if *payload.Status == site.StatusCompleted {
			setClauses = append(setClauses, "completed_at = @completed_at")
			args["completed_at"] = time.Now()
		} else if *payload.Status != site.StatusCompleted {
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
	stmt += " WHERE id = @site_id AND user_id = @user_id RETURNING *"

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	updatedSite, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[site.Site])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:sites: %w", err)
	}

	return &updatedSite, nil
}

func (r *SiteRepository) DeleteSite(ctx context.Context, userID string, siteID uuid.UUID) error {
	stmt := `
		DELETE FROM sites
		WHERE
			id=@site_id
			AND user_id=@user_id
	`

	result, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"site_id": siteID,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if result.RowsAffected() == 0 {
		code := "TODO_NOT_FOUND"
		return errs.NewNotFoundError("site not found", false, &code)
	}

	return nil
}

func (r *SiteRepository) GetSiteStats(ctx context.Context, userID string) (*site.SiteStats, error) {
	stmt := `
		SELECT
			COUNT(*) AS total,
			COUNT(
				CASE
					WHEN status='draft' THEN 1
				END
			) AS draft,
			COUNT(
				CASE
					WHEN status='active' THEN 1
				END
			) AS active,
			COUNT(
				CASE
					WHEN status='completed' THEN 1
				END
			) AS completed,
			COUNT(
				CASE
					WHEN status='archived' THEN 1
				END
			) AS archived,
			COUNT(
				CASE
					WHEN due_date<NOW()
					AND status!='completed' THEN 1
				END
			) AS overdue
		FROM
			sites
		WHERE
			user_id=@user_id
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	stats, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[site.SiteStats])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:sites: %w", err)
	}

	return &stats, nil
}
