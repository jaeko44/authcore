// Contact is depreciated in 1.0. This file is maintained for backward-compatibility only. New
// functions should not rely on Contact.

package user

import (
	"time"

	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/nulls"
)

// ContactType is a type enumerating the type for contact.
type ContactType int32

// Enumerates the ContactType
const (
	ContactEMAIL ContactType = 0
	ContactPHONE ContactType = 1
)

// Contact represents a contact information (currently email) of an user.
type Contact struct {
	ID         int64
	UserID     int64        `db:"user_id"`
	Type       ContactType  `db:"type"`
	Value      nulls.String `db:"value" validate:"email_or_phone=Contact.Type"`
	IsPrimary  bool         `db:"is_primary"`
	UpdatedAt  time.Time    `db:"updated_at"`
	CreatedAt  time.Time    `db:"created_at"`
	VerifiedAt nulls.Time   `db:"verified_at"`
}

// Validate validates the struct of a Contact model.
func (contact *Contact) Validate() error {
	return validator.Validate.Struct(contact)
}

// IsVerified checks if a contact is verified
func (contact *Contact) IsVerified() bool {
	return contact.VerifiedAt.Valid
}
