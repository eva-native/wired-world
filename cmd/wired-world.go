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

const (
	DBPath = "./wired.db"
)

var dsn = flag.String("db", ":memory:", "Sqlite3 DSN")
var addr = flag.String("addr", ":8080", "HTTP server listen address")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	db, err := repository.NewPostDB(ctx, *dsn)
	if err != nil {
		log.Fatalln("Database error: ", err)
	}
	defer db.DB.Close()
	log.Printf("Database open: %s", *dsn)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(web.Content)))
	mux.Handle("/post", handlers.AllPost(&db))
	mux.Handle("POST /post", handlers.AddNewPost(&db))

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
