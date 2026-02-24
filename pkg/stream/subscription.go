package stream

import (
	"context"
	"strings"
	"sync"

	"github.com/citizenadam/go-schwabapi/pkg/logger"
	"github.com/citizenadam/go-schwabapi/pkg/types"
	"log/slog"
)

// Manager manages stream subscriptions
type Manager struct {
	mu            sync.RWMutex
	subscriptions map[string]map[string][]string // service -> key -> fields
	logger        *slog.Logger
}

// NewManager creates a new subscription manager
func NewManager(logger *slog.Logger) *Manager {
	return &Manager{
		subscriptions: make(map[string]map[string][]string),
		logger:        logger,
	}
}

// RecordRequest records a subscription request for crash recovery
func (m *Manager) RecordRequest(ctx context.Context, req *types.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.Parameters == nil || req.Service == "" {
		return nil
	}

	keys := strToList(req.Parameters.Keys)
	fields := strToList(req.Parameters.Fields)

	// Add service to subscriptions if not already there
	if _, exists := m.subscriptions[req.Service]; !exists {
		m.subscriptions[req.Service] = make(map[string][]string)
	}

	switch req.Command {
	case "ADD":
		for _, key := range keys {
			if _, exists := m.subscriptions[req.Service][key]; !exists {
				m.subscriptions[req.Service][key] = fields
			} else {
				// Merge fields, removing duplicates
				merged := mergeFields(m.subscriptions[req.Service][key], fields)
				m.subscriptions[req.Service][key] = merged
			}
		}
		logger.DebugCtx(ctx, m.logger, "Recorded ADD subscription",
			"service", req.Service,
			"keys", keys,
			"fields", fields)

	case "SUBS":
		// Replace all subscriptions for this service
		m.subscriptions[req.Service] = make(map[string][]string)
		for _, key := range keys {
			m.subscriptions[req.Service][key] = fields
		}
		logger.DebugCtx(ctx, m.logger, "Recorded SUBS subscription",
			"service", req.Service,
			"keys", keys,
			"fields", fields)

	case "UNSUBS":
		for _, key := range keys {
			if _, exists := m.subscriptions[req.Service][key]; exists {
				delete(m.subscriptions[req.Service], key)
			}
		}
		logger.DebugCtx(ctx, m.logger, "Recorded UNSUBS subscription",
			"service", req.Service,
			"keys", keys)

	case "VIEW":
		// Update fields for all existing keys in this service
		for key := range m.subscriptions[req.Service] {
			m.subscriptions[req.Service][key] = fields
		}
		logger.DebugCtx(ctx, m.logger, "Recorded VIEW subscription",
			"service", req.Service,
			"fields", fields)

	default:
		logger.WarnCtx(ctx, m.logger, "Unknown subscription command",
			"command", req.Command)
	}

	return nil
}

// GetSubscriptions returns all recorded subscriptions
func (m *Manager) GetSubscriptions() map[string]map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a deep copy to prevent external modification
	result := make(map[string]map[string][]string)
	for service, keys := range m.subscriptions {
		result[service] = make(map[string][]string)
		for key, fields := range keys {
			result[service][key] = append([]string{}, fields...)
		}
	}
	return result
}

// Clear removes all subscriptions
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscriptions = make(map[string]map[string][]string)
}

// strToList converts a comma-separated string to a list of strings
func strToList(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// mergeFields merges two field lists, removing duplicates
func mergeFields(existing, new []string) []string {
	seen := make(map[string]bool)
	var result []string

	// Add existing fields
	for _, f := range existing {
		if !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}

	// Add new fields
	for _, f := range new {
		if !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}

	return result
}
