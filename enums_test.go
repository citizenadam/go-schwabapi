package schwabdev

import (
	"testing"
)

func TestTimeFormatConstants(t *testing.T) {
	cases := []struct {
		name     string
		constant TimeFormat
		expected string
	}{
		{"TimeFormatISO8601", TimeFormatISO8601, "8601"},
		{"TimeFormatEPOCH", TimeFormatEPOCH, "epoch"},
		{"TimeFormatEPOCHMS", TimeFormatEPOCHMS, "epoch_ms"},
		{"TimeFormatYYYYMMDD", TimeFormatYYYYMMDD, "YYYY-MM-DD"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if string(tc.constant) != tc.expected {
				t.Errorf("%s = %q, want %q", tc.name, string(tc.constant), tc.expected)
			}
		})
	}
}

func TestTimeFormatString(t *testing.T) {
	cases := []struct {
		name     string
		value    TimeFormat
		expected string
	}{
		{"ISO8601", TimeFormatISO8601, "8601"},
		{"EPOCH", TimeFormatEPOCH, "epoch"},
		{"EPOCHMS", TimeFormatEPOCHMS, "epoch_ms"},
		{"YYYYMMDD", TimeFormatYYYYMMDD, "YYYY-MM-DD"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.value.String()
			if got != tc.expected {
				t.Errorf("TimeFormat(%q).String() = %q, want %q", tc.value, got, tc.expected)
			}
		})
	}
}

func TestEnums(t *testing.T) {
	t.Run("TimeFormatISO8601", func(t *testing.T) {
		if TimeFormatISO8601 != "8601" {
			t.Errorf("TimeFormatISO8601 = %q, want \"8601\"", TimeFormatISO8601)
		}
	})
	t.Run("TimeFormatEPOCH", func(t *testing.T) {
		if TimeFormatEPOCH != "epoch" {
			t.Errorf("TimeFormatEPOCH = %q, want \"epoch\"", TimeFormatEPOCH)
		}
	})
	t.Run("TimeFormatEPOCHMS", func(t *testing.T) {
		if TimeFormatEPOCHMS != "epoch_ms" {
			t.Errorf("TimeFormatEPOCHMS = %q, want \"epoch_ms\"", TimeFormatEPOCHMS)
		}
	})
	t.Run("TimeFormatYYYYMMDD", func(t *testing.T) {
		if TimeFormatYYYYMMDD != "YYYY-MM-DD" {
			t.Errorf("TimeFormatYYYYMMDD = %q, want \"YYYY-MM-DD\"", TimeFormatYYYYMMDD)
		}
	})
}
