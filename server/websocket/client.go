package websocket

import (
	"errors"
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

func newClient() *client {
	return &client{
		uid: genUid(),
		// https://linguavolt.com/how-many-letters-are-in-an-average-word/
		// English avg word length is 4.793, lets double that
		keystrokes: make(game.Keystrokes, 0, game.WORDS_LENGTH*10),
	}
}

func (c *client) appendKeyStrokes(ks game.Keystrokes) error {
	if len(c.keystrokes)+len(ks) > cap(c.keystrokes) {
		return errors.New("You've written more keystrokes than humanly possible")
	}
	c.Lock()
	defer c.Unlock()
	c.keystrokes = append(c.keystrokes, ks...)
	return nil
}

func (c *client) String() string {
	str := c.uid + " "
	for _, v := range c.keystrokes {
		str += strconv.QuoteRune(v.Key)
	}
	return str
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
				panic(invalidTypeMsg(key, &fiberWS.Conn{}))
			}
			if !ok1 {
				panic(invalidTypeMsg(value, &client{}))
			}
			return yield(conn, cli)
		})
	}
}

func (s cliStore) String() string {
	str := ""
	for _, cli := range s.All() {
		str += cli.String() + ";;"
	}
	return str
}

func (s cliStore) resetClientKeystrokes() {
	for _, cli := range s.All() {
		cli.Lock()
		defer cli.Unlock()
		cli.keystrokes = cli.keystrokes[:0]
	}
}

func (s cliStore) clientCount() int {
	count := 0
	for range s.All() {
		count += 1
	}
	return count
}
