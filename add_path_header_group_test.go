package add_path_header_group

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddPathHeader_SetsConfiguredHeader(t *testing.T) {
	cfg := CreateConfig()
	cfg.HeaderName = "X-Custom-Path"

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		got := req.Header.Get("X-Custom-Path")
		if got != "/some/path" {
			t.Errorf("expected header X-Custom-Path to be /some/path, got %q", got)
		}
	})

	handler, err := New(context.Background(), next, cfg, "test-middleware")
	if err != nil {
		t.Fatalf("unexpected error creating middleware: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)
}

func TestAddPathHeader_DefaultHeaderName(t *testing.T) {
	cfg := CreateConfig()

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		got := req.Header.Get("x-path-group")
		if got != "/default/path" {
			t.Errorf("expected header x-path-group to be /default/path, got %q", got)
		}
	})

	handler, err := New(context.Background(), next, cfg, "test-middleware")
	if err != nil {
		t.Fatalf("unexpected error creating middleware: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/default/path", nil)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)
}

func TestAddPathHeader_EmptyHeaderNameFallsBackToDefault(t *testing.T) {
	cfg := CreateConfig()
	cfg.HeaderName = ""

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		got := req.Header.Get("x-path-group")
		if got != "/fallback" {
			t.Errorf("expected header x-path-group to be /fallback, got %q", got)
		}
	})

	handler, err := New(context.Background(), next, cfg, "test-middleware")
	if err != nil {
		t.Fatalf("unexpected error creating middleware: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/fallback", nil)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)
}

func TestAddPathHeader_ExtractsPathGroup(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "UUID replacement",
			path:     "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/profile",
			expected: "/api/v1/users/*/profile",
		},
		{
			name:     "Numeric ID replacement",
			path:     "/api/v1/courts/42/bookings",
			expected: "/api/v1/courts/*/bookings",
		},
		{
			name:     "Alphanumeric slug replacement",
			path:     "/api/v1/bookings/booking-abc-99/details",
			expected: "/api/v1/bookings/*/details",
		},
		{
			name:     "Slug with underscore",
			path:     "/api/v1/users/user_42/profile",
			expected: "/api/v1/users/*/profile",
		},
		{
			name:     "Mixed IDs",
			path:     "/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000/courts/42/bookings/booking-abc-99",
			expected: "/api/v1/tenants/*/courts/*/bookings/*",
		},
		{
			name:     "Non-ID segments preserved",
			path:     "/api/v1/users/profile",
			expected: "/api/v1/users/profile",
		},
		{
			name:     "Version prefix preserved",
			path:     "/api/v1/users/123",
			expected: "/api/v1/users/*",
		},
		{
			name:     "Plain words preserved",
			path:     "/api/users/profile",
			expected: "/api/users/profile",
		},
		{
			name:     "Root path",
			path:     "/",
			expected: "/",
		},
		{
			name:     "Multiple numeric IDs",
			path:     "/api/v1/123/456/789",
			expected: "/api/v1/*/*/*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := CreateConfig()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				got := req.Header.Get("x-path-group")
				if got != tt.expected {
					t.Errorf("expected path group %q, got %q", tt.expected, got)
				}
			})

			handler, err := New(context.Background(), next, cfg, "test-middleware")
			if err != nil {
				t.Fatalf("unexpected error creating middleware: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)
		})
	}
}
