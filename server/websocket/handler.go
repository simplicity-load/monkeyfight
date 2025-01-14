package websocket

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	fiberWS "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"monkeyfight.com/game"
)

func UpgradeWall(c *fiber.Ctx) error {
	if fiberWS.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return c.SendStatus(fiber.StatusUpgradeRequired)
}

func Game(event chan<- CliEvent, validator *validator.Validate) fiber.Handler {
	return fiberWS.New(func(c *fiberWS.Conn) {
		sendInternalEvent := func(t EventType, msg game.Keystrokes) {
			event <- CliEvent{conn: c, t: t, msg: msg}
		}
		NoMsg := game.Keystrokes{}

		sendInternalEvent(Register, NoMsg)
		defer sendInternalEvent(Unregister, NoMsg)

		// Read client input
		for {
			msg := &CliMsg{}
			if err := c.ReadJSON(msg); err != nil {
				slog.Error("Game:ReadJSON", "error", err)
				return // close connection
			}

			if err := validator.Struct(msg); err != nil {
				slog.Error("Game:Validate", "error", err)
				// sendInternalEvent(BadData, NoMsg)
				continue
			}

			sendInternalEvent(msg.Event, msg.Keystrokes)
		}
	})
}
