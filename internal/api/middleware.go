package api

import (
	"net"
	"net/http"
	"strings"

	"linkshortener/internal/model"
)

func (h *handler) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.rateLimitPerMin <= 0 {
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r)
		allowed, err := h.store.AllowRequest(r.Context(), ip, h.rateLimitPerMin)
		if err != nil {
			h.logger.Warn("rate limiter failed open", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		if !allowed {
			writeJSON(w, http.StatusTooManyRequests, model.ErrorResponse{Error: "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
