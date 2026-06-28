package auth

import (
	"github.com/haipham22/golang-sample/internal/domain"
	"github.com/haipham22/golang-sample/internal/schemas"
)

// modelToSchemaUser converts domain User to schema User
func modelToSchemaUser(u *domain.User) *schemas.User {
	if u == nil {
		return nil
	}

	return &schemas.User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
