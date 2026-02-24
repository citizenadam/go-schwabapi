package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/logger"
	"github.com/citizenadam/go-schwabapi/pkg/types"
	"log/slog"
)

// Router routes incoming stream messages to appropriate handlers
type Router struct {
	handler *Handler
	logger  *slog.Logger
}

// NewRouter creates a new message router
func NewRouter(handler *Handler, logger *slog.Logger) *Router {
	return &Router{
		handler: handler,
		logger:  logger,
	}
}

// RouteMessage routes an incoming message to the appropriate handler
// This method uses goroutines for concurrent handling and context with deadline
// to avoid blocking indefinitely
func (r *Router) RouteMessage(ctx context.Context, data []byte) error {
	// Create a context with deadline to avoid blocking indefinitely
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse the message
	msg, err := r.handler.ParseMessage(data)
	if err != nil {
		logger.ErrorCtx(ctx, r.logger, "Failed to parse message",
			"error", err,
			"data", string(data))
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Route based on service and command
	// Use goroutine for concurrent handling
	go func() {
		if err := r.routeByServiceAndCommand(ctx, msg); err != nil {
			logger.ErrorCtx(ctx, r.logger, "Failed to route message",
				"error", err,
				"service", msg.Service,
				"command", msg.Command)
		}
	}()

	return nil
}

// routeByServiceAndCommand routes a message based on its service and command
func (r *Router) routeByServiceAndCommand(ctx context.Context, msg *types.Message) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	// Log the incoming message
	logger.DebugCtx(ctx, r.logger, "Routing message",
		"service", msg.Service,
		"command", msg.Command,
		"requestId", msg.RequestID)

	// Route based on service
	switch msg.Service {
	case "ADMIN":
		return r.routeAdminCommand(ctx, msg)
	case "LEVELONE_EQUITIES":
		return r.routeLevelOneEquities(ctx, msg)
	case "LEVELONE_OPTIONS":
		return r.routeLevelOneOptions(ctx, msg)
	case "LEVELONE_FUTURES":
		return r.routeLevelOneFutures(ctx, msg)
	case "LEVELONE_FUTURES_OPTIONS":
		return r.routeLevelOneFuturesOptions(ctx, msg)
	case "LEVELONE_FOREX":
		return r.routeLevelOneForex(ctx, msg)
	case "NYSE_BOOK":
		return r.routeNyseBook(ctx, msg)
	case "NASDAQ_BOOK":
		return r.routeNasdaqBook(ctx, msg)
	case "OPTIONS_BOOK":
		return r.routeOptionsBook(ctx, msg)
	case "CHART_EQUITY":
		return r.routeChartEquity(ctx, msg)
	case "CHART_FUTURES":
		return r.routeChartFutures(ctx, msg)
	case "SCREENER_EQUITY":
		return r.routeScreenerEquity(ctx, msg)
	case "SCREENER_OPTION":
		return r.routeScreenerOption(ctx, msg)
	case "ACCT_ACTIVITY":
		return r.routeAccountActivity(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown service",
			"service", msg.Service,
			"command", msg.Command)
		return nil
	}
}

// routeAdminCommand routes ADMIN service commands
func (r *Router) routeAdminCommand(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "LOGIN", "LOGOUT", "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown ADMIN command",
			"command", msg.Command)
		return nil
	}
}

// routeLevelOneEquities routes LEVELONE_EQUITIES service commands
func (r *Router) routeLevelOneEquities(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown LEVELONE_EQUITIES command",
			"command", msg.Command)
		return nil
	}
}

// routeLevelOneOptions routes LEVELONE_OPTIONS service commands
func (r *Router) routeLevelOneOptions(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown LEVELONE_OPTIONS command",
			"command", msg.Command)
		return nil
	}
}

// routeLevelOneFutures routes LEVELONE_FUTURES service commands
func (r *Router) routeLevelOneFutures(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown LEVELONE_FUTURES command",
			"command", msg.Command)
		return nil
	}
}

// routeLevelOneFuturesOptions routes LEVELONE_FUTURES_OPTIONS service commands
func (r *Router) routeLevelOneFuturesOptions(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown LEVELONE_FUTURES_OPTIONS command",
			"command", msg.Command)
		return nil
	}
}

// routeLevelOneForex routes LEVELONE_FOREX service commands
func (r *Router) routeLevelOneForex(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown LEVELONE_FOREX command",
			"command", msg.Command)
		return nil
	}
}

// routeNyseBook routes NYSE_BOOK service commands
func (r *Router) routeNyseBook(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown NYSE_BOOK command",
			"command", msg.Command)
		return nil
	}
}

// routeNasdaqBook routes NASDAQ_BOOK service commands
func (r *Router) routeNasdaqBook(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown NASDAQ_BOOK command",
			"command", msg.Command)
		return nil
	}
}

// routeOptionsBook routes OPTIONS_BOOK service commands
func (r *Router) routeOptionsBook(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown OPTIONS_BOOK command",
			"command", msg.Command)
		return nil
	}
}

// routeChartEquity routes CHART_EQUITY service commands
func (r *Router) routeChartEquity(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown CHART_EQUITY command",
			"command", msg.Command)
		return nil
	}
}

// routeChartFutures routes CHART_FUTURES service commands
func (r *Router) routeChartFutures(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown CHART_FUTURES command",
			"command", msg.Command)
		return nil
	}
}

// routeScreenerEquity routes SCREENER_EQUITY service commands
func (r *Router) routeScreenerEquity(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown SCREENER_EQUITY command",
			"command", msg.Command)
		return nil
	}
}

// routeScreenerOption routes SCREENER_OPTION service commands
func (r *Router) routeScreenerOption(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "ADD", "SUBS", "UNSUBS", "VIEW":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown SCREENER_OPTION command",
			"command", msg.Command)
		return nil
	}
}

// routeAccountActivity routes ACCT_ACTIVITY service commands
func (r *Router) routeAccountActivity(ctx context.Context, msg *types.Message) error {
	switch msg.Command {
	case "SUBS", "UNSUBS":
		return r.handler.HandleMessage(ctx, msg)
	default:
		logger.WarnCtx(ctx, r.logger, "Unknown ACCT_ACTIVITY command",
			"command", msg.Command)
		return nil
	}
}

// RouteStreamRequest routes a stream request wrapper containing multiple subscriptions
func (r *Router) RouteStreamRequest(ctx context.Context, data []byte) error {
	// Create a context with deadline to avoid blocking indefinitely
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse the stream request
	req, err := r.handler.ParseStreamRequest(data)
	if err != nil {
		logger.ErrorCtx(ctx, r.logger, "Failed to parse stream request",
			"error", err,
			"data", string(data))
		return fmt.Errorf("failed to parse stream request: %w", err)
	}

	// Route each subscription in the request
	for _, sub := range req.Requests {
		// Convert subscription to message format
		msg := &types.Message{
			Service:   sub.Service,
			Command:   sub.Command,
			RequestID: sub.RequestID,
		}

		// Use goroutine for concurrent handling
		go func(m *types.Message) {
			if err := r.routeByServiceAndCommand(ctx, m); err != nil {
				logger.ErrorCtx(ctx, r.logger, "Failed to route subscription",
					"error", err,
					"service", m.Service,
					"command", m.Command)
			}
		}(msg)
	}

	return nil
}

// ValidateService validates if a service is supported
func (r *Router) ValidateService(service string) bool {
	validServices := map[string]bool{
		"ADMIN":                    true,
		"LEVELONE_EQUITIES":        true,
		"LEVELONE_OPTIONS":         true,
		"LEVELONE_FUTURES":         true,
		"LEVELONE_FUTURES_OPTIONS": true,
		"LEVELONE_FOREX":           true,
		"NYSE_BOOK":                true,
		"NASDAQ_BOOK":              true,
		"OPTIONS_BOOK":             true,
		"CHART_EQUITY":             true,
		"CHART_FUTURES":            true,
		"SCREENER_EQUITY":          true,
		"SCREENER_OPTION":          true,
		"ACCT_ACTIVITY":            true,
	}
	return validServices[service]
}

// GetHandler returns the underlying handler
func (r *Router) GetHandler() *Handler {
	return r.handler
}
