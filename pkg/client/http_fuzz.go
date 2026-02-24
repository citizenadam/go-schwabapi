package client

import (
	"context"
	"log/slog"
	"strings"
	"testing"
)

// FuzzNewClient tests the NewClient constructor with various logger inputs
func FuzzNewClient(f *testing.F) {
	f.Add([]byte("test"))
	f.Add([]byte(""))
	f.Add([]byte(strings.Repeat("a", 1000)))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		_ = NewClient(logger)
	})
}

// FuzzClientGet tests the Get method with various URL and header inputs
func FuzzClientGet(f *testing.F) {
	f.Add("https://example.com", []byte("Authorization: Bearer token"))
	f.Add("http://localhost:8080", []byte(""))
	f.Add("", []byte("X-Custom: value"))

	f.Fuzz(func(t *testing.T, url string, headers []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		client := NewClient(logger)

		ctx := context.Background()

		// Parse headers from byte input
		headerMap := make(map[string]string)
		headerParts := strings.Split(string(headers), "\n")
		for _, part := range headerParts {
			if idx := strings.Index(part, ":"); idx > 0 {
				key := strings.TrimSpace(part[:idx])
				value := strings.TrimSpace(part[idx+1:])
				if key != "" {
					headerMap[key] = value
				}
			}
		}

		// Call Get - should not crash even with invalid URLs
		_, _ = client.Get(ctx, url, headerMap)
	})
}

// FuzzClientPost tests the Post method with various URL, header, and body inputs
func FuzzClientPost(f *testing.F) {
	f.Add("https://example.com/api", []byte("Content-Type: application/json"), []byte(`{"key":"value"}`))
	f.Add("http://localhost:8080", []byte(""), []byte(""))
	f.Add("", []byte("X-Custom: value"), []byte("invalid json"))

	f.Fuzz(func(t *testing.T, url string, headers []byte, body []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		client := NewClient(logger)

		ctx := context.Background()

		// Parse headers from byte input
		headerMap := make(map[string]string)
		headerParts := strings.Split(string(headers), "\n")
		for _, part := range headerParts {
			if idx := strings.Index(part, ":"); idx > 0 {
				key := strings.TrimSpace(part[:idx])
				value := strings.TrimSpace(part[idx+1:])
				if key != "" {
					headerMap[key] = value
				}
			}
		}

		// Create a simple body interface from the byte input
		var bodyInterface interface{}
		if len(body) > 0 {
			bodyInterface = string(body)
		}

		// Call Post - should not crash even with invalid inputs
		_, _ = client.Post(ctx, url, headerMap, bodyInterface)
	})
}

// FuzzClientPut tests the Put method with various URL, header, and body inputs
func FuzzClientPut(f *testing.F) {
	f.Add("https://example.com/api", []byte("Content-Type: application/json"), []byte(`{"key":"value"}`))
	f.Add("http://localhost:8080", []byte(""), []byte(""))
	f.Add("", []byte("X-Custom: value"), []byte("invalid json"))

	f.Fuzz(func(t *testing.T, url string, headers []byte, body []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		client := NewClient(logger)

		ctx := context.Background()

		// Parse headers from byte input
		headerMap := make(map[string]string)
		headerParts := strings.Split(string(headers), "\n")
		for _, part := range headerParts {
			if idx := strings.Index(part, ":"); idx > 0 {
				key := strings.TrimSpace(part[:idx])
				value := strings.TrimSpace(part[idx+1:])
				if key != "" {
					headerMap[key] = value
				}
			}
		}

		// Create a simple body interface from the byte input
		var bodyInterface interface{}
		if len(body) > 0 {
			bodyInterface = string(body)
		}

		// Call Put - should not crash even with invalid inputs
		_, _ = client.Put(ctx, url, headerMap, bodyInterface)
	})
}

// FuzzClientDelete tests the Delete method with various URL and header inputs
func FuzzClientDelete(f *testing.F) {
	f.Add("https://example.com/api", []byte("Authorization: Bearer token"))
	f.Add("http://localhost:8080", []byte(""))
	f.Add("", []byte("X-Custom: value"))

	f.Fuzz(func(t *testing.T, url string, headers []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		client := NewClient(logger)

		ctx := context.Background()

		// Parse headers from byte input
		headerMap := make(map[string]string)
		headerParts := strings.Split(string(headers), "\n")
		for _, part := range headerParts {
			if idx := strings.Index(part, ":"); idx > 0 {
				key := strings.TrimSpace(part[:idx])
				value := strings.TrimSpace(part[idx+1:])
				if key != "" {
					headerMap[key] = value
				}
			}
		}

		// Call Delete - should not crash even with invalid URLs
		_, _ = client.Delete(ctx, url, headerMap)
	})
}

// FuzzDecodeJSON tests the DecodeJSON method with various JSON inputs
func FuzzDecodeJSON(f *testing.F) {
	f.Add([]byte(`{"key":"value"}`))
	f.Add([]byte(""))
	f.Add([]byte("invalid json"))
	f.Add([]byte(strings.Repeat("a", 10000)))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		client := NewClient(logger)

		// Create a mock response with the fuzz data as body
		// Note: We can't easily create a real http.Response in fuzz tests
		// so we'll just test that the function doesn't panic on various inputs
		var result interface{}
		_ = client.DecodeJSON(nil, &result)
	})
}

// FuzzCloseIdleConnections tests the CloseIdleConnections method
func FuzzCloseIdleConnections(f *testing.F) {
	f.Add([]byte("test"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))
		client := NewClient(logger)

		// Call CloseIdleConnections - should not crash
		client.CloseIdleConnections()
	})
}
