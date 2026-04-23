package dataclient

import (
	"context"
	"fmt"

	"bov-2/metricscapture/metrics"
	"go.viam.com/rdk/app"
	"go.viam.com/rdk/logging"
)

// Adapter wraps Viam's app.DataClient and implements metrics.TabularDataClient.
type Adapter struct {
	dataClient *app.DataClient
	retry      RetryConfig
}

// Ensure interface compatibility at compile time.
var _ metrics.TabularDataClient = (*Adapter)(nil)

func NewAdapter(dataClient *app.DataClient, retry RetryConfig) *Adapter {
	return &Adapter{
		dataClient: dataClient,
		retry:      retry.normalized(),
	}
}

// NewAdapterFromViamClient builds an adapter from an existing Viam app client.
func NewAdapterFromViamClient(client *app.ViamClient, retry RetryConfig) (*Adapter, error) {
	if client == nil {
		return nil, fmt.Errorf("viam client is nil")
	}
	return NewAdapter(client.DataClient(), retry), nil
}

// NewAdapterFromAPIKey builds a fresh Viam client and data adapter.
// Caller should close the returned cleanup func when done.
func NewAdapterFromAPIKey(
	ctx context.Context,
	logger logging.Logger,
	apiKey string,
	apiKeyID string,
	options app.Options,
	retry RetryConfig,
) (*Adapter, func() error, error) {
	if logger == nil {
		return nil, nil, fmt.Errorf("logger is nil")
	}
	vc, err := app.CreateViamClientWithAPIKey(ctx, options, apiKey, apiKeyID, logger)
	if err != nil {
		return nil, nil, err
	}

	adapter := NewAdapter(vc.DataClient(), retry)
	return adapter, vc.Close, nil
}

func (a *Adapter) TabularDataCaptureUpload(ctx context.Context, req metrics.TabularUploadRequest) (string, error) {
	if a == nil || a.dataClient == nil {
		return "", fmt.Errorf("data client adapter is not configured")
	}
	tabularData, partID, componentType, componentName, methodName, dataRequestTimes, tags :=
		toAppTabularUploadArgs(req)

	if len(tabularData) == 0 {
		return "", fmt.Errorf("tabular data cannot be empty")
	}
	if len(tabularData) != len(dataRequestTimes) {
		return "", fmt.Errorf("tabular_data and data_request_times length mismatch: %d vs %d", len(tabularData), len(dataRequestTimes))
	}

	options := &app.TabularDataCaptureUploadOptions{
		Tags: tags,
	}
	return retryUpload(ctx, a.retry, func(callCtx context.Context) (string, error) {
		return a.dataClient.TabularDataCaptureUpload(
			callCtx,
			tabularData,
			partID,
			componentType,
			componentName,
			methodName,
			dataRequestTimes,
			options,
		)
	})
}
