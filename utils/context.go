// utils/context.go
package utils

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func GetUserID(c *fiber.Ctx) (uint, error) {
	token, ok := c.Locals("jwt").(*jwt.Token)
	if !ok || token == nil {
		return 0, errors.New("no jwt token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if v, ok := claims["user_id"]; ok {
			switch t := v.(type) {
			case float64:
				return uint(t), nil
			case int:
				return uint(t), nil
			}
		}
	}
	return 0, errors.New("invalid claims")
}
