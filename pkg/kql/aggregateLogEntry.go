package kql

import "time"

// AggregateLogEntry represents a single log entry to be saved in Log Analytics.
// It contains the time the log was generated and the current time, due to logs being able to be
// sent to log analytics only within 2 days to the past. The OriginalTimeGenerated field allows the user to overcome this limitation.
type AggregateLogEntry struct {
	TimeGenerated         time.Time  `json:"TimeGenerated"`
	OriginalTimeGenerated *time.Time `json:"OriginalTimeGenerated"`
	Name                  string     `json:"Name"`
	Value                 float64    `json:"Value"`
}