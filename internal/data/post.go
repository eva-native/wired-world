package data

import (
	"errors"
	"strings"
	"time"
)

var ErrTooLongMessage = errors.New("message too long")
var ErrEmptyMessage = errors.New("message empty")

type Post struct {
	Number  uint
	Time    time.Time
	Message string
}

func NewPost(n uint, t time.Time, m string) Post {
	return Post{
		Number:  n,
		Time:    t,
		Message: m,
	}
}

func (p *Post) FormatTime() string {
	return p.Time.Format(time.RFC3339)
}

func PrepareMessage(m string) string {
	return strings.TrimSpace(m)
}

func ValidateMessage(m string) error {
	switch s := len(m); {
	case s == 0: return ErrEmptyMessage
	case s > 0x80: return ErrTooLongMessage
	}
	return nil
}
