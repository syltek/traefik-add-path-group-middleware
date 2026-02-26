package traefik_add_path_group_middleware

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
			expected: "/api/v1/users/uuid/profile",
		},
		{
			name:     "Numeric ID replacement",
			path:     "/api/v1/courts/42/bookings",
			expected: "/api/v1/courts/numeric_id/bookings",
		},
		{
			name:     "Alphanumeric slug replacement",
			path:     "/api/v1/bookings/booking-abc-99/details",
			expected: "/api/v1/bookings/slug/details",
		},
		{
			name:     "Slug with underscore",
			path:     "/api/v1/users/user_42/profile",
			expected: "/api/v1/users/slug/profile",
		},
		{
			name:     "Mixed IDs",
			path:     "/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000/courts/42/bookings/booking-abc-99",
			expected: "/api/v1/tenants/uuid/courts/numeric_id/bookings/slug",
		},
		{
			name:     "Non-ID segments preserved",
			path:     "/api/v1/users/profile",
			expected: "/api/v1/users/profile",
		},
		{
			name:     "Version prefix preserved",
			path:     "/api/v1/users/123",
			expected: "/api/v1/users/numeric_id",
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
			expected: "/api/v1/numeric_id/numeric_id/numeric_id",
		},
		{
			name:     "ULID replacement",
			path:     "/api/v1/users/01ARZ3NDEKTSV4RRFFQ69G5FAV/profile",
			expected: "/api/v1/users/ulid/profile",
		},
		{
			name:     "CUID replacement",
			path:     "/api/v1/users/clh3am1g30000udocl363eofy/profile",
			expected: "/api/v1/users/cuid/profile",
		},
		{
			name:     "CUID2 replacement",
			path:     "/api/v1/users/tz4a98xxat96iws9zmbrgj3a/profile",
			expected: "/api/v1/users/cuid2/profile",
		},
		{
			name:     "NanoID replacement",
			path:     "/api/v1/users/V1StGXR8_Z5jdHi6B-myT/profile",
			expected: "/api/v1/users/nanoid/profile",
		},
		{
			name:     "Prefixed UUID with colon",
			path:     "/api/v1/users/usr:550e8400-e29b-41d4-a716-446655440000/profile",
			expected: "/api/v1/users/uuid/profile",
		},
		{
			name:     "Prefixed UUID with underscore",
			path:     "/api/v1/users/usr_550e8400-e29b-41d4-a716-446655440000/profile",
			expected: "/api/v1/users/uuid/profile",
		},
		{
			name:     "Prefixed numeric ID with colon",
			path:     "/api/v1/courts/court:12345/bookings",
			expected: "/api/v1/courts/numeric_id/bookings",
		},
		{
			name:     "Prefixed numeric ID with underscore",
			path:     "/api/v1/courts/court_12345/bookings",
			expected: "/api/v1/courts/numeric_id/bookings",
		},
		{
			name:     "Prefixed NanoID with colon",
			path:     "/api/v1/users/usr:V1StGXR8_Z5jdHi6B-myT/profile",
			expected: "/api/v1/users/nanoid/profile",
		},
		{
			name:     "Prefixed NanoID with underscore",
			path:     "/api/v1/users/usr_V1StGXR8_Z5jdHi6B-myT/profile",
			expected: "/api/v1/users/nanoid/profile",
		},
		{
			name:     "Prefixed ULID with colon",
			path:     "/api/v1/users/usr:01ARZ3NDEKTSV4RRFFQ69G5FAV/profile",
			expected: "/api/v1/users/ulid/profile",
		},
		{
			name:     "Prefixed CUID with colon",
			path:     "/api/v1/users/usr:clh3am1g30000udocl363eofy/profile",
			expected: "/api/v1/users/cuid/profile",
		},
		{
			name:     "Prefixed CUID2 with colon",
			path:     "/api/v1/users/usr:tz4a98xxat96iws9zmbrgj3a/profile",
			expected: "/api/v1/users/cuid2/profile",
		},
		{
			name:     "Mixed prefixed and non-prefixed IDs",
			path:     "/api/v1/tenants/usr:550e8400-e29b-41d4-a716-446655440000/courts/42/bookings/booking-abc-99",
			expected: "/api/v1/tenants/uuid/courts/numeric_id/bookings/slug",
		},
		{
			name:     "Segment with colon but invalid prefix",
			path:     "/api/v1/users/not-a-prefix:invalid-id/profile",
			expected: "/api/v1/users/not-a-prefix:invalid-id/profile",
		},
		{
			name:     "Segment with underscore but invalid prefix",
			path:     "/api/v1/users/not_a_prefix_123/profile",
			expected: "/api/v1/users/slug/profile",
		},
		{
			name:     "Short string that doesn't match any ID type",
			path:     "/api/v1/users/abc/profile",
			expected: "/api/v1/users/abc/profile",
		},
		{
			name:     "NanoID with underscore in middle (not prefix)",
			path:     "/api/v1/users/V1StGXR8_Z5jdHi6B-myT/profile",
			expected: "/api/v1/users/nanoid/profile",
		},
		{
			name:     "All ID types in one path",
			path:     "/api/v1/123/550e8400-e29b-41d4-a716-446655440000/01ARZ3NDEKTSV4RRFFQ69G5FAV/clh3am1g30000udocl363eofy/tz4a98xxat96iws9zmbrgj3a/V1StGXR8_Z5jdHi6B-myT",
			expected: "/api/v1/numeric_id/uuid/ulid/cuid/cuid2/nanoid",
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
