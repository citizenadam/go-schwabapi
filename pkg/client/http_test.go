package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestNewClient(t *testing.T) {
	logger := slog.Default()
	client := NewClient(logger)

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, logger, client.logger)
	assert.Equal(t, 60*time.Second, client.httpClient.Timeout)
}

func TestClient_Get_Success(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful GET request",
			url:            "/test",
			headers:        map[string]string{"Authorization": "Bearer token"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "GET request without headers",
			url:            "/test",
			headers:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

				// Check headers
				for key, value := range tt.headers {
					assert.Equal(t, value, r.Header.Get(key))
				}

				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte(tt.expectedBody))
			}))
			defer server.Close()

			logger := slog.Default()
			client := NewClient(logger)

			ctx := context.Background()
			resp, err := client.Get(ctx, server.URL+tt.url, tt.headers)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

func TestClient_Get_Error(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid URL",
			url:         "://invalid-url",
			expectError: true,
			errorMsg:    "failed to create GET request",
		},
		{
			name:        "context cancelled",
			url:         "/test",
			expectError: true,
			errorMsg:    "GET request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			client := NewClient(logger)

			ctx := context.Background()

			if tt.name == "context cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			resp, err := client.Get(ctx, tt.url, nil)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resp)
			}
		})
	}
}

func TestClient_Post_Success(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		body           any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful POST request",
			url:            "/test",
			headers:        map[string]string{"Authorization": "Bearer token"},
			body:           map[string]string{"key": "value"},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"created"}`,
		},
		{
			name:           "POST request with nil body",
			url:            "/test",
			headers:        nil,
			body:           nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"ok"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Check headers
				for key, value := range tt.headers {
					assert.Equal(t, value, r.Header.Get(key))
				}

				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte(tt.expectedBody))
			}))
			defer server.Close()

			logger := slog.Default()
			client := NewClient(logger)

			ctx := context.Background()
			resp, err := client.Post(ctx, server.URL+tt.url, tt.headers, tt.body)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

func TestClient_Post_Error(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		body        any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid URL",
			url:         "://invalid-url",
			body:        map[string]string{"key": "value"},
			expectError: true,
			errorMsg:    "failed to create POST request",
		},
		{
			name:        "unmarshalable body",
			url:         "/test",
			body:        make(chan int),
			expectError: true,
			errorMsg:    "failed to marshal request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			client := NewClient(logger)

			ctx := context.Background()
			resp, err := client.Post(ctx, tt.url, nil, tt.body)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resp)
			}
		})
	}
}

func TestClient_Put_Success(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		body           any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful PUT request",
			url:            "/test",
			headers:        map[string]string{"Authorization": "Bearer token"},
			body:           map[string]string{"key": "value"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"updated"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Check headers
				for key, value := range tt.headers {
					assert.Equal(t, value, r.Header.Get(key))
				}

				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte(tt.expectedBody))
			}))
			defer server.Close()

			logger := slog.Default()
			client := NewClient(logger)

			ctx := context.Background()
			resp, err := client.Put(ctx, server.URL+tt.url, tt.headers, tt.body)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

func TestClient_Put_Error(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		body        any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid URL",
			url:         "://invalid-url",
			body:        map[string]string{"key": "value"},
			expectError: true,
			errorMsg:    "failed to create PUT request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			client := NewClient(logger)

			ctx := context.Background()
			resp, err := client.Put(ctx, tt.url, nil, tt.body)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resp)
			}
		})
	}
}

func TestClient_Delete_Success(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful DELETE request",
			url:            "/test",
			headers:        map[string]string{"Authorization": "Bearer token"},
			expectedStatus: http.StatusNoContent,
			expectedBody:   ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)

				// Check headers
				for key, value := range tt.headers {
					assert.Equal(t, value, r.Header.Get(key))
				}

				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte(tt.expectedBody))
			}))
			defer server.Close()

			logger := slog.Default()
			client := NewClient(logger)

			ctx := context.Background()
			resp, err := client.Delete(ctx, server.URL+tt.url, tt.headers)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

func TestClient_Delete_Error(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid URL",
			url:         "://invalid-url",
			expectError: true,
			errorMsg:    "failed to create DELETE request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			client := NewClient(logger)

			ctx := context.Background()
			resp, err := client.Delete(ctx, tt.url, nil)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resp)
			}
		})
	}
}

func TestClient_DecodeJSON_Success(t *testing.T) {
	tests := []struct {
		name     string
		response string
		target   any
	}{
		{
			name:     "decode JSON object",
			response: `{"message":"success","value":42}`,
			target:   &map[string]any{},
		},
		{
			name:     "decode JSON array",
			response: `[1,2,3]`,
			target:   &[]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			client := NewClient(logger)

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}

			// Create a response with the test body
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			resp, err := http.Get(server.URL)
			require.NoError(t, err)

			err = client.DecodeJSON(resp, tt.target)
			assert.NoError(t, err)
		})
	}
}

func TestClient_DecodeJSON_Error(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		target      any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid JSON",
			response:    `{invalid json}`,
			target:      &map[string]any{},
			expectError: true,
			errorMsg:    "failed to unmarshal response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			client := NewClient(logger)

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}

			// Create a response with the test body
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			resp, err := http.Get(server.URL)
			require.NoError(t, err)

			err = client.DecodeJSON(resp, tt.target)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestClient_CloseIdleConnections(t *testing.T) {
	logger := slog.Default()
	client := NewClient(logger)

	// Should not panic
	assert.NotPanics(t, func() {
		client.CloseIdleConnections()
	})
}
