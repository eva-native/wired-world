package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"

	"github.com/eva-native/wired-world/internal/handlers"
	"github.com/eva-native/wired-world/internal/middleware"
	"github.com/eva-native/wired-world/internal/repository"
	"github.com/eva-native/wired-world/web"
)

var redisAddr = flag.String("redis", "localhost:6379", "Redis address host:port")
var addr = flag.String("addr", ":8080", "HTTP server listen address")
var metricsAddr = flag.String("metrics-addr", ":9090", "Internal address for /metrics endpoint")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	rdb, err := repository.NewPostRedis(ctx, *redisAddr)
	if err != nil {
		logger.Error("redis error", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()
	logger.Info("redis open", "addr", *redisAddr)

	rl := middleware.NewRateLimiter(rate.Every(5*time.Second), 2)
	go rl.Cleanup(ctx, 10*time.Minute)

	chain := func(h http.Handler) http.Handler {
		return middleware.Logging(logger)(middleware.Metrics(h))
	}

	mux := http.NewServeMux()
	mux.Handle("/", chain(http.FileServerFS(web.Content())))
	mux.Handle("/post", chain(handlers.AllPost(&rdb, logger)))
	mux.Handle("POST /post", chain(rl.Middleware(handlers.AddNewPost(&rdb, logger))))

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Info("starting metrics server", "addr", *metricsAddr)
		if err := http.ListenAndServe(*metricsAddr, metricsMux); err != nil {
			logger.Error("metrics server error", "err", err)
		}
	}()

	if err := listenAndServe(ctx, *addr, mux, logger); err != nil {
		logger.Error("server error", "err", err)
	}
}

func listenAndServe(ctx context.Context, addr string, mux *http.ServeMux, logger *slog.Logger) error {
	srv := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	errch := make(chan error, 1)

	go func() {
		logger.Info("starting server", "addr", srv.Addr)
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
