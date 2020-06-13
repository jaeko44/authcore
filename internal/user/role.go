package user

import (
	"time"

	"authcore.io/authcore/internal/validator"
)

// Role represents a role record in database.
type Role struct {
	ID           int64
	Name         string    `db:"name"`
	IsSystemRole bool      `db:"is_system_role"`
	UpdatedAt    time.Time `db:"updated_at"`
	CreatedAt    time.Time `db:"created_at"`
}

// RoleUser represents a subject assignment (role-user relation) in database.
type RoleUser struct {
	ID     int64
	RoleID int64 `db:"role_id"`
	UserID int64 `db:"user_id"`
}

// Validate validates the struct of a Role model.
func (role *Role) Validate() error {
	return validator.Validate.Struct(role)
}

// Validate validates the struct of a RoleUser model.
func (roleUser *RoleUser) Validate() error {
	return validator.Validate.Struct(roleUser)
}
