package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Store interface {
	Ping(ctx context.Context) error
	CreateShortURL(ctx context.Context, code, originalURL string, ttl time.Duration) error
	GetOriginalURL(ctx context.Context, code string) (string, error)
	IncrementClicks(ctx context.Context, code string) error
	AllowRequest(ctx context.Context, ip string, maxPerMinute int) (bool, error)
}

type Dependencies struct {
	Store            Store
	BaseURL          string
	DefaultCodeLen   int
	RateLimitPerMin  int
	ReservedKeywords map[string]struct{}
	Logger           *slog.Logger
}

func NewRouter(deps Dependencies) http.Handler {
	h := newHandler(deps)

	r := chi.NewRouter()
	r.Use(h.rateLimitMiddleware)
	r.Get("/health", h.health)
	r.Post("/api/shorten", h.shorten)
	r.Get("/{code}", h.redirect)
	return r
}
