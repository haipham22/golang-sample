package schemas

type UserRegisterRequest struct {
	Username string `form:"username" json:"username" validate:"required"`
	Password string `form:"password" json:"password" validate:"required"`
	Email    string `form:"email" json:"email" validate:"required,email"`
	FullName string `form:"full_name" json:"full_name" validate:"required"`
}

type LoginRequest struct {
	Username string `form:"username" json:"username" validate:"required"`
	Password string `form:"password" json:"password" validate:"required"`
}
