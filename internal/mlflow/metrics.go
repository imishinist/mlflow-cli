package mlflow

import (
	"context"
	"fmt"
	"time"

	"github.com/databricks/databricks-sdk-go/service/ml"

	"github.com/imishinist/mlflow-cli/internal/models"
)

func (c *Client) LogMetric(ctx context.Context, runID string, key string, value float64, timestamp *time.Time, step *int64) error {
	logMetric := ml.LogMetric{
		RunId: runID,
		Key:   key,
		Value: value,
	}

	if timestamp != nil {
		logMetric.Timestamp = timestamp.UnixMilli()
	} else {
		logMetric.Timestamp = time.Now().UnixMilli()
	}

	if step != nil {
		logMetric.Step = *step
	}

	err := c.client.Experiments.LogMetric(ctx, logMetric)
	if err != nil {
		return fmt.Errorf("failed to log metric %s: %w", key, err)
	}

	return nil
}

func (c *Client) LogMetrics(ctx context.Context, runID string, metrics []models.Metric) error {
	for _, metric := range metrics {
		if err := c.LogMetric(ctx, runID, metric.Key, metric.Value, &metric.Timestamp, &metric.Step); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) LogBatchMetrics(ctx context.Context, runID string, metrics []models.Metric) error {
	// For now, log metrics one by one to avoid batch API issues
	for _, metric := range metrics {
		if err := c.LogMetric(ctx, runID, metric.Key, metric.Value, &metric.Timestamp, &metric.Step); err != nil {
			return err
		}
	}
	return nil
}
