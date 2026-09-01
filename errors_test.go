package schwabdev

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

// TestSchwabAPIError_Is verifies errors.Is matches the sentinel for the
// status class of a SchwabAPIError, and does not match non-matching statuses.
func TestSchwabAPIError_Is(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		target   error
		expected bool
	}{
		{"400 bad request", http.StatusBadRequest, ErrBadRequest, true},
		{"400 not unauthorized", http.StatusBadRequest, ErrUnauthorized, false},
		{"401 unauthorized", http.StatusUnauthorized, ErrUnauthorized, true},
		{"403 forbidden", http.StatusForbidden, ErrForbidden, true},
		{"404 not found", http.StatusNotFound, ErrNotFound, true},
		{"429 rate limited", http.StatusTooManyRequests, ErrRateLimited, true},
		{"429 not bad request", http.StatusTooManyRequests, ErrBadRequest, false},
		{"500 server", http.StatusInternalServerError, ErrServer, true},
		{"503 server", http.StatusServiceUnavailable, ErrServer, true},
		{"499 not server", 499, ErrServer, false},
		{"unknown target", http.StatusBadRequest, ErrStreamerUnavailable, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &SchwabAPIError{StatusCode: tc.status}
			if got := errors.Is(e, tc.target); got != tc.expected {
				t.Fatalf("errors.Is(SchwabAPIError{%d}, %v) = %v, want %v",
					tc.status, tc.target, got, tc.expected)
			}
		})
	}
}

// TestParseErrorBody covers the tolerated Schwab error body shapes.
func TestParseErrorBody(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		wantMessage string
		wantCode    string
	}{
		{"message only", `{"message":"bad thing"}`, "bad thing", ""},
		{"error key", `{"error":"nope"}`, "nope", ""},
		{"error_description", `{"error_description":"oops"}`, "oops", ""},
		{"errors array", `{"errors":[{"message":"not allowed","code":"1000"}]}`, "not allowed", "1000"},
		{"raw text fallback", `plain text body`, "plain text body", ""},
		{"empty body", ``, "", ""},
		{"non-object json", `[1,2,3]`, "[1,2,3]", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, code := parseErrorBody([]byte(tc.body))
			if msg != tc.wantMessage {
				t.Fatalf("message = %q, want %q", msg, tc.wantMessage)
			}
			if code != tc.wantCode {
				t.Fatalf("code = %q, want %q", code, tc.wantCode)
			}
		})
	}
}

// TestDoRequest_CapturesRequestIDAndMessage verifies doRequest populates the
// RequestID and Message fields from a non-2xx response.
func TestDoRequest_CapturesRequestIDAndMessage(t *testing.T) {
	client, _ := newQuotesClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("schwab-client-correl-id", "correl-123")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid symbol"}`))
	})

	_, err := client.Quotes(context.Background(), "SPX", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *SchwabAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *SchwabAPIError, got %T", err)
	}
	if apiErr.RequestID != "correl-123" {
		t.Fatalf("RequestID = %q, want %q", apiErr.RequestID, "correl-123")
	}
	if apiErr.Message != "invalid symbol" {
		t.Fatalf("Message = %q, want %q", apiErr.Message, "invalid symbol")
	}
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected errors.Is(err, ErrBadRequest)")
	}
}

// TestDoRequest_ParsesRetryAfter verifies Retry-After is parsed into a
// duration on 429 responses, for both the integer-seconds and HTTP-date forms.
func TestDoRequest_ParsesRetryAfter(t *testing.T) {
	t.Run("integer seconds", func(t *testing.T) {
		client, _ := newQuotesClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"slow down"}`))
		})
		_, err := client.Quotes(context.Background(), "SPX", nil, nil)
		var apiErr *SchwabAPIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *SchwabAPIError, got %T", err)
		}
		if apiErr.RetryAfter != 5*time.Second {
			t.Fatalf("RetryAfter = %v, want 5s", apiErr.RetryAfter)
		}
		if !errors.Is(err, ErrRateLimited) {
			t.Fatalf("expected errors.Is(err, ErrRateLimited)")
		}
	})

	t.Run("http date", func(t *testing.T) {
		client, _ := newQuotesClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", time.Now().Add(3*time.Second).UTC().Format(http.TimeFormat))
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{}`))
		})
		_, err := client.Quotes(context.Background(), "SPX", nil, nil)
		var apiErr *SchwabAPIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *SchwabAPIError, got %T", err)
		}
		if apiErr.RetryAfter <= 0 || apiErr.RetryAfter > 4*time.Second {
			t.Fatalf("RetryAfter = %v, want ~3s", apiErr.RetryAfter)
		}
	})

	t.Run("absent header", func(t *testing.T) {
		client, _ := newQuotesClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{}`))
		})
		_, err := client.Quotes(context.Background(), "SPX", nil, nil)
		var apiErr *SchwabAPIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *SchwabAPIError, got %T", err)
		}
		if apiErr.RetryAfter != 0 {
			t.Fatalf("RetryAfter = %v, want 0", apiErr.RetryAfter)
		}
	})
}

// TestParseErrorBody_Correlates ensures a 429 message is surfaced and matched.
func TestDoRequest_429MessageAndMatch(t *testing.T) {
	client, _ := newQuotesClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"message":"rate limited","code":"429"}]}`))
	})
	_, err := client.Quotes(context.Background(), "SPX", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *SchwabAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *SchwabAPIError, got %T", err)
	}
	if apiErr.Message != "rate limited" {
		t.Fatalf("Message = %q, want %q", apiErr.Message, "rate limited")
	}
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected errors.Is(err, ErrRateLimited)")
	}
}
