package site

import (
	"github.com/goku-m/main/apps/agrifolio/api/model"
)

type User struct {
	model.Base
	Email    string  `json:"email" db:"email"`
	Password *string `json:"password" db:"password_hash"`
	Plan     *string `json:"plan" db:"plan"`
}
