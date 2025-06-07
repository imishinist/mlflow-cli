package models

import "time"

type MetricPoint struct {
	Timestamp     *time.Time `json:"timestamp,omitempty"`
	Step          *int64     `json:"step,omitempty"`
	ExecutionTime float64    `json:"execution_time,omitempty"`
	SuccessRate   float64    `json:"success_rate,omitempty"`
	ErrorCount    float64    `json:"error_count,omitempty"`
	// Add more fields as needed
}

type MetricsFile struct {
	Metrics []MetricPoint `json:"metrics"`
}

type Metric struct {
	Key       string    `json:"key"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Step      int64     `json:"step"`
}

type TimeConfig struct {
	Resolution string // 1m, 5m, 1h
	Alignment  string // floor, ceil, round
	StepMode   string // auto, timestamp, sequence
}
