package metricscapture

import (
	"context"
	"fmt"
)

// DoCommand routes custom commands. Metrics commands are delegated first.
func (s *ModuleService) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	if s.metrics != nil {
		resp, handled, err := s.metrics.HandleDoCommand(ctx, cmd)
		if handled {
			return resp, err
		}
	}

	return nil, fmt.Errorf("unknown command")
}
