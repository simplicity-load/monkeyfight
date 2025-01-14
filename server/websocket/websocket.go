package websocket

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	fiberWS "github.com/gofiber/contrib/websocket"
	"monkeyfight.com/game"
)

type uid = string
type client struct {
	sync.Mutex
	uid        uid
	keystrokes game.Keystrokes
}

func genUid() uid {
	r := rand.IntN(60) | 1
	now := uint64(time.Now().UnixMilli() << r)
	return fmt.Sprintf("mf-%d", now)
}

type cliStore = *sync.Map

// in seconds
const GAME_LENGTH = 10
const gameLength = GAME_LENGTH * time.Second

func Dispatcher(event <-chan CliEvent) {
	var clients cliStore = &sync.Map{}
	var g = game.New()
	var timer = time.NewTimer(gameLength)

	for {
		slog.Debug("Dispatcher", "game", g.Words[:100], "state", g.State)
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
			} else if clientCount(clients) >= 2 {
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
		slog.Debug("eventProcessor:Insert", "event", event)
		cli, ok := clientByConn(event.conn, clients)
		if !ok {
			return // TODO error msg
		}

		cli.Lock()
		defer cli.Unlock()

		cli.keystrokes = append(cli.keystrokes, event.msg...)
	}
}

func broadcastEvent(clients cliStore, event CliEvent) {
	eventCli, ok := clientByConn(event.conn, clients)
	if !ok {
		// FIXME on register, broadcast runs before
		// process-ing so we don't have a client in the map
		slog.Error("broadcastEvent:clientByConn", "ip", event.conn.Conn.LocalAddr())
		return
	}
	clients.Range(func(c, cli any) bool {
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
	clients.Range(func(c, cli any) bool {
		conn, ok := c.(*fiberWS.Conn)
		if !ok {
			return true // TODO error msg
		}
		sendMessage(conn, msg)
		return true
	})
}

func clientCount(clients cliStore) int {
	count := 0
	clients.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

func clientByConn(conn *fiberWS.Conn, clients cliStore) (*client, bool) {
	v, ok := clients.Load(conn)
	if !ok {
		slog.Error("clientByConn:Load")
		return nil, false
	}
	cli, ok := v.(*client)
	if !ok {
		slog.Error("clientByConn:*client")
		panic("YOURE WRONG GGINA")
		// return nil, false
	}
	return cli, true
}

func sendMessage(conn *fiberWS.Conn, v any) {
	go func() {
		err := conn.Conn.WriteJSON(v)
		if err != nil {
			slog.Error("sendMessage", "error", err, "v", v)
		}
	}()
}
