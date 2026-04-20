package data

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateMessage(t *testing.T) {
	cases := []struct {
		input string
		want  error
	}{
		{"", ErrEmptyMessage},
		{PrepareMessage("   "), ErrEmptyMessage},
		{strings.Repeat("a", 129), ErrTooLongMessage},
		{strings.Repeat("a", 128), nil},
		{"hello", nil},
	}
	for _, c := range cases {
		got := ValidateMessage(c.input)
		if !errors.Is(got, c.want) {
			t.Errorf("ValidateMessage(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestPrepareMessage(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"  hi", "hi"},
		{"hi  ", "hi"},
		{"  hi  ", "hi"},
		{"hi", "hi"},
		{"\nhello\n", "hello"},
	}
	for _, c := range cases {
		got := PrepareMessage(c.input)
		if got != c.want {
			t.Errorf("PrepareMessage(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
