package api

import (
	"encoding/json"
	"time"

	"github.com/equals215/deepsentinel/monitoring"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func newServer(payloadChannel chan *monitoring.Payload) *fiber.App {
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

		parsedPayload := &monitoring.Payload{}
		err := json.Unmarshal(c.Body(), &parsedPayload)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "fail",
				"machine": machine,
				"error":   err.Error(),
			})
		}

		parsedPayload.Timestamp = time.Now()
		parsedPayload.Machine = machine

		payloadChannel <- parsedPayload
		return c.SendStatus(fiber.StatusAccepted)
	})

	app.Delete("/probe/:machine", func(c *fiber.Ctx) error {
		machine := c.Params("machine")

		// This shouldn't happen, desgined to catch Fiber's bug if ever
		if machine == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status": "fail",
				"error":  "machine name is required",
			})
		}

		payload := &monitoring.Payload{
			Machine:       machine,
			MachineStatus: "delete",
			Timestamp:     time.Now(),
		}

		payloadChannel <- payload
		return c.SendStatus(fiber.StatusAccepted)
	})

	return app
}
