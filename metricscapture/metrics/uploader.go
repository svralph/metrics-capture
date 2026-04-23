package metrics

import (
	"context"
	"time"
)

// Uploader sends a batch of metrics to Viam's data APIs.
type Uploader interface {
	Upload(ctx context.Context, batch []Metric) ([]string, error)
}

// RequestTime captures when data was requested and then received.
type RequestTime struct {
	RequestedAt time.Time
	ReceivedAt  time.Time
}

// TabularUploadRequest mirrors what Viam's TabularDataCaptureUpload needs.
type TabularUploadRequest struct {
	TabularData       []map[string]any
	PartID            string
	ComponentType     string
	ComponentName     string
	MethodName        string
	DataRequestTimes  []RequestTime
	Tags              []string
}

// TabularDataClient defines the minimal method needed from the Viam data client.
type TabularDataClient interface {
	TabularDataCaptureUpload(ctx context.Context, req TabularUploadRequest) (string, error)
}

// DataClientUploader converts internal metrics into tabular upload requests.
type DataClientUploader struct {
	client        TabularDataClient
	partID        string
	componentType string
	componentName string
	methodName    string
	tags          []string
}

func NewDataClientUploader(
	client TabularDataClient,
	partID, componentType, componentName, methodName string,
	tags []string,
) *DataClientUploader {
	return &DataClientUploader{
		client:        client,
		partID:        partID,
		componentType: componentType,
		componentName: componentName,
		methodName:    methodName,
		tags:          tags,
	}
}

func (u *DataClientUploader) Upload(ctx context.Context, batch []Metric) ([]string, error) {
	if len(batch) == 0 {
		return nil, nil
	}

	tabular := make([]map[string]any, 0, len(batch))
	reqTimes := make([]RequestTime, 0, len(batch))
	now := time.Now().UTC()

	for _, m := range batch {
		ts := m.Timestamp
		if ts.IsZero() {
			ts = now
		}
		row := map[string]any{
			"readings": map[string]any{
				"metric_name": m.Name,
				"value":       m.Value,
				"timestamp":   ts.Format(time.RFC3339Nano),
				"tags":        m.Tags,
			},
		}
		tabular = append(tabular, row)
		reqTimes = append(reqTimes, RequestTime{
			RequestedAt: ts,
			ReceivedAt:  now,
		})
	}

	fileID, err := u.client.TabularDataCaptureUpload(ctx, TabularUploadRequest{
		TabularData:      tabular,
		PartID:           u.partID,
		ComponentType:    u.componentType,
		ComponentName:    u.componentName,
		MethodName:       u.methodName,
		DataRequestTimes: reqTimes,
		Tags:             u.tags,
	})
	if err != nil {
		return nil, err
	}
	return []string{fileID}, nil
}
