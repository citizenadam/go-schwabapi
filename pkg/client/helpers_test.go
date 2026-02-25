package client

import (
	"testing"
	"time"
)

// TestFormatISO8601Date verifies ISO8601 date formatting
func TestFormatISO8601Date(t *testing.T) {
	// Fixed test time: 2024-03-15 12:34:56 UTC
	testTime := time.Date(2024, 3, 15, 12, 34, 56, 0, time.UTC)

	result := FormatISO8601Date(testTime)
	expected := "2024-03-15T12:34:56Z"

	if result != expected {
		t.Errorf("FormatISO8601Date() = %v, want %v", result, expected)
	}
}

// TestFormatEPOCH verifies Unix timestamp in seconds
func TestFormatEPOCH(t *testing.T) {
	// Fixed test time: 2024-03-15 12:34:56 UTC
	// Unix timestamp: 1710506096
	testTime := time.Date(2024, 3, 15, 12, 34, 56, 0, time.UTC)

	result := FormatEPOCH(testTime)
	expected := "1710506096"

	if result != expected {
		t.Errorf("FormatEPOCH() = %v, want %v", result, expected)
	}
}

// TestFormatEPOCH_MS verifies Unix timestamp in milliseconds
func TestFormatEPOCH_MS(t *testing.T) {
	// Fixed test time: 2024-03-15 12:34:56 UTC
	// Unix timestamp in milliseconds: 1710506096000
	testTime := time.Date(2024, 3, 15, 12, 34, 56, 0, time.UTC)

	result := FormatEPOCH_MS(testTime)
	expected := "1710506096000"

	if result != expected {
		t.Errorf("FormatEPOCH_MS() = %v, want %v", result, expected)
	}
}

// TestFormatYYYYMMDD verifies YYYY-MM-DD format
func TestFormatYYYYMMDD(t *testing.T) {
	// Fixed test time: 2024-03-15 12:34:56 UTC
	testTime := time.Date(2024, 3, 15, 12, 34, 56, 0, time.UTC)

	result := FormatYYYYMMDD(testTime)
	expected := "2024-03-15"

	if result != expected {
		t.Errorf("FormatYYYYMMDD() = %v, want %v", result, expected)
	}
}

// TestFormatISO8601DateWithNow verifies current time formatting
func TestFormatISO8601DateWithNow(t *testing.T) {
	now := time.Now().UTC()
	result := FormatISO8601Date(now)

	// Verify it's a valid RFC3339 format by parsing it back
	_, err := time.Parse(time.RFC3339, result)
	if err != nil {
		t.Errorf("FormatISO8601Date() produced invalid RFC3339 format: %v", err)
	}
}

// TestFormatEPOCHWithNow verifies current time epoch formatting
func TestFormatEPOCHWithNow(t *testing.T) {
	now := time.Now()
	result := FormatEPOCH(now)

	// Verify it's a valid number by checking length
	if len(result) < 10 {
		t.Errorf("FormatEPOCH() produced too short result: %v", result)
	}
}

// TestFormatEPOCH_MSWithNow verifies current time epoch ms formatting
func TestFormatEPOCH_MSWithNow(t *testing.T) {
	now := time.Now()
	result := FormatEPOCH_MS(now)

	// Verify it's a valid number by checking length (should be 13 digits for ms)
	if len(result) < 13 {
		t.Errorf("FormatEPOCH_MS() produced too short result: %v", result)
	}
}

// TestFormatYYYYMMDDWithNow verifies current time date formatting
func TestFormatYYYYMMDDWithNow(t *testing.T) {
	now := time.Now().UTC()
	result := FormatYYYYMMDD(now)

	// Verify it's a valid date format by parsing it back
	_, err := time.Parse("2006-01-02", result)
	if err != nil {
		t.Errorf("FormatYYYYMMDD() produced invalid date format: %v", err)
	}
}
