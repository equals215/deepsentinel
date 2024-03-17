package api

import (
	"github.com/gofiber/fiber/v2"
)

// NewServer creates a new server instance
func NewServer() *fiber.App {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋")
	})

	return app
}
