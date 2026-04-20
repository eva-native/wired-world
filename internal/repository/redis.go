package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eva-native/wired-world/internal/data"
	"github.com/redis/go-redis/v9"
)

const (
	streamKey  = "posts"
	counterKey = "posts:counter"
)

type PostRedis struct {
	rdb *redis.Client
}

func NewPostRedis(ctx context.Context, addr string) (PostRedis, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return PostRedis{}, fmt.Errorf("redis ping: %w", err)
	}
	return PostRedis{rdb: rdb}, nil
}

func (p *PostRedis) Close() error {
	return p.rdb.Close()
}

func (p *PostRedis) All(ctx context.Context) ([]data.Post, error) {
	entries, err := p.rdb.XRevRange(ctx, streamKey, "+", "-").Result()
	if err != nil {
		return nil, err
	}
	posts := make([]data.Post, 0, len(entries))
	for _, e := range entries {
		t, err := parseStreamID(e.ID)
		if err != nil {
			return nil, err
		}
		num, err := strconv.ParseUint(e.Values["number"].(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse post number: %w", err)
		}
		msg, _ := e.Values["message"].(string)
		posts = append(posts, data.NewPost(uint(num), t, msg))
	}
	return posts, nil
}

func (p *PostRedis) Add(ctx context.Context, t time.Time, msg string) (data.Post, error) {
	num, err := p.rdb.Incr(ctx, counterKey).Result()
	if err != nil {
		return data.Post{}, fmt.Errorf("incr counter: %w", err)
	}
	_, err = p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]any{
			"number":  strconv.FormatInt(num, 10),
			"message": msg,
		},
	}).Result()
	if err != nil {
		return data.Post{}, fmt.Errorf("xadd: %w", err)
	}
	return data.NewPost(uint(num), t, msg), nil
}

// parseStreamID extracts a time.Time from a Redis stream entry ID.
// Entry IDs have the form "<milliseconds-unix-epoch>-<sequence>".
func parseStreamID(id string) (time.Time, error) {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("unexpected stream ID format: %q", id)
	}
	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse stream ID milliseconds: %w", err)
	}
	return time.UnixMilli(ms).UTC(), nil
}
