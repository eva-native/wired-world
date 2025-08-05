package data

import (
	"errors"
	"time"
)

var ErrTooLongMessage = errors.New("message too long")

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

func ValidateMessage(m string) error {
	if len(m) > 0x80 {
		return ErrTooLongMessage
	}
	return nil
}
