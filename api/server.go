package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func newServer() *fiber.App {
	app := fiber.New()

	app.Use(keyauth.New(keyauth.Config{
		KeyLookup: "header:Authorization",
		Validator: validateAuth,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋")
	})

	return app
}
