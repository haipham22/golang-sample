package schemas

import "github.com/golang-jwt/jwt/v5"

type JwtClaims struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JwtResponse struct {
	Token string `json:"token"`
}
