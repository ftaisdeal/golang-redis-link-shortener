package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"linkshortener/internal/model"
	"linkshortener/internal/store"
	"linkshortener/internal/util"
)

type handler struct {
	store            Store
	baseURL          string
	defaultCodeLen   int
	rateLimitPerMin  int
	reservedKeywords map[string]struct{}
	logger           *slog.Logger
}

func newHandler(deps Dependencies) *handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &handler{
		store:            deps.Store,
		baseURL:          strings.TrimSuffix(deps.BaseURL, "/"),
		defaultCodeLen:   deps.DefaultCodeLen,
		rateLimitPerMin:  deps.RateLimitPerMin,
		reservedKeywords: deps.ReservedKeywords,
		logger:           logger,
	}
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.store.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, model.ErrorResponse{Error: "redis unavailable"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) shorten(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid JSON body"})
		return
	}

	normalizedURL, err := util.ValidateAndNormalizeURL(req.URL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	if req.ExpiresInSeconds < 0 {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "expires_in_seconds must be >= 0"})
		return
	}

	var ttl time.Duration
	if req.ExpiresInSeconds > 0 {
		ttl = time.Duration(req.ExpiresInSeconds) * time.Second
	}

	code := strings.TrimSpace(req.CustomCode)
	if code != "" {
		if err := util.ValidateCustomCode(code); err != nil {
			writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}
		if _, reserved := h.reservedKeywords[strings.ToLower(code)]; reserved {
			writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "custom_code is reserved"})
			return
		}
		err = h.store.CreateShortURL(r.Context(), code, normalizedURL, ttl)
		if errors.Is(err, store.ErrCodeAlreadyExists) {
			writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: "custom_code already exists"})
			return
		}
		if err != nil {
			h.logger.Error("failed to create custom code", "error", err)
			writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
			return
		}
	} else {
		for i := 0; i < 5; i++ {
			generated, genErr := util.GenerateCode(h.defaultCodeLen)
			if genErr != nil {
				h.logger.Error("failed to generate code", "error", genErr)
				writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
				return
			}
			if _, reserved := h.reservedKeywords[strings.ToLower(generated)]; reserved {
				continue
			}
			err = h.store.CreateShortURL(r.Context(), generated, normalizedURL, ttl)
			if errors.Is(err, store.ErrCodeAlreadyExists) {
				continue
			}
			if err != nil {
				h.logger.Error("failed to create short url", "error", err)
				writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
				return
			}
			code = generated
			break
		}
		if code == "" {
			writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to generate unique code"})
			return
		}
	}

	resp := model.ShortenResponse{
		Code:     code,
		ShortURL: fmt.Sprintf("%s/%s", h.baseURL, code),
	}
	if ttl > 0 {
		expires := time.Now().UTC().Add(ttl)
		resp.ExpiresAt = &expires
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *handler) redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.NotFound(w, r)
		return
	}

	originalURL, err := h.store.GetOriginalURL(r.Context(), code)
	if errors.Is(err, store.ErrCodeNotFound) {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "short code not found"})
		return
	}
	if err != nil {
		h.logger.Error("failed to get original url", "error", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}

	if err := h.store.IncrementClicks(r.Context(), code); err != nil {
		h.logger.Warn("failed to increment clicks", "code", code, "error", err)
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
