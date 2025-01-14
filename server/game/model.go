package game

type KeyEvent struct {
	Key  rune  `json:"k" validate:"required"` // TODO custom validator
	Time int64 `json:"t"`
}

type Keystrokes []KeyEvent
