package middlewares

import (
	"context"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gofiber/fiber/v2"
	"go.messenger/config"
	"go.messenger/services"
)

func FirebaseAuthMiddleware(authClient *auth.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header missing"})
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Authorization header format"})
		}

		if config.LocalMode() {
			uid, err := services.ParseToken(tokenString)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
			}
			c.Locals("firebaseId", uid)
			return c.Next()
		}

		if authClient == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Firebase auth client not initialized"})
		}

		token, err := authClient.VerifyIDToken(context.Background(), tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		c.Locals("firebaseId", token.UID)

		return c.Next()
	}
}
