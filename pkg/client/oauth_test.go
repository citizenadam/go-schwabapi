package client

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/citizenadam/go-schwabapi/pkg/types"
)

// mockTokenGetter implements TokenGetter for testing
type mockTokenGetter struct {
	token string
}

func (m *mockTokenGetter) GetAccessToken() string {
	return m.token
}

func TestGetStreamerInfo(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   any
		responseStatus int
		wantErr        bool
		errContains    string
	}{
		{
			name: "successfully retrieves streamer info",
			responseBody: types.PreferencesResponse{
				StreamerInfo: &types.StreamerInfo{
					AccountID:      "test-account-id",
					AccountIDType:  "test-type",
					Token:          "test-token",
					TokenTimestamp: "2024-03-15T12:34:56Z",
					UserID:         "test-user-id",
					AppID:          "test-app-id",
					Secret:         "test-secret",
					AccessLevel:    "test-level",
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "handles missing streamer info",
			responseBody:   types.PreferencesResponse{},
			responseStatus: http.StatusOK,
			wantErr:        true,
			errContains:    "streamer info not found",
		},
		{
			name:           "handles HTTP error",
			responseBody:   types.PreferencesResponse{},
			responseStatus: http.StatusUnauthorized,
			wantErr:        true,
			errContains:    "streamer info not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/trader/v1/userPreference" {
					t.Errorf("expected path /trader/v1/userPreference, got %s", r.URL.Path)
				}

				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					t.Error("expected Authorization header")
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			mockToken := &mockTokenGetter{token: "test-access-token"}
			logger := slog.Default()
			client := NewClient(logger)
			oauthClient := NewOAuthClient(client, logger, "app-key", "app-secret", "callback-url", mockToken)
			oauthClient.baseURL = server.URL

			ctx := context.Background()
			result, err := oauthClient.GetStreamerInfo(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result, got nil")
				return
			}

			if result.AccountID != "test-account-id" {
				t.Errorf("expected AccountID test-account-id, got %s", result.AccountID)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
