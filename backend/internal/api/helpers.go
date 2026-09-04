package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/gongcha-cup/backend/internal/domain"
)

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}

// writeError writes a stable JSON error envelope.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

// writeDomainError maps a domain error to an HTTP envelope.
// Driver/SQL text never goes on the wire: 5xx is a fixed string, and 4xx
// falls back if the wrap still looks like Postgres.
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", publicMessage(err, "resource not found"))
	case errors.Is(err, domain.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", publicMessage(err, "conflict"))
	case errors.Is(err, domain.ErrValidation):
		writeError(w, http.StatusBadRequest, "validation_error", publicMessage(err, "validation error"))
	case errors.Is(err, domain.ErrInsufficient):
		writeError(w, http.StatusUnprocessableEntity, "insufficient_inventory", publicMessage(err, "insufficient inventory"))
	case errors.Is(err, domain.ErrCycle):
		writeError(w, http.StatusUnprocessableEntity, "dependency_cycle", publicMessage(err, "dependency cycle detected"))
	default:
		slog.Error("unmapped error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func publicMessage(err error, fallback string) string {
	if err == nil || looksLikeDriver(err.Error()) {
		return fallback
	}
	return err.Error()
}

func looksLikeDriver(msg string) bool {
	lower := strings.ToLower(msg)
	for _, n := range []string{
		"duplicate key",
		"violates",
		"sqlstate",
		"pq:",
		"permission denied",
		"current transaction is aborted",
	} {
		if strings.Contains(lower, n) {
			return true
		}
	}
	return false
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// requestLogger logs each request with id, method, path, status, and duration.
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			logger.InfoContext(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}

// corsMiddleware applies the configured CORS allowlist from CORS_ORIGIN.
func corsMiddleware() func(http.Handler) http.Handler {
	allowed := strings.Split(strings.TrimSpace(os.Getenv("CORS_ORIGIN")), ",")
	set := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		if o = strings.TrimSpace(o); o != "" {
			set[o] = struct{}{}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if _, ok := set[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key")
					w.Header().Set("Access-Control-Max-Age", "600")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
