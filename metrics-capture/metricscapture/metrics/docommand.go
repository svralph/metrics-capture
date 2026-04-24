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
	case "record_payload_now":
		name, ok := cmd["name"].(string)
		if !ok || name == "" {
			return nil, true, fmt.Errorf("record_payload_now requires non-empty string field 'name'")
		}
		value := 1.0
		if rawValue, hasValue := cmd["value"]; hasValue {
			parsedValue, ok := rawValue.(float64)
			if !ok {
				return nil, true, fmt.Errorf("record_payload_now field 'value' must be a number")
			}
			value = parsedValue
		}
		tags := map[string]string{}
		if rawTags, hasTags := cmd["tags"]; hasTags {
			tagMap, ok := rawTags.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("record_payload_now field 'tags' must be an object")
			}
			for k, v := range tagMap {
				str, ok := v.(string)
				if !ok {
					return nil, true, fmt.Errorf("record_payload_now tag %q must be a string", k)
				}
				tags[k] = str
			}
		}
		payload := map[string]any{}
		if rawPayload, hasPayload := cmd["payload"]; hasPayload {
			payloadMap, ok := rawPayload.(map[string]any)
			if !ok {
				return nil, true, fmt.Errorf("record_payload_now field 'payload' must be an object")
			}
			payload = payloadMap
		}
		s.EmitPayload(name, value, tags, payload)
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
