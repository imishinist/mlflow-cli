package mlflow

import (
	"context"
	"fmt"
	"time"

	"github.com/databricks/databricks-sdk-go/service/ml"

	"github.com/imishinist/mlflow-cli/internal/models"
)

func (c *Client) CreateRun(ctx context.Context, config *models.RunConfig) (*models.RunInfo, error) {
	var experimentID string

	// Get experiment ID
	if config.ExperimentID != nil {
		experimentID = *config.ExperimentID
	} else {
		return nil, fmt.Errorf("experiment ID must be provided")
	}

	// Generate run name if not provided
	runName := "run-" + time.Now().Format("2006-01-02-15-04-05")
	if config.RunName != nil {
		runName = *config.RunName
	}

	// Prepare tags
	tags := make([]ml.RunTag, 0)
	if config.Tags != nil {
		for key, value := range config.Tags {
			tags = append(tags, ml.RunTag{
				Key:   key,
				Value: value,
			})
		}
	}

	// Add run name as tag
	tags = append(tags, ml.RunTag{
		Key:   "mlflow.runName",
		Value: runName,
	})

	// Add description as tag if provided
	if config.Description != nil {
		tags = append(tags, ml.RunTag{
			Key:   "mlflow.note.content",
			Value: *config.Description,
		})
	}

	// Create run
	startTime := time.Now()
	resp, err := c.client.Experiments.CreateRun(ctx, ml.CreateRun{
		ExperimentId: experimentID,
		RunName:      runName,
		StartTime:    startTime.UnixMilli(),
		Tags:         tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	return &models.RunInfo{
		RunID:        resp.Run.Info.RunId,
		ExperimentID: experimentID,
		RunName:      runName,
		Status:       string(models.RunStatusRunning),
		StartTime:    startTime,
		Tags:         config.Tags,
		Description: func() string {
			if config.Description != nil {
				return *config.Description
			}
			return ""
		}(),
	}, nil
}

func (c *Client) UpdateRun(ctx context.Context, runID string, status models.RunStatus) error {
	// Convert status to MLflow status type
	var mlStatus ml.UpdateRunStatus
	switch status {
	case models.RunStatusRunning:
		mlStatus = ml.UpdateRunStatusRunning
	case models.RunStatusFinished:
		mlStatus = ml.UpdateRunStatusFinished
	case models.RunStatusFailed:
		mlStatus = ml.UpdateRunStatusFailed
	case models.RunStatusKilled:
		mlStatus = ml.UpdateRunStatusKilled
	default:
		mlStatus = ml.UpdateRunStatusFinished
	}

	updateRun := ml.UpdateRun{
		RunId:  runID,
		Status: mlStatus,
	}

	// Set end time for terminal statuses
	if status == models.RunStatusFinished || status == models.RunStatusFailed || status == models.RunStatusKilled {
		updateRun.EndTime = time.Now().UnixMilli()
	}

	_, err := c.client.Experiments.UpdateRun(ctx, updateRun)
	if err != nil {
		return fmt.Errorf("failed to update run: %w", err)
	}

	return nil
}

func (c *Client) GetRun(ctx context.Context, runID string) (*models.RunInfo, error) {
	resp, err := c.client.Experiments.GetRun(ctx, ml.GetRunRequest{
		RunId: runID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	run := resp.Run
	tags := make(map[string]string)
	for _, tag := range run.Data.Tags {
		tags[tag.Key] = tag.Value
	}

	runInfo := &models.RunInfo{
		RunID:        run.Info.RunId,
		ExperimentID: run.Info.ExperimentId,
		Status:       string(run.Info.Status),
		StartTime:    time.Unix(run.Info.StartTime/1000, 0),
		Tags:         tags,
	}

	if run.Info.EndTime != 0 {
		endTime := time.Unix(run.Info.EndTime/1000, 0)
		runInfo.EndTime = &endTime
	}

	if runName, exists := tags["mlflow.runName"]; exists {
		runInfo.RunName = runName
	}

	if description, exists := tags["mlflow.note.content"]; exists {
		runInfo.Description = description
	}

	return runInfo, nil
}
