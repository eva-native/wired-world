package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/eva-native/wired-world/internal/handlers"
	"github.com/eva-native/wired-world/internal/repository"
	"github.com/eva-native/wired-world/web"
)

var dsn = flag.String("db", ":memory:", "SQLite DSN (used when -redis is not set)")
var redisAddr = flag.String("redis", "", "Redis address host:port")
var addr = flag.String("addr", ":8080", "HTTP server listen address")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	var posts repository.Posts

	if *redisAddr != "" {
		rdb, err := repository.NewPostRedis(ctx, *redisAddr)
		if err != nil {
			log.Fatalln("Redis error:", err)
		}
		defer rdb.Close()
		log.Printf("Redis open: %s", *redisAddr)
		posts = &rdb
	} else {
		db, err := repository.NewPostDB(ctx, *dsn)
		if err != nil {
			log.Fatalln("Database error:", err)
		}
		defer db.DB.Close()
		log.Printf("Database open: %s", *dsn)
		posts = &db
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(web.Content()))
	mux.Handle("/post", handlers.AllPost(posts))
	mux.Handle("POST /post", handlers.AddNewPost(posts))

	if err := listenAndServe(ctx, *addr, mux); err != nil {
		log.Printf("Server error: %v", err)
	}
}

func listenAndServe(ctx context.Context, addr string, mux *http.ServeMux) error {
	srv := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	errch := make(chan error, 1)

	go func() {
		log.Printf("Starting server on %s...", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errch <- err
		}
	}()

	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			srv.Close()
		}
		return err
	case err := <-errch:
		return err
	}
}
