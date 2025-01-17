package websocket

import (
	"fmt"
	"iter"
	"math/rand/v2"
	"reflect"
	"strconv"
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

type cliStore struct {
	*sync.Map
}

type sKey = *fiberWS.Conn
type sVal = *client

func invalidTypeMsg(x any, y any) string {
	xType := reflect.TypeOf(x)
	yType := reflect.TypeOf(y)
	return fmt.Sprintf("Type is %s, must be %s", xType, yType)
}

func (s cliStore) Store(k sKey, v sVal) {
	s.Map.Store(k, v)
}

func (s cliStore) Delete(k sKey) {
	s.Map.Delete(k)
}

func (s cliStore) Load(k sKey) (sVal, bool) {
	val, ok := s.Map.Load(k)
	if !ok {
		return nil, false
	}
	cli, ok := val.(*client)
	if !ok {
		invalidTypeMsg(cli, &client{})
	}
	return cli, true
}

func (s cliStore) All() iter.Seq2[sKey, sVal] {
	return func(yield func(sKey, sVal) bool) {
		s.Range(func(key, value any) bool {
			conn, ok := key.(sKey)
			cli, ok1 := value.(sVal)
			if !ok {
				panic(invalidTypeMsg(conn, &fiberWS.Conn{}))
			}
			if !ok1 {
				panic(invalidTypeMsg(conn, &client{}))
			}
			return yield(conn, cli)
		})
	}
}

func (s cliStore) String() string {
	str := ""
	for _, cli := range s.All() {
		str += cli.uid + " "
		for _, v := range cli.keystrokes {
			str += strconv.QuoteRune(v.Key)
		}
		str += " "
	}
	return str
}

func (s cliStore) resetClientKeystrokes() {
	for _, cli := range s.All() {
		cli.Lock()
		defer cli.Unlock()
		cli.keystrokes = game.Keystrokes{}
	}
}

func (s cliStore) clientCount() int {
	count := 0
	for range s.All() {
		count += 1
	}
	return count
}
