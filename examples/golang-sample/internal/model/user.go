package model

import (
	"time"
)

// User represents a domain user with business logic.
// This is a pure domain entity with no dependencies on persistence (ORM) or API (schemas) layers.
type User struct {
	ID        uint
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks if user data is valid according to business rules.
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrUsernameRequired
	}
	if u.Email == "" {
		return ErrEmailRequired
	}
	if len(u.Username) < 3 {
		return ErrUsernameTooShort
	}
	if len(u.Username) > 50 {
		return ErrUsernameTooLong
	}
	return nil
}

// CanLogin checks if user can perform login operation.
func (u *User) CanLogin() bool {
	if u == nil {
		return false
	}
	return u.Username != "" && u.Email != ""
}

// IsNew checks if user is not yet persisted (ID not set).
func (u *User) IsNew() bool {
	if u == nil {
		return true
	}
	return u.ID == 0
}

// IsEqual checks if two users are the same by comparing IDs.
func (u *User) IsEqual(other *User) bool {
	if u == nil || other == nil {
		return false
	}
	return u.ID == other.ID && u.Username == other.Username
}

// Clone creates a deep copy of the user.
func (u *User) Clone() *User {
	if u == nil {
		return nil
	}
	return &User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
