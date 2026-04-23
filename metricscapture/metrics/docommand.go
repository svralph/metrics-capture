package metrics

import (
	"context"
	"fmt"
)

// HandleDoCommand handles metrics-specific commands and reports whether handled.
func (s *Service) HandleDoCommand(ctx context.Context, cmd map[string]any) (map[string]any, bool, error) {
	raw, ok := cmd["command"]
	if !ok {
		return nil, false, nil
	}
	cmdName, ok := raw.(string)
	if !ok {
		return nil, true, fmt.Errorf("command must be a string")
	}

	switch cmdName {
	case "metrics_flush_now":
		res, err := s.FlushNow(ctx)
		if err != nil {
			return map[string]any{
				"status":         "error",
				"uploaded_count": res.UploadedCount,
				"error":          err.Error(),
			}, true, err
		}
		return map[string]any{
			"status":         "ok",
			"uploaded_count": res.UploadedCount,
			"file_ids":       res.FileIDs,
		}, true, nil
	case "metrics_stats":
		st := s.Stats()
		return map[string]any{
			"status":            "ok",
			"queue_depth":       st.QueueDepth,
			"dropped_count":     st.DroppedCount,
			"uploaded_count":    st.UploadedCount,
			"upload_fail_count": st.UploadFailCount,
			"last_upload_at":    st.LastUploadAt,
			"last_upload_error": st.LastUploadError,
			"max_queue_size":    st.MaxQueueSize,
			"batch_size":        st.BatchSize,
		}, true, nil
	default:
		return nil, false, nil
	}
}
