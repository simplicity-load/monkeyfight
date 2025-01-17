package main

import (
	"log/slog"
	"os"
	"time"

	"monkeyfight.com/websocket"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/lmittmann/tint"
)

func main() {
	validate := validator.New(validator.WithRequiredStructEnabled())
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(func(c *fiber.Ctx) error {
		slog.Debug("Log", "ip", c.IP(), "path", c.Path())
		return c.Next()
	})

	v1 := app.Group("/api/v1")

	wsGr := v1.Group("/ws")
	wsGr.Use(websocket.UpgradeWall)

	event := make(chan websocket.CliEvent)
	go websocket.Dispatcher(event)
	wsGr.Get("/game", websocket.Game(event, validate))

	app.Listen(":8800")
}
