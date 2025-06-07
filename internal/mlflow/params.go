package mlflow

import (
	"context"
	"fmt"

	"github.com/databricks/databricks-sdk-go/service/ml"
	"github.com/imishinist/mlflow-cli/internal/models"
)

func (c *Client) LogParam(ctx context.Context, runID string, key string, value string) error {
	err := c.client.Experiments.LogParam(ctx, ml.LogParam{
		RunId: runID,
		Key:   key,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to log parameter %s: %w", key, err)
	}

	return nil
}

func (c *Client) LogParams(ctx context.Context, runID string, params []models.Parameter) error {
	for _, param := range params {
		if err := c.LogParam(ctx, runID, param.Key, param.Value); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) LogParamsFromMap(ctx context.Context, runID string, params map[string]string) error {
	for key, value := range params {
		if err := c.LogParam(ctx, runID, key, value); err != nil {
			return err
		}
	}

	return nil
}
