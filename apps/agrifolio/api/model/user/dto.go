package site

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateUserPayload struct {
	Title       string  `json:"title" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

func (p *CreateUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type UpdateUserPayload struct {
	ID          uuid.UUID `param:"id" validate:"required,uuid"`
	Title       *string   `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string   `json:"description" validate:"omitempty,max=1000"`
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
