package middlewares

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JWTConfig struct {
	Secret string
}

// JWTAuthMiddleware validates JWT tokens and adds user claims to context
func JWTAuthMiddleware(config JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Missing authorization header",
				})
			}

			// Extract token from "Bearer <token>"
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Invalid authorization format",
				})
			}

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.ErrUnauthorized
				}
				return []byte(config.Secret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Invalid or expired token",
				})
			}

			// Extract claims and set in context
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user_id", claims["id"])
				c.Set("email", claims["email"])
				c.Set("username", claims["username"])
			}

			return next(c)
		}
	}
}
