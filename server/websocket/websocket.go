package websocket

import (
	"log/slog"
	"sync"
	"time"

	fiberWS "github.com/gofiber/contrib/websocket"
	"monkeyfight.com/game"
)

// in seconds
const GAME_LENGTH = 30
const gameLength = GAME_LENGTH * time.Second

func Dispatcher(event <-chan CliEvent) {
	clients := cliStore{&sync.Map{}}
	g := game.New()
	timer := time.NewTimer(gameLength)

	for {
		slog.Debug("Dispatcher", "game", g.Words[:100], "state", g.State)
		slog.Debug("Players", "p", clients)
		select {
		case e := <-event:
			go processEvent(clients, e, *g)
			if e.t == Insert {
				go broadcastEvent(clients, e)
			}
		case <-timer.C:
			timer.Reset(gameLength)

			if g.IsPlaying() {
				g = game.New()
				clients.resetClientKeystrokes()
			} else if clients.clientCount() >= 2 {
				g.Start()
			}
			go broadcastMessage(clients, GameUpdate{*g})
		}

	}
}

func processEvent(clients cliStore, event CliEvent, g game.Game) {
	switch event.t {
	case Register:
		slog.Debug("eventProcessor:Register", "event", event)
		clients.Store(event.conn, &client{uid: genUid()})
		sendMessage(event.conn, GameUpdate{g})
	case Unregister:
		slog.Debug("eventProcessor:Unregister", "event", event)
		clients.Delete(event.conn)
		// TODO possibly notify other players that player quit
	case Insert:
		if !g.IsPlaying() {
			return
		}
		slog.Debug("eventProcessor:Insert", "event", event)
		cli, ok := clients.Load(event.conn)
		if !ok {
			return // TODO error msg
		}

		cli.Lock()
		defer cli.Unlock()

		cli.keystrokes = append(cli.keystrokes, event.msg...)
	}
}

func broadcastEvent(clients cliStore, event CliEvent) {
	eventCli, ok := clients.Load(event.conn)
	if !ok {
		// FIXME on register, broadcast runs before
		// process-ing so we don't have a client in the map
		slog.Error("broadcastEvent:clients.clientByConn", "ip", event.conn.Conn.LocalAddr())
		return
	}
	clients.Range(func(c, _ any) bool {
		conn, ok := c.(*fiberWS.Conn)
		if !ok {
			return true // TODO error msg
		}
		if conn == event.conn {
			return true
		}

		sendMessage(conn, CliUpdate{
			Uid:        eventCli.uid,
			Keystrokes: event.msg,
		})
		return true
	})
}

func broadcastMessage(clients cliStore, msg any) {
	for conn := range clients.All() {
		sendMessage(conn, msg)
	}
}

func sendMessage(conn *fiberWS.Conn, v any) {
	go func() {
		err := conn.Conn.WriteJSON(v)
		if err != nil {
			slog.Error("sendMessage", "error", err, "v", v)
		}
	}()
}
