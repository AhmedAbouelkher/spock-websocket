package main

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get token from request
		token := strings.TrimPrefix(strings.TrimSpace(c.Get("Authorization")), "Bearer ")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header is required")
		}
		// validate token
		user, err := GetUserWithToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		// set user to context
		c.Locals("user", user)
		return c.Next()
	}
}

func WSAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := strings.TrimSpace(c.Query("token"))
		// get token from request
		token := strings.TrimPrefix(strings.TrimSpace(c.Get("Authorization")), "Bearer ")
		if token == "" && accessToken == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header or access_token query param is required")
		}
		if accessToken != "" {
			token = accessToken
		}
		// validate token
		user, err := GetUserWithToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		// set user to context
		c.Locals("user", user)
		return c.Next()
	}
}
