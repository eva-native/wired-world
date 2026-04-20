package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/eva-native/wired-world/internal/handlers"
	"github.com/eva-native/wired-world/internal/middleware"
	"github.com/eva-native/wired-world/internal/repository"
	"github.com/eva-native/wired-world/web"
)

func main() {
	redisAddr := flag.String("redis", "localhost:6379", "Redis address host:port")
	addr := flag.String("addr", ":8080", "HTTP server listen address")
	metricsAddr := flag.String("metrics-addr", ":9090", "Internal address for /metrics endpoint")

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

	metricsMiddleware := middleware.NewMetrics(prometheus.DefaultRegisterer)
	chain := func(h http.Handler) http.Handler {
		return middleware.Logging(logger)(metricsMiddleware(h))
	}

	mux := http.NewServeMux()
	mux.Handle("/", chain(http.FileServerFS(web.Content())))
	mux.Handle("/post", chain(handlers.AllPost(&rdb, logger)))
	mux.Handle("POST /post", chain(handlers.AddNewPost(&rdb, logger)))

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := serve(ctx, *metricsAddr, metricsMux, logger); err != nil {
			logger.Error("metrics server error", "err", err)
			os.Exit(1)
		}
	}()

	if err := serve(ctx, *addr, mux, logger); err != nil {
		logger.Error("server error", "err", err)
	}
}

func serve(ctx context.Context, addr string, handler http.Handler, logger *slog.Logger) error {
	srv := &http.Server{Addr: addr, Handler: handler}

	errch := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", addr)
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
