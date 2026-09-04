package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gongcha-cup/backend/internal/domain"
)

func TestWriteDomainError_HidesPostgresText(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code int
	}{
		{
			name: "conflict wrap with unique violation text",
			err:  fmt.Errorf("%w: duplicate key value violates unique constraint \"products_pkey\"", domain.ErrConflict),
			code: http.StatusConflict,
		},
		{
			name: "unmapped driver error",
			err:  errors.New("ERROR: permission denied for table products (SQLSTATE 42501)"),
			code: http.StatusInternalServerError,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeDomainError(rec, tc.err)
			if rec.Code != tc.code {
				t.Fatalf("status: got %d want %d body=%s", rec.Code, tc.code, rec.Body)
			}
			body := rec.Body.String()
			for _, leak := range []string{
				"duplicate key",
				"products_pkey",
				"SQLSTATE",
				"permission denied",
				"violates",
			} {
				if strings.Contains(body, leak) {
					t.Fatalf("leaked %q: %s", leak, body)
				}
			}
		})
	}
}

func TestWriteDomainError_KeepsDomainDetail(t *testing.T) {
	rec := httptest.NewRecorder()
	writeDomainError(rec, fmt.Errorf("%w: MP010 has 0, needs 100", domain.ErrInsufficient))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status: got %d want 422", rec.Code)
	}
	var env errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("json: %v", err)
	}
	if env.Error.Code != "insufficient_inventory" {
		t.Fatalf("code: %q", env.Error.Code)
	}
	if !strings.Contains(env.Error.Message, "MP010") {
		t.Fatalf("want domain detail, got %q", env.Error.Message)
	}
}

func TestWriteDomainError_UnmappedIsOpaque(t *testing.T) {
	rec := httptest.NewRecorder()
	writeDomainError(rec, errors.New("dial tcp 127.0.0.1:5432: connect: connection refused"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want 500", rec.Code)
	}
	var env errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("json: %v", err)
	}
	if env.Error.Message != "internal error" {
		t.Fatalf("want opaque 500, got %q", env.Error.Message)
	}
}
