package repository

import (
	"context"
	"time"

	"github.com/eva-native/wired-world/internal/data"
)

type Posts interface {
	All(ctx context.Context) ([]data.Post, error)
	Add(ctx context.Context, t time.Time, msg string) (data.Post, error)
}
