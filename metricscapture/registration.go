package metricscapture

import (
	"context"
	"fmt"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
)

var Model = resource.NewModel("bov", "metricscapture", "service")

func init() {
	resource.RegisterService(generic.API, Model, resource.Registration[resource.Resource, *Config]{
		Constructor: newService,
	})
}

type Config struct {
	PartID           string   `json:"part_id"`
	ComponentType    string   `json:"component_type,omitempty"`
	ComponentName    string   `json:"component_name,omitempty"`
	MethodName       string   `json:"method_name,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	MaxQueueSize     int      `json:"max_queue_size,omitempty"`
	BatchSize        int      `json:"batch_size,omitempty"`
	RetryMaxAttempts int      `json:"retry_max_attempts,omitempty"`
	RetryBaseDelayMs int      `json:"retry_base_delay_ms,omitempty"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.PartID == "" {
		return nil, nil, fmt.Errorf("%s.part_id is required", path)
	}
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
