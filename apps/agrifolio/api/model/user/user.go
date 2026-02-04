package user

import (
	"github.com/goku-m/main/apps/agrifolio/api/model"
)

type User struct {
	model.Base
	Name     string  `json:"name" db:"name"`
	Email    string  `json:"email" db:"email"`
	Password *string `json:"password" db:"password_hash"`
}
