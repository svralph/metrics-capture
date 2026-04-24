package metricscapture

import (
	"context"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
)

var Model = resource.NewModel("sab-viam", "metricscapture", "service")

func init() {
	resource.RegisterService(generic.API, Model, resource.Registration[resource.Resource, *Config]{
		Constructor: newService,
	})
}

type Config struct {
	PartID           string   `json:"part_id,omitempty"`
	ComponentType    string   `json:"component_type,omitempty"`
	ComponentName    string   `json:"component_name,omitempty"`
	MethodName       string   `json:"method_name,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	MaxQueueSize     int      `json:"max_queue_size,omitempty"`
	BatchSize        int      `json:"batch_size,omitempty"`
	AutoFlushOnEmit  bool     `json:"auto_flush_on_emit,omitempty"`
	AutoFlushTimeoutMs int    `json:"auto_flush_timeout_ms,omitempty"`
	RetryMaxAttempts int      `json:"retry_max_attempts,omitempty"`
	RetryBaseDelayMs int      `json:"retry_base_delay_ms,omitempty"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	return nil, nil, nil
}

func newService(
	ctx context.Context,
	deps resource.Dependencies,
	rawConf resource.Config,
	logger logging.Logger,
) (resource.Resource, error) {
	_ = deps
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	startupCfg := StartupConfig{
		PartID:        conf.PartID,
		ComponentType: conf.ComponentType,
		ComponentName: conf.ComponentName,
		MethodName:    conf.MethodName,
		Tags:          conf.Tags,
		MaxQueueSize:  conf.MaxQueueSize,
		BatchSize:     conf.BatchSize,
		AutoFlushOnEmit: conf.AutoFlushOnEmit,
		AutoFlushTimeoutMs: conf.AutoFlushTimeoutMs,
	}
	if conf.RetryMaxAttempts > 0 {
		startupCfg.Retry.MaxAttempts = conf.RetryMaxAttempts
	}
	if conf.RetryBaseDelayMs > 0 {
		startupCfg.Retry.BaseDelay = time.Duration(conf.RetryBaseDelayMs) * time.Millisecond
	}

	svc, err := NewModuleServiceFromEnv(ctx, logger, startupCfg)
	if err != nil {
		return nil, err
	}
	svc.name = rawConf.ResourceName()
	return svc, nil
}
