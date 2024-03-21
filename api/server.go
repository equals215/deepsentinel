package api

import (
	"github.com/equals215/deepsentinel/monitoring"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func newServer() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "DeepSentinel API",
	})

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

	app.Post("/probe/:machine/report", func(c *fiber.Ctx) error {
		machine := c.Params("machine")

		// This shouldn't happen, desgined to catch Fiber's bug if ever
		if machine == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "fail",
				"error":  "machine name is required",
			})
		}
		if c.Body() == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"machine": machine,
				"error":   "payload is required",
			})
		}

		err := monitoring.IngestPayload(machine, c.Body())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "fail",
				"machine": machine,
				"error":   err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"status":  "pass",
			"machine": machine,
		})
	})

	app.Delete("/probe/:machine/service/:service", func(c *fiber.Ctx) error {
		machine := c.Params("machine")
		service := c.Params("service")

		// This shouldn't happen, desgined to catch Fiber's bug if ever
		if machine == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "fail",
				"error":  "machine name is required",
			})
		}
		if service == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"machine": machine,
				"error":   "service name is required",
			})
		}
		if c.Body() == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"machine": machine,
				"error":   "payload is required",
			})
		}

		return nil
	})

	return app
}
