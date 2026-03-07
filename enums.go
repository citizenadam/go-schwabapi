package schwabdev

// TimeFormat represents the different time format options for API responses.
// These formats match the Python implementation for cross-language compatibility.
type TimeFormat string

const (
	TimeFormatISO8601  TimeFormat = "8601"
	TimeFormatEPOCH    TimeFormat = "epoch"
	TimeFormatEPOCHMS  TimeFormat = "epoch_ms"
	TimeFormatYYYYMMDD TimeFormat = "YYYY-MM-DD"
)

func (tf TimeFormat) String() string {
	return string(tf)
}
