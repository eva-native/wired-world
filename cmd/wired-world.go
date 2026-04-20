package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/time/rate"

	"github.com/eva-native/wired-world/internal/handlers"
	"github.com/eva-native/wired-world/internal/middleware"
	"github.com/eva-native/wired-world/internal/repository"
	"github.com/eva-native/wired-world/web"
)

func main() {
	redisAddr := flag.String("redis", "localhost:6379", "Redis address host:port")
	addr := flag.String("addr", ":8080", "HTTP server listen address")
	behindProxy := flag.Bool("behind-proxy", false, "Trust X-Real-IP / X-Forwarded-For headers for rate limiting")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	rdb, err := repository.NewPostRedis(ctx, *redisAddr)
	if err != nil {
		log.Fatalln("Redis error:", err)
	}
	defer rdb.Close()
	log.Printf("Redis open: %s", *redisAddr)

	rl := middleware.NewRateLimiter(rate.Every(5*time.Second), 2, *behindProxy)
	go rl.Cleanup(ctx, 10*time.Minute)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(web.Content()))
	mux.Handle("/post", handlers.AllPost(&rdb))
	mux.Handle("POST /post", rl.Middleware(handlers.AddNewPost(&rdb)))

	if err := listenAndServe(ctx, *addr, mux); err != nil {
		log.Printf("Server error: %v", err)
	}
}

func listenAndServe(ctx context.Context, addr string, mux *http.ServeMux) error {
	srv := &http.Server{Addr: addr, Handler: mux}

	errch := make(chan error, 1)
	go func() {
		log.Printf("Starting server on %s...", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errch <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			srv.Close()
			return err
		}
		return nil
	case err := <-errch:
		return err
	}
}
