package middlewares_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/middlewares"
)

func TestAuthorizationMiddleware(t *testing.T) {
	apiSecret := "secret123"

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"The auth token is missing."}`,
		},
		{
			name:           "Invalid Authorization Header Format",
			authHeader:     "BearerTokenWithoutSpace",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid Authorization header."}`,
		},
		{
			name:           "Invalid Token",
			authHeader:     "Bearer invalidtoken",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token."}`,
		},
		{
			name:           "Valid Token",
			authHeader:     "Bearer secret123",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the given authorization header
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Create a test handler that writes "success" on successful execution
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Wrap the test handler with the AuthorizationMiddleware
			handler := middlewares.AuthenticationMiddleware(apiSecret, slog.Default())(nextHandler)

			// Serve the HTTP request
			handler.ServeHTTP(rr, req)

			// Check the status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check the response body
			if strings.TrimSpace(rr.Body.String()) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
