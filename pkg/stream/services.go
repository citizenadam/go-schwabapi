package stream

import (
	"context"
	"encoding/json"

	"github.com/citizenadam/go-schwabapi/pkg/types"
)

// LevelOneEquities subscribes to Level One equity data
func (c *Client) LevelOneEquities(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "LEVELONE_EQUITIES",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// LevelOneOptions subscribes to Level One options data
func (c *Client) LevelOneOptions(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "LEVELONE_OPTIONS",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// LevelOneFutures subscribes to Level One futures data
func (c *Client) LevelOneFutures(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "LEVELONE_FUTURES",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// LevelOneFuturesOptions subscribes to Level One futures options data
func (c *Client) LevelOneFuturesOptions(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "LEVELONE_FUTURES_OPTIONS",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// LevelOneForex subscribes to Level One forex data
func (c *Client) LevelOneForex(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "LEVELONE_FOREX",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// NyseBook subscribes to NYSE book data
func (c *Client) NyseBook(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "NYSE_BOOK",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// NasdaqBook subscribes to NASDAQ book data
func (c *Client) NasdaqBook(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "NASDAQ_BOOK",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// OptionsBook subscribes to options book data
func (c *Client) OptionsBook(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "OPTIONS_BOOK",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// ChartEquity subscribes to Chart equity data
func (c *Client) ChartEquity(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "CHART_EQUITY",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}

// ChartFutures subscribes to Chart futures data
func (c *Client) ChartFutures(ctx context.Context, manager *Manager, keys string, fields string, command string) error {
	req := &types.Subscription{
		Service:   "CHART_FUTURES",
		Command:   command,
		RequestID: 0,
		Parameters: &types.SubscriptionParams{
			Keys:   keys,
			Fields: fields,
		},
	}

	// Record subscription for crash recovery
	if err := manager.RecordRequest(ctx, req); err != nil {
		return err
	}

	// Send subscription request
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Write(data)
}
