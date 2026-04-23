package metricscapture

import (
	"context"

	"bov-2/metricscapture/metrics"
	"go.viam.com/rdk/resource"
)

// ModuleService is the top-level resource/service wrapper for this module.
type ModuleService struct {
	resource.AlwaysRebuild

	name    resource.Name
	metrics *metrics.Service
	cleanup func() error
}

func NewModuleService(metricsSvc *metrics.Service, cleanup func() error) *ModuleService {
	return &ModuleService{
		metrics: metricsSvc,
		cleanup: cleanup,
	}
}

func (s *ModuleService) Name() resource.Name {
	return s.name
}

// EmitMetric is a helper that your business logic can call.
func (s *ModuleService) EmitMetric(name string, value float64, tags map[string]string) {
	if s.metrics == nil {
		return
	}
	s.metrics.Emit(name, value, tags)
}

func (s *ModuleService) Close(context.Context) error {
	if s.cleanup != nil {
		return s.cleanup()
	}
	return nil
}

func (s *ModuleService) Status(context.Context) (map[string]interface{}, error) {
	if s.metrics == nil {
		return map[string]interface{}{
			"status": "not_configured",
		}, nil
	}
	st := s.metrics.Stats()
	return map[string]interface{}{
		"queue_depth":       st.QueueDepth,
		"dropped_count":     st.DroppedCount,
		"uploaded_count":    st.UploadedCount,
		"upload_fail_count": st.UploadFailCount,
		"last_upload_at":    st.LastUploadAt,
		"last_upload_error": st.LastUploadError,
		"max_queue_size":    st.MaxQueueSize,
		"batch_size":        st.BatchSize,
	}, nil
}
