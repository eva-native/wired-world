package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/eva-native/wired-world/internal/data"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbDriver = "sqlite3"
)

type PostDB struct {
	DB *sql.DB
}

func NewPostDB(ctx context.Context, dsn string) (PostDB, error) {
	db, err := openAndInitDB(ctx, dsn)
	return PostDB{DB: db}, err
}

func (p *PostDB) All(ctx context.Context) ([]data.Post, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT number, create_at, message FROM posts;`)
	if err != nil {
		return []data.Post{}, err
	}
	defer rows.Close()

	posts := make([]data.Post, 0)

	for rows.Next() {
		var num uint
		var ts int64
		var msg string
		if err := rows.Scan(&num, &ts, &msg); err != nil {
			return []data.Post{}, err
		}
		posts = append(posts, data.NewPost(num, time.Unix(ts, 0), msg))
	}

	if err := rows.Err(); err != nil {
		return []data.Post{}, err
	}

	return posts, nil
}

func (p *PostDB) Add(ctx context.Context, t time.Time, msg string) (data.Post, error) {
	ts := t.Unix()
	var num uint
	err := p.DB.QueryRowContext(ctx, `INSERT INTO posts(create_at, message) VALUES (?, ?) RETURNING number;`, ts, msg).Scan(&num)
	if err != nil {
		return data.Post{}, err
	}
	return data.NewPost(num, time.Unix(ts, 0), msg), nil
}

func openAndInitDB(ctx context.Context, dsn string) (*sql.DB, error) {
	r, err := sql.Open(dbDriver, dsn)
	if err != nil {
		return nil, err
	}

	if err := pingWithTimeout(ctx, r); err != nil {
		return nil, err
	}

	if err := initTable(ctx, r); err != nil {
		return nil, err
	}

	return r, nil
}

func pingWithTimeout(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	return db.PingContext(ctx)
}

func initTable(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	_, err := db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS posts (
		number INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		create_at INTEGER DEFAULT CURRENT_TIMESTAMP,
		message TEXT NOT NULL
	);
	`)
	return err
}
