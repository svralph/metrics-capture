package metricscapture

import (
	"context"
	"fmt"
	"os"
	"time"

	"metricscapture/internal/dataclient"
	"metricscapture/metricscapture/metrics"
	"go.viam.com/rdk/app"
	"go.viam.com/rdk/logging"
)

// StartupConfig holds required metadata and runtime options for upload wiring.
type StartupConfig struct {
	PartID        string
	ComponentType string
	ComponentName string
	MethodName    string
	Tags          []string

	MaxQueueSize int
	BatchSize    int
	AutoFlushOnEmit bool
	AutoFlushTimeoutMs int
	Retry        dataclient.RetryConfig
}

// NewModuleServiceFromEnv wires the real Viam DataClient adapter at startup.
// Required env vars: VIAM_API_KEY and VIAM_API_KEY_ID.
func NewModuleServiceFromEnv(ctx context.Context, logger logging.Logger, cfg StartupConfig) (*ModuleService, error) {
	apiKey := os.Getenv("VIAM_API_KEY")
	apiKeyID := os.Getenv("VIAM_API_KEY_ID")
	if apiKey == "" || apiKeyID == "" {
		return nil, fmt.Errorf("missing required env vars VIAM_API_KEY and/or VIAM_API_KEY_ID")
	}
	if cfg.PartID == "" {
		cfg.PartID = os.Getenv("VIAM_PART_ID")
	}
	if cfg.PartID == "" {
		return nil, fmt.Errorf("part id is required; set attributes.part_id or VIAM_PART_ID")
	}
	if cfg.ComponentType == "" {
		cfg.ComponentType = "rdk:component:sensor"
	}
	if cfg.ComponentName == "" {
		cfg.ComponentName = "metricscapture"
	}
	if cfg.MethodName == "" {
		cfg.MethodName = "Readings"
	}
	if cfg.Retry.MaxAttempts == 0 {
		cfg.Retry.MaxAttempts = 3
	}
	if cfg.Retry.BaseDelay == 0 {
		cfg.Retry.BaseDelay = 300 * time.Millisecond
	}

	adapter, cleanup, err := dataclient.NewAdapterFromAPIKey(
		ctx,
		logger,
		apiKey,
		apiKeyID,
		app.Options{},
		cfg.Retry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create viam data adapter: %w", err)
	}

	uploader := metrics.NewDataClientUploader(
		adapter,
		cfg.PartID,
		cfg.ComponentType,
		cfg.ComponentName,
		cfg.MethodName,
		cfg.Tags,
	)

	metricsSvc := metrics.NewService(metrics.Config{
		MaxQueueSize: cfg.MaxQueueSize,
		BatchSize:    cfg.BatchSize,
		AutoFlushOnEmit: cfg.AutoFlushOnEmit,
		AutoFlushTimeout: time.Duration(cfg.AutoFlushTimeoutMs) * time.Millisecond,
	}, uploader)

	return NewModuleService(metricsSvc, cleanup), nil
}
