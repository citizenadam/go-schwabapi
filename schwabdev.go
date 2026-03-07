// Package schwabdev provides a Go client for the Charles Schwab API.
// It offers both synchronous and asynchronous operations for API access,
// real-time data streaming, and order management.
//
// This package is not affiliated with or endorsed by Schwab.
// Licensed under the MIT license and acts in accordance with Schwab's API terms and conditions.
//
// Basic usage:
//
//	client := schwabdev.NewClient(appKey, appSecret, callbackURL)
//	stream := schwabdev.NewStream(client)
package schwabdev
