package api

import (
	"os"

	"github.com/equals215/deepsentinel/monitoring"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	log "github.com/sirupsen/logrus"
)

func initLogger() {
	log.SetOutput(os.Stdout)
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.Info("deepSentinel API server starting...")
}

func newServer() *fiber.App {
	initLogger()

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

		// This shouldn't happen, desgined to catch Fiber's bug if ever
		if machine == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "fail",
				"error":  "machine name is required",
			})
		} else if c.Body() == nil {
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

	return app
}
