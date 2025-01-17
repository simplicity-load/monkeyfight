package websocket

import (
	"log/slog"
	"sync"
	"time"

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
			if e.t == Insert && g.IsPlaying() {
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
		slog.Debug("processEvent:Register", "event", event)
		cli := &client{uid: genUid()}
		clients.Store(event.conn, cli)
		sendMessage(event.conn, cli, GameUpdate{g})
	case Unregister:
		slog.Debug("processEvent:Unregister", "event", event)
		clients.Delete(event.conn)
		// TODO possibly notify other players that player quit
	case Insert:
		if !g.IsPlaying() {
			return
		}
		slog.Debug("processEvent:Insert", "event", event)
		cli, ok := clients.Load(event.conn)
		if !ok {
			return // TODO error msg
		}
		err := cli.appendKeyStrokes(event.msg)
		if err != nil {
			slog.Error("processEvent:appendKeyStrokes", "err", err, "cli", cli)
			sendMessage(event.conn, cli, ErrUpdate{
				Msg: err.Error(),
			})
		}
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
	for conn, cli := range clients.All() {
		if conn == event.conn {
			continue
		}
		sendMessage(conn, cli, CliUpdate{
			Uid:        eventCli.uid,
			Keystrokes: event.msg,
		})
	}
}

func broadcastMessage(clients cliStore, msg any) {
	for conn, cli := range clients.All() {
		sendMessage(conn, cli, msg)
	}
}

func sendMessage(conn sKey, cli sVal, v any) {
	go func() {
		cli.Lock()
		defer cli.Unlock()
		err := conn.Conn.WriteJSON(v)
		if err != nil {
			slog.Error("sendMessage", "error", err, "v", v)
		}
	}()
}
