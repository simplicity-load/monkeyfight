package game

import (
	"math/rand/v2"
	"strings"
)

type State string

const (
	Wait    State = "wait"
	Playing       = "playing"
)

type Game struct {
	Words string `json:"w"`
	State State  `json:"s"`
}

const WORDS_LENGTH = 300 // 300 WPM -> 150 WP30s
func New() *Game {
	return &Game{
		Words: genWords(WORDS_LENGTH),
		State: Wait,
	}
}

func (g *Game) IsPlaying() bool {
	return g.State == Playing
}

func (g *Game) Start() {
	g.State = Playing
}

func genWords(length int) string {
	words := make([]string, 0, length)
	prev := -1
	for range cap(words) {
		i := uniqueIndex(prev)
		prev = i
		words = append(words, wordList[i])
	}
	return strings.Join(words, " ")
}

func uniqueIndex(prev int) int {
	i := -1
	for {
		i = rand.IntN(len(wordList))
		if prev != i {
			break
		}
	}
	return i
}
