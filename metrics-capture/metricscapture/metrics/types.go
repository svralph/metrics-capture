package metrics

import "time"

// Metric is the normalized event payload captured by the service.
type Metric struct {
	Name      string
	Value     float64
	Timestamp time.Time
	Tags      map[string]string
	Payload   map[string]any
}

// Stats is a snapshot of service health and upload state.
type Stats struct {
	QueueDepth       int       `json:"queue_depth"`
	DroppedCount     uint64    `json:"dropped_count"`
	UploadedCount    uint64    `json:"uploaded_count"`
	UploadFailCount  uint64    `json:"upload_fail_count"`
	LastUploadAt     time.Time `json:"last_upload_at,omitempty"`
	LastUploadError  string    `json:"last_upload_error,omitempty"`
	MaxQueueSize     int       `json:"max_queue_size"`
	BatchSize        int       `json:"batch_size"`
}

// FlushResult summarizes one on-demand upload attempt.
type FlushResult struct {
	UploadedCount int      `json:"uploaded_count"`
	FileIDs       []string `json:"file_ids,omitempty"`
}
