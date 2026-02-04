package user

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateUserPayload struct {
	Name     string  `json:"name" db:"name"`
	Email    *string `json:"email" db:"email"`
	Password *string `json:"password" db:"password_hash"`
}

func (p *CreateUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
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

type UpdateUserPayload struct {
	ID       uuid.UUID `param:"id" validate:"required,uuid"`
	Name     *string    `json:"name" db:"name"`
	Password *string   `json:"password" db:"password_hash"`
}

func (p *UpdateUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type DeleteUserPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *DeleteUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type GetUserQuery struct {
	Page   *int    `query:"page" validate:"omitempty,min=1"`
	Limit  *int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort   *string `query:"sort" validate:"omitempty,oneof=created_at updated_at title priority due_date status"`
	Order  *string `query:"order" validate:"omitempty,oneof=asc desc"`
	Search *string `query:"search" validate:"omitempty,min=1"`
}

func (q *GetUserQuery) Validate() error {
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
