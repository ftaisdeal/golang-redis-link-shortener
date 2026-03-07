package model

import "time"

type ShortenRequest struct {
	URL              string `json:"url"`
	CustomCode       string `json:"custom_code,omitempty"`
	ExpiresInSeconds int64  `json:"expires_in_seconds,omitempty"`
}

type ShortenResponse struct {
	Code      string     `json:"code"`
	ShortURL  string     `json:"short_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
