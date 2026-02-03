package site

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateSitePayload struct {
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description *string    `json:"description" validate:"omitempty,max=1000"`
	Priority    *Priority  `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"dueDate"`
}

func (p *CreateSitePayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type UpdateSitePayload struct {
	ID          uuid.UUID `param:"id" validate:"required,uuid"`
	Title       *string   `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string   `json:"description" validate:"omitempty,max=1000"`
	Status      *Status   `json:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority    *Priority `json:"priority" validate:"omitempty,oneof=low medium high"`
}

func (p *UpdateSitePayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type GetSitesQuery struct {
	Page      *int      `query:"page" validate:"omitempty,min=1"`
	Limit     *int      `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort      *string   `query:"sort" validate:"omitempty,oneof=created_at updated_at title priority due_date status"`
	Order     *string   `query:"order" validate:"omitempty,oneof=asc desc"`
	Search    *string   `query:"search" validate:"omitempty,min=1"`
	Status    *Status   `query:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority  *Priority `query:"priority" validate:"omitempty,oneof=low medium high"`
	Overdue   *bool     `query:"overdue"`
	Completed *bool     `query:"completed"`
}

func (q *GetSitesQuery) Validate() error {
	validate := validator.New()

	if err := validate.Struct(q); err != nil {
		return err
	}

	// Set defaults for pagination
	if q.Page == nil {
		defaultPage := 1
		q.Page = &defaultPage
	}
	if q.Limit == nil {
		defaultLimit := 20
		q.Limit = &defaultLimit
	}
	if q.Sort == nil {
		defaultSort := "created_at"
		q.Sort = &defaultSort
	}
	if q.Order == nil {
		defaultOrder := "desc"
		q.Order = &defaultOrder
	}

	return nil
}

// ------------------------------------------------------------

type GetSiteByIDPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *GetSiteByIDPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type DeleteSitePayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *DeleteSitePayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
