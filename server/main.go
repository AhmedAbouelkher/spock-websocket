package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

var (
	Port = GetenvDef("PORT", "3000")
)

func main() {
	// handle signals, graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	setupServer()
	defer CloseDB()

	app := fiber.New(fiber.Config{
		AppName:      "Spock",
		ServerHeader: "Spock",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*MultipleErrorsErr); ok {
				return c.Status(e.StatusCode).JSON(fiber.Map{
					"status":  e.StatusCode,
					"message": e.Message,
					"errors":  e.Errors,
				})
			}
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"status":  code,
				"message": err.Error(),
			})
		},
	})

	app.Get("/metrics", monitor.New(monitor.Config{Title: "Spock Metrics Page"}))
	app.Use(etag.New(), cors.New(), requestid.New(), compress.New(), logger.New())

	SetupRouters(app)

	go func() {
		err := app.Listen(":" + Port)
		if err != nil {
			signals <- syscall.SIGKILL
		}
	}()

	<-signals
	go func() {
		<-signals
		os.Exit(1)
	}()

	if err := app.Shutdown(); err != nil {
		AppLogger.WithError(err).Error("failed to shutdown server")
	}

	wsClientsPool.close()

	AppLogger.Info("server stopped")
}

func setupServer() {
	if err := InitLogger(); err != nil {
		panic(err)
	}
	if err := OpenDB(); err != nil {
		panic(err)
	}
}
