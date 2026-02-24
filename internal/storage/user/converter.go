package user

import (
	"golang-sample/internal/model"
	"golang-sample/internal/orm"
)

// ormToModel converts ORM User to domain User
func ormToModel(u *orm.User) *model.User {
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

// modelToORM converts domain User to ORM User
func modelToORM(u *model.User) *orm.User {
	if u == nil {
		return nil
	}

	return &orm.User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ormSliceToModelSlice converts slice of ORM Users to domain Users
func ormSliceToModelSlice(users []*orm.User) []*model.User {
	if users == nil {
		return nil
	}

	result := make([]*model.User, len(users))
	for i, u := range users {
		result[i] = ormToModel(u)
	}
	return result
}
