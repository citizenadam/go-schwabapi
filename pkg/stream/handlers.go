package stream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/citizenadam/go-schwabapi/pkg/logger"
	"github.com/citizenadam/go-schwabapi/pkg/types"
	"log/slog"
)

// Handler handles WebSocket messages
type Handler struct {
	logger *slog.Logger
}

// NewHandler creates a new message handler
func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// ParseMessage parses a raw WebSocket message into a Message struct
func (h *Handler) ParseMessage(data []byte) (*types.Message, error) {
	var msg types.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	return &msg, nil
}

// HandleMessage handles an incoming message based on its command
func (h *Handler) HandleMessage(ctx context.Context, msg *types.Message) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	switch msg.Command {
	case "LOGIN":
		return h.handleLogin(ctx, msg)
	case "LOGOUT":
		return h.handleLogout(ctx, msg)
	case "ADD":
		return h.handleAdd(ctx, msg)
	case "SUBS":
		return h.handleSubs(ctx, msg)
	case "UNSUBS":
		return h.handleUnsubs(ctx, msg)
	case "VIEW":
		return h.handleView(ctx, msg)
	default:
		logger.WarnCtx(ctx, h.logger, "Unknown message command",
			"command", msg.Command,
			"service", msg.Service)
		return nil
	}
}

// handleLogin handles LOGIN messages
func (h *Handler) handleLogin(ctx context.Context, msg *types.Message) error {
	logger.InfoCtx(ctx, h.logger, "Handling LOGIN message",
		"service", msg.Service,
		"requestId", msg.RequestID)

	// LOGIN is typically a response from the server
	// The client sends LOGIN, server responds with login status
	// This handler processes the server's response

	if msg.Content != nil {
		// Check for login success/failure in content
		if status, ok := msg.Content["status"].(string); ok {
			if status == "success" || status == "ok" {
				logger.InfoCtx(ctx, h.logger, "Login successful")
			} else {
				logger.WarnCtx(ctx, h.logger, "Login failed",
					"status", status)
			}
		}
	}

	return nil
}

// handleLogout handles LOGOUT messages
func (h *Handler) handleLogout(ctx context.Context, msg *types.Message) error {
	logger.InfoCtx(ctx, h.logger, "Handling LOGOUT message",
		"service", msg.Service,
		"requestId", msg.RequestID)

	// LOGOUT is typically a response from the server
	// The client sends LOGOUT, server responds with logout confirmation

	if msg.Content != nil {
		if status, ok := msg.Content["status"].(string); ok {
			logger.InfoCtx(ctx, h.logger, "Logout response received",
				"status", status)
		}
	}

	return nil
}

// handleAdd handles ADD subscription messages
func (h *Handler) handleAdd(ctx context.Context, msg *types.Message) error {
	logger.InfoCtx(ctx, h.logger, "Handling ADD subscription message",
		"service", msg.Service,
		"requestId", msg.RequestID)

	// ADD is used to add subscriptions to existing ones
	// Server responds with subscription confirmation

	if msg.Content != nil {
		if response, ok := msg.Content["response"].(map[string]interface{}); ok {
			if code, ok := response["code"].(float64); ok {
				if code == 0 {
					logger.InfoCtx(ctx, h.logger, "ADD subscription successful")
				} else {
					logger.WarnCtx(ctx, h.logger, "ADD subscription failed",
						"code", code)
				}
			}
		}
	}

	return nil
}

// handleSubs handles SUBS subscription messages
func (h *Handler) handleSubs(ctx context.Context, msg *types.Message) error {
	logger.InfoCtx(ctx, h.logger, "Handling SUBS subscription message",
		"service", msg.Service,
		"requestId", msg.RequestID)

	// SUBS is used to replace all subscriptions for a service
	// Server responds with subscription confirmation

	if msg.Content != nil {
		if response, ok := msg.Content["response"].(map[string]interface{}); ok {
			if code, ok := response["code"].(float64); ok {
				if code == 0 {
					logger.InfoCtx(ctx, h.logger, "SUBS subscription successful")
				} else {
					logger.WarnCtx(ctx, h.logger, "SUBS subscription failed",
						"code", code)
				}
			}
		}
	}

	return nil
}

// handleUnsubs handles UNSUBS subscription messages
func (h *Handler) handleUnsubs(ctx context.Context, msg *types.Message) error {
	logger.InfoCtx(ctx, h.logger, "Handling UNSUBS subscription message",
		"service", msg.Service,
		"requestId", msg.RequestID)

	// UNSUBS is used to remove subscriptions
	// Server responds with unsubscription confirmation

	if msg.Content != nil {
		if response, ok := msg.Content["response"].(map[string]interface{}); ok {
			if code, ok := response["code"].(float64); ok {
				if code == 0 {
					logger.InfoCtx(ctx, h.logger, "UNSUBS subscription successful")
				} else {
					logger.WarnCtx(ctx, h.logger, "UNSUBS subscription failed",
						"code", code)
				}
			}
		}
	}

	return nil
}

// handleView handles VIEW subscription messages
func (h *Handler) handleView(ctx context.Context, msg *types.Message) error {
	logger.InfoCtx(ctx, h.logger, "Handling VIEW subscription message",
		"service", msg.Service,
		"requestId", msg.RequestID)

	// VIEW is used to update fields for all existing subscriptions
	// Server responds with update confirmation

	if msg.Content != nil {
		if response, ok := msg.Content["response"].(map[string]interface{}); ok {
			if code, ok := response["code"].(float64); ok {
				if code == 0 {
					logger.InfoCtx(ctx, h.logger, "VIEW subscription successful")
				} else {
					logger.WarnCtx(ctx, h.logger, "VIEW subscription failed",
						"code", code)
				}
			}
		}
	}

	return nil
}

// ParseSubscription parses a subscription request from raw data
func (h *Handler) ParseSubscription(data []byte) (*types.Subscription, error) {
	var sub types.Subscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}
	return &sub, nil
}

// ParseStreamRequest parses a stream request wrapper from raw data
func (h *Handler) ParseStreamRequest(data []byte) (*types.StreamRequest, error) {
	var req types.StreamRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse stream request: %w", err)
	}
	return &req, nil
}

// ValidateCommand validates if a command is supported
func (h *Handler) ValidateCommand(command string) bool {
	validCommands := map[string]bool{
		"LOGIN":  true,
		"LOGOUT": true,
		"ADD":    true,
		"SUBS":   true,
		"UNSUBS": true,
		"VIEW":   true,
	}
	return validCommands[command]
}
