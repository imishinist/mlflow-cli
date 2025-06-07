package timeutils

import (
	"fmt"
	"time"

	"github.com/imishinist/mlflow-cli/internal/models"
)

// AlignTimestamp aligns timestamp to the specified resolution and alignment
func AlignTimestamp(t time.Time, resolution string, alignment string) (time.Time, error) {
	var duration time.Duration

	switch resolution {
	case "1m":
		duration = time.Minute
	case "5m":
		duration = 5 * time.Minute
	case "1h":
		duration = time.Hour
	default:
		return t, fmt.Errorf("unsupported resolution: %s", resolution)
	}

	// Truncate to the resolution
	aligned := t.Truncate(duration)

	switch alignment {
	case "floor":
		return aligned, nil
	case "ceil":
		if t.After(aligned) {
			return aligned.Add(duration), nil
		}
		return aligned, nil
	case "round":
		half := duration / 2
		if t.Sub(aligned) >= half {
			return aligned.Add(duration), nil
		}
		return aligned, nil
	default:
		return t, fmt.Errorf("unsupported alignment: %s", alignment)
	}
}

// ProcessMetrics processes metrics according to time configuration
func ProcessMetrics(metrics []models.MetricPoint, config models.TimeConfig, baseTime *time.Time) ([]models.Metric, error) {
	var result []models.Metric
	var base time.Time

	if baseTime != nil {
		base = *baseTime
	} else if len(metrics) > 0 && metrics[0].Timestamp != nil {
		base = *metrics[0].Timestamp
	} else {
		base = time.Now()
	}

	for _, point := range metrics {
		var timestamp time.Time
		var step int64

		// Determine timestamp
		if point.Timestamp != nil {
			var err error
			timestamp, err = AlignTimestamp(*point.Timestamp, config.Resolution, config.Alignment)
			if err != nil {
				return nil, err
			}
		} else {
			timestamp = time.Now()
		}

		// Determine step
		if point.Step != nil {
			step = *point.Step
		} else {
			switch config.StepMode {
			case "timestamp":
				// Convert timestamp to minutes from base time
				step = int64(timestamp.Sub(base).Minutes())
			case "sequence":
				step = int64(len(result))
			case "auto":
				if point.Timestamp != nil {
					step = int64(timestamp.Sub(base).Minutes())
				} else {
					step = int64(len(result))
				}
			}
		}

		// Convert each field to a separate metric
		if point.ExecutionTime != 0 {
			result = append(result, models.Metric{
				Key:       "execution_time",
				Value:     point.ExecutionTime,
				Timestamp: timestamp,
				Step:      step,
			})
		}

		if point.SuccessRate != 0 {
			result = append(result, models.Metric{
				Key:       "success_rate",
				Value:     point.SuccessRate,
				Timestamp: timestamp,
				Step:      step,
			})
		}

		// ErrorCount can be 0, so we always include it
		result = append(result, models.Metric{
			Key:       "error_count",
			Value:     point.ErrorCount,
			Timestamp: timestamp,
			Step:      step,
		})
	}

	return result, nil
}
