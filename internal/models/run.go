package models

import "time"

type RunConfig struct {
	ExperimentID *string           `json:"experiment_id,omitempty"`
	RunName      *string           `json:"run_name,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Description  *string           `json:"description,omitempty"`
}

type RunInfo struct {
	RunID        string            `json:"run_id"`
	ExperimentID string            `json:"experiment_id"`
	RunName      string            `json:"run_name"`
	Status       string            `json:"status"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      *time.Time        `json:"end_time,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Description  string            `json:"description,omitempty"`
}

type RunStatus string

const (
	RunStatusRunning  RunStatus = "RUNNING"
	RunStatusFinished RunStatus = "FINISHED"
	RunStatusFailed   RunStatus = "FAILED"
	RunStatusKilled   RunStatus = "KILLED"
)
