package auth

import (
	"golang-sample/internal/model"
	"golang-sample/internal/schemas"
)

// modelToSchemaUser converts domain User to schema User
func modelToSchemaUser(u *model.User) *schemas.User {
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

// schemaToModelUser converts schema User to domain User
func schemaToModelUser(u *schemas.User) *model.User {
	if u == nil {
		return nil
	}

	return &model.User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
