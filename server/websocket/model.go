package websocket

import (
	fiberWS "github.com/gofiber/contrib/websocket"
	"monkeyfight.com/game"
)

type EventType string

const (
	BadData    EventType = "bad_data"
	Register             = "register"
	Unregister           = "unregister"
	Join                 = "join"
	Insert               = "insert"
)

type CliEvent struct {
	conn *fiberWS.Conn
	t    EventType
	msg  game.Keystrokes
}

type CliMsg struct {
	Event      EventType       `json:"e" validate:"required,oneof=join insert"`
	Keystrokes game.Keystrokes `json:"ks"`
	// Status   int     `json:"s"`
}

type CliUpdate struct {
	Uid        uid             `json:"uid"`
	Keystrokes game.Keystrokes `json:"ks"`
}
type GameUpdate struct {
	Game game.Game `json:"g"`
}
type ErrUpdate struct {
	Msg string `json:"err"`
}
