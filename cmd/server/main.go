package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"linkshortener/internal/api"
	"linkshortener/internal/config"
	"linkshortener/internal/store"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	redisStore := store.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer redisStore.Close()

	router := api.NewRouter(api.Dependencies{
		Store:            redisStore,
		BaseURL:          cfg.BaseURL,
		DefaultCodeLen:   cfg.CodeLength,
		RateLimitPerMin:  cfg.RateLimitPerMinute,
		ReservedKeywords: cfg.ReservedKeywords,
		Logger:           logger,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}
