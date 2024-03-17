package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func newServer() *fiber.App {
	app := fiber.New()

	app.Use(keyauth.New(keyauth.Config{
		Next:      authFilter,
		KeyLookup: "header:Authorization",
		Validator: validateAuth,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "pass",
		})
	})

	app.Post("/report/:machine", func(c *fiber.Ctx) error {
		machine := c.Params("machine")
		return c.JSON(fiber.Map{
			"status":  "received",
			"machine": machine,
		})
	})

	return app
}
