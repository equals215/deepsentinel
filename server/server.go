package server

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/equals215/deepsentinel/dashboard"
	"github.com/equals215/deepsentinel/monitoring"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/utils"
)

//go:embed static/*
var dashboardStatic embed.FS

func newServer(monitoringOperator *monitoring.Operator, dashboardOperator *dashboard.Operator) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "DeepSentinel API",
	})

	fiberSetAuth(app)

	app.Get("/health", getHealthHandler)

	app.Post("/probe/:machine/report", func(c *fiber.Ctx) error {
		return postProbeReportHandler(c, monitoringOperator.In)
	})

	app.Delete("/probe/:machine", func(c *fiber.Ctx) error {
		return deleteProbeHandler(c, monitoringOperator.In)
	})

	if dashboardOperator != nil {
		app.Use("/dashws", func(c *fiber.Ctx) error {
			if websocket.IsWebSocketUpgrade(c) {
				c.Locals("allowed", true)
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		})

		app.Get("/dashws", websocket.New(func(c *websocket.Conn) {
			worker, workerID := dashboardOperator.NewWorker()
			defer dashboardOperator.RemoveWorker(workerID)

			for {
				select {
				case data := <-worker:
					c.WriteJSON(data)
				}
			}
		}))

		app.Get("/dashboard", func(c *fiber.Ctx) error {
			return filesystem.SendFile(c, http.FS(dashboardStatic), "static/index.html")
		})

		app.Get("/", func(c *fiber.Ctx) error {
			return filesystem.SendFile(c, http.FS(dashboardStatic), "static/login.html")
		})
	}

	return app
}

func getHealthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "pass",
	})
}

func postProbeReportHandler(c *fiber.Ctx, payloadChannel chan *monitoring.Payload) error {
	machine := utils.CopyString(c.Params("machine"))

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
	err := json.Unmarshal(c.Body(), parsedPayload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "fail",
			"machine": machine,
			"error":   err.Error(),
		})
	}

	parsedPayload.Timestamp = time.Now()
	parsedPayload.Machine = strings.TrimSpace(machine)

	payloadChannel <- parsedPayload
	return c.SendStatus(fiber.StatusAccepted)
}

func deleteProbeHandler(c *fiber.Ctx, payloadChannel chan *monitoring.Payload) error {
	machine := utils.CopyString(c.Params("machine"))

	// This shouldn't happen, desgined to catch Fiber's bug if ever
	if machine == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "machine name is required",
		})
	}

	parsedPayload := &monitoring.Payload{
		Machine:       strings.TrimSpace(machine),
		MachineStatus: "delete",
		Timestamp:     time.Now(),
	}

	payloadChannel <- parsedPayload
	return c.SendStatus(fiber.StatusAccepted)
}
